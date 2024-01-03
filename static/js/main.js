window.onload = function () {
  initNavigation()

  fetch('/api/v1/users/me')
    .then(response => {
      if (response.ok) {
        return response.json()
      }

      throw new Error(`HTTP error! status: ${response.status}`);
    })
    .then(currentUser => {
      var categories = Array.from(document.querySelectorAll('#categoryModal input')).reduce((map, input, index) => {
        map[input.value] = Math.pow(2, index);
        return map;
      }, {});

      document.getElementById('search').addEventListener('input', function (e) {
        searchProducts(e.target.value);
      });

      function searchProducts(query) {
        let category = 0;
        let selectedCategories = Array.from(document.querySelectorAll('.ctaegoryInput:checked')).map(input => input.value);
        if (selectedCategories.length != 0) {
          selectedCategories.forEach(c => {
            category |= categories[c]
          })
        }

        fetch(`/api/v1/products?name=${query}&categories=${category}`)
          .then(response => {
            if (!response.ok) {
              throw new Error(`HTTP error! status: ${response.status}`);
            }

            return response.json();
          })
          .then(result => {
            displayProductsWithPagination(result)
          })
          .catch(e => {
            console.error('An error occurred while loading the products:', e);
          });
      }

      function loadProducts() {
        fetch('/api/v1/products')
          .then(response => {
            if (!response.ok) {
              throw new Error(`HTTP error! status: ${response.status}`);
            }

            return response.json();
          })
          .then(result => {
            displayProductsWithPagination(result)
          })
          .catch(e => {
            console.error('An error occurred while loading the products:', e);
          });
      }

      function displayProductsWithPagination(result) {
        const productsDiv = document.getElementById('products');
        productsDiv.innerHTML = '';
        result.products.forEach(product => {
          const productDiv = document.createElement('div');

          productDiv.className = "productClass"
          productDiv.id = `productTile:${product.id}`



          productDiv.innerHTML = `
            <h2>${product.name}</h2>
            <p>${priceToString(product.price)}</p>
            <p>Rating: ${product.rating}</p>
          `;

          fetch(`/api/v1/products/${product.id}/images`)
            .then(response => {
              if (response.ok) {
                return response.json()
              }

              throw new Error(`HTTP error! status: ${response.status}`);
            })
            .then(images => {
              const img = document.createElement('img');
              img.src = images[0].data;
              img.alt = 'no image'
              productDiv.prepend(img);
            })
            .catch(error => {
              console.error(`Failed to fetch image for product ${product.id}:`, error);
              const img = document.createElement('img');
              img.src = "https://media.istockphoto.com/id/1206806317/vector/shopping-cart-icon-isolated-on-white-background.jpg?s=612x612&w=0&k=20&c=1RRQJs5NDhcB67necQn1WCpJX2YMfWZ4rYi1DFKlkNA=";
              img.alt = 'no image'
              productDiv.prepend(img);
            });

          productsDiv.appendChild(productDiv);
          let p = document.getElementById(`productTile:${product.id}`)
          p.addEventListener("click", function () {
            window.location.href = '/store/products/' + product.id
          })
        })
      }

      var modal = document.getElementById("categoryModal");

      document.getElementById("categoryButton").addEventListener("click", function () {
        modal.style.display = "block";
      });

      document.getElementById("closeCategories").addEventListener("click", function () {
        modal.style.display = "none";
      });

      window.addEventListener("click", function (event) {
        if (event.target == modal) {
          modal.style.display = "none";
        }
      });

      document.getElementById("searchButton").addEventListener('click', function () {
        var searchInput = document.getElementById("search").value;
        searchProducts(searchInput)
      });

      loadProducts();
    })

  function priceToString(price) {
    p1 = Math.floor(price.units / 100)
    p2 = price.units % 100
    cur = price.currency == 1 ? 'BGN' : '-'
    return `${p1},${p2} ${cur}`
  }
}