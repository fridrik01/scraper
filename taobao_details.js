var Nightmare = require('nightmare');
var nightmare = Nightmare({ show: false})

var args = process.argv.slice(2);
if (args.length < 1) {
  console.log("usage: node taobao_details.js url");
  process.exit(1);
}

var product_url = args[0];

// render the page and scape out all the product details urls and print to STDOUT
nightmare
  //.viewport(1000, 6000)
  .goto(product_url)
  .wait()
  .evaluate(function (url) {
    var resp = {};
    resp["name"] = document.querySelector(".t-title");
    resp["url"] = url;

    if (resp["name"] != null) {
        resp["name"] = resp["name"].textContent;
    } else {
      resp["name"] = document.querySelector(".tb-detail-hd h1");
      if (resp["name"] != null) {
        resp["name"] = resp["name"].textContent;
      }
    }

    resp["price"] = document.querySelector(".price-show .tb-rmb-num span")
    if (resp["price"] != null) {
      resp["price"] = resp["price"].textContent;
    } else {
      resp["price"] = document.querySelector(".tm-price")
      if (resp["price"] != null) {
        resp["price"] = resp["price"].textContent;
      }
    }

    var re = new RegExp(/\.(jpg|jpeg|png)$/i);

    resp["images"] = [];
    var images = document.querySelectorAll(".item-gallery img");
    [].forEach.call(images, function(image) {
      if (re.test(image.src)) {
        resp["images"].push(image.src);
      }
    });

    if (resp["images"].length < 1) {
      var images = document.querySelectorAll("img");
      [].forEach.call(images, function(image) {
        if (re.test(image.src)) {
          resp["images"].push(image.src);
        }
      });
    }

    return JSON.stringify(resp, null, 4);
  }, product_url)
  .end()
  .then(function (result) {
    console.log(result)
  })
