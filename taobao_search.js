var Nightmare = require('nightmare');
var nightmare = Nightmare({ show: false })

var args = process.argv.slice(2);
if (args.length < 2) {
  console.log("usage: node taobao_search.js 'search term' page");
  process.exit(1);
}

// search is the product search term
var search = args[0];

// page is the page offset we want (0-based)
var page = parseInt(args[1]);

// render the page and scape out all the product details urls and print to STDOUT
nightmare
  .viewport(1000, 6000)
  .goto("https://world.taobao.com/search/search.htm?s=" + (page*60) + "&q=" + encodeURIComponent(search))
  .wait()
  .evaluate(function () {
    var resp = {};
    resp["product_urls"] = [];

    var urls = document.querySelectorAll(".m-items .pic a");
    [].forEach.call(urls, function(url) {
      resp["product_urls"].push(url.href);
    });

    return JSON.stringify(resp, null, 4);
  })
  .end()
  .then(function (result) {
    console.log(result)
  })
