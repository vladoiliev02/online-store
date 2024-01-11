window.onload = function () {
    initNavigation();

    let userId = window.location.pathname.split('/').pop();

    let currentUser;
    let pageUser;

    function getCategoriesMap() {
        var categoriesMap = Array.from(categories).reduce((map, input, index) => {
            map[input.value] = { code: Math.pow(2, index), el: input }
            return map;
        }, {});
        return categoriesMap
    }

    fetchWithStatusCheck('/api/v1/users/' + userId)
        .then(response => response.json())
        .then(user => {
            pageUser = user
            document.getElementById('name').textContent = user.name;
            document.getElementById('firstName').textContent = user.firstName;
            document.getElementById('lastName').textContent = user.lastName;
            document.getElementById('pictureURL').src = user.pictureUrl;
            document.getElementById('email').textContent = user.email;
            document.getElementById('createdAt').textContent = new Date(user.createdAt).toLocaleString();
            return user
        }).then(data => {
            fetchWithStatusCheck('/api/v1/users/me')
                .then(response => response.json())
                .then(user => {
                    currentUser = user
                    return user
                })
                .then(user => {
                    fetchWithStatusCheck(`/api/v1/products?userId=${pageUser.id}`)
                        .then(response => response.json())
                        .then(result => {
                            displayProductsWithPagination(result)
                        })
                        .catch(e => {
                            console.error('An error occurred while loading the products:', e);
                        });

                    if (pageUser.id === currentUser.id) {
                        document.getElementById("userActions").style.display = 'block'

                        var addProduct = document.getElementById("addProduct")

                        var modal = document.getElementById('create-product-modal');
                        var createButton = document.getElementById('post-product');
                        var cancelButton = document.getElementById('cancel');

                        var nameInput = document.getElementById('nameInput');
                        var descriptionInput = document.getElementById('description');
                        var priceInput = document.getElementById('price');
                        var currencyInput = document.getElementById('currency');
                        var quantityInput = document.getElementById('quantity');
                        var availableCheckBox = document.getElementById('availableCheckBox');
                        var categories = document.querySelectorAll('#categorySelect input');
                        var fileInput = document.getElementById("img-file-input");

                        function getCategoriesMap() {
                            var categoriesMap = Array.from(categories).reduce((map, input, index) => {
                                map[input.value] = { code: Math.pow(2, index), el: input }
                                return map;
                            }, {});
                            return categoriesMap
                        }

                        addProduct.onclick = function () {
                            modal.style.display = 'block';
                        }

                        cancelButton.onclick = function () {
                            modal.style.display = 'none';
                        }

                        fileInput.onchange = function () {
                            var files = fileInput.files;

                            var imagesDiv = document.getElementById('images');
                            imagesDiv.innerHTML = ''
                            for (var i = 0; i < files.length; i++) {
                                var img = document.createElement('img');
                                img.src = URL.createObjectURL(files[i]);
                                imagesDiv.appendChild(img);
                            }
                        }

                        createButton.onclick = function () {
                            const categoriesMap = getCategoriesMap()

                            var productName = nameInput.value;
                            var description = descriptionInput.value;
                            var priceStr = priceInput.value;
                            var currency = currencyInput.value;
                            var quantity = Number(quantityInput.value);
                            var available = Boolean(availableCheckBox.checked ? true : false);
                            var price = { units: Number(priceStr.replace(',', '')), currency: 1 };

                            var category = 0
                            for (let c in categoriesMap) {
                                const input = categoriesMap[c]
                                if (input.el.checked == true) {
                                    category |= input.code
                                }
                            }

                            const payload = JSON.stringify({
                                "name": productName,
                                description: description,
                                price: price,
                                quantity: quantity,
                                available: available,
                                category: category
                            });

                            fetchWithStatusCheck(`/api/v1/products`, {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json'
                                },
                                body: payload
                            })
                                .then(response => response.json())
                                .then(product => {
                                    var files = fileInput.files;

                                    let filesArray = Array.from(files)

                                    if (filesArray.length == 0) {
                                        displayProductsWithPagination({ products: [product] }, false)
                                        return product
                                    }

                                    filesArray.forEach(function (file, i) {
                                        var reader = new FileReader();

                                        reader.onloadend = function () {
                                            var base64Image = reader.result;
                                            var imageFormat = file.type.split('/')[1];

                                            const payload = JSON.stringify({
                                                data: base64Image,
                                                format: imageFormat
                                            });

                                            fetchWithStatusCheck('/api/v1/products/' + product.id + '/images', {
                                                method: 'POST',
                                                headers: {
                                                    'Content-Type': 'application/json'
                                                },
                                                body: payload
                                            })
                                                .then(response => response.json())
                                                .then(image => {
                                                    if (i == 0) {
                                                        displayProductsWithPagination({ products: [product] }, false)
                                                    }
                                                });
                                        }

                                        reader.readAsDataURL(file);
                                    })
                                })

                            cancelButton.click()
                        }

                        const ordersButton = document.getElementById('ordersButton');
                        const ordersModal = document.getElementById('ordersModal');
                        const ordersList = document.getElementById('ordersList');
                        const closeButton = document.querySelector('.close');

                        ordersButton.addEventListener('click', () => {
                            fetchWithStatusCheck('/api/v1/orders')
                                .then(response => response.json())
                                .then(orders => {
                                    ordersList.innerHTML = '';

                                    orders.forEach(order => {
                                        if (order.status != 1) {
                                            const div = document.createElement('div');
                                            const p = document.createElement('p');
                                            p.textContent = `Order ID: ${order.id}`;
                                            div.appendChild(p);

                                            div.addEventListener('click', () => {
                                                window.location.href = `/store/orders/${order.id}`;
                                            });

                                            ordersList.appendChild(div);
                                        }
                                    });

                                    ordersModal.style.display = 'block';
                                });
                        });

                        closeButton.addEventListener('click', () => {
                            ordersModal.style.display = 'none';
                        });
                    }
                })
        });


    function displayProductsWithPagination(result, resetProducts = true) {
        const productsDiv = document.getElementById('products');
        if (resetProducts) {
            productsDiv.innerHTML = '';
        }
        result.products.forEach(product => {
            const productDiv = document.createElement('div');
            productDiv.className = "productClass"

            productDiv.id = `productTile:${product.id}`

            p1 = Math.floor(product.price.units / 100)
            p2 = product.price.units % 100
            cur = product.price.currency == 1 ? 'BGN' : '-'

            productDiv.innerHTML = `
                <h2>${product.name}</h2>
                <p>${p1},${p2} ${cur}</p>
                <p>Rating: ${product.rating}</p>
              `;

            fetchWithStatusCheck(`/api/v1/products/${product.id}/images`, {}, false)
                .then(response => response.json())
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

}
