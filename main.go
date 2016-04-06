package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type SearchResult struct {
	ProductURLS []string `json:"product_urls"`
}

type ProductInfo struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Price  string   `json:"price"`
	Images []string `json:"images"`
}

func scrapeSearchPage(searchTerm string, page int) (SearchResult, error) {
	log.Printf("Searching %s at page %d", searchTerm, page)

	args := []string{
		"taobao_search.js",
		searchTerm,
		strconv.Itoa(page),
	}

	cmd := exec.Command("node", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Start()
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var output []byte
	select {
	case <-time.After(1 * time.Minute):
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
		log.Println("process killed as timeout reached")
		return nil
	case err := <-done:
		if err != nil {
			log.Printf("process done with error = %v", err)
			return nil
		}
		output = out.Bytes()
	}

	// unmarshal the response
	var sr SearchResult
	err = json.Unmarshal(output, &sr)
	if err != nil {
		return SearchResult{}, err
	}

	return sr, nil
}

func scrapeDetailsPage(searchTerm, detailsPageURL string) error {
	log.Printf("Downloading product info %s", detailsPageURL)

	args := []string{
		"taobao_details.js",
		detailsPageURL,
	}

	cmd := exec.Command("node", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Start()
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var output []byte
	select {
	case <-time.After(1 * time.Minute):
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
		log.Println("process killed as timeout reached")
		return nil
	case err := <-done:
		if err != nil {
			log.Printf("process done with error = %v", err)
			return nil
		}
		output = out.Bytes()
	}

	log.Printf(string(output))

	// unmarshal the response
	var pi ProductInfo
	err = json.Unmarshal(output, &pi)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join("downloads", searchTerm, pi.Name), 0777)
	if err != nil {
		return err
	}

	for _, imageURL := range pi.Images {
		filename := path.Base(imageURL)
		dst := filepath.Join("downloads", searchTerm, pi.Name, filename)
		err = download(imageURL, dst)
		if err != nil {
			return err
		}
	}

	dst := filepath.Join("downloads", searchTerm, pi.Name, "info.json")
	ioutil.WriteFile(dst, output, 0777)

	return nil
}

func download(url, dst string) error {
	log.Printf("Downloading %s to %s", url, dst)

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	ioutil.WriteFile(dst, data, 0777)
	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Usage: crawl dict.txt")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	page := 0
	for {
		for scanner.Scan() {
			searchTerm := scanner.Text()
			sr, err := scrapeSearchPage(searchTerm, page)
			if err != nil {
				log.Printf("Error %s when running scrapeSearchPage", err)
				continue
			}

			for _, url := range sr.ProductURLS {
				err = scrapeDetailsPage(searchTerm, url)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		page += 1
	}

	log.Print("Done")
}
