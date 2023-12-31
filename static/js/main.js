document.getElementById('search').addEventListener('input', function (e) {
  searchProducts(e.target.value);
});

document.getElementById('profile').addEventListener('click', function () {
  window.location.href = '/static/profile';
});

function searchProducts(query) {
  // Implement API call to search products by name
}

function filterProducts(category) {
  // Implement API call to filter products by category
}

function loadProducts() {
  fetch('/api/v1/products')
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return response.json();
    })
    .then(products => {
      const productsDiv = document.getElementById('products');
      productsDiv.innerHTML = '';
      products.forEach(product => {
        const productDiv = document.createElement('div');
        productDiv.innerHTML = `
            <h2>${product.name}</h2>
            <p>${product.description}</p>
            <p>Price: ${product.price}</p>
            <p>Quantity: ${product.quantity}</p>
            <p>Category: ${product.category}</p>
            <p>Available: ${product.available}</p>
            <p>Rating: ${product.rating}</p>
            <p>Ratings Count: ${product.ratingsCount}</p>
            <p>Created At: ${product.createdAt}</p>
        `;
        productsDiv.appendChild(productDiv);

      })
    })
    .catch(e => {
      console.error('An error occurred while loading the products:', e);
    });
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
  if (searchInput != "") {
    console.log("Search for: ", searchInput);
    searchProducts(searchInput)
  }
});

fetch('/api/v1/users/me')
  .then(response => response.json())
  .then(data => {
    console.log(data);
    document.getElementById('profilePic').src = data.pictureUrl;
  });

loadProducts();