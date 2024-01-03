window.onload = function () {
    initNavigation();

    let productId = window.location.pathname.split('/').pop();

    var currentUser

    var editButton = document.getElementById('edit-product');

    var product
    var p1
    var p2
    var cur

    fetch('/api/v1/users/me')
        .then(response => response.json())
        .then(user => {
            currentUser = user

            fetch('/api/v1/products/' + productId)
                .then(response => response.json())
                .then(data => {
                    product = data
                    p1 = Math.floor(data.price.units / 100)
                    p2 = data.price.units % 100
                    cur = data.price.currency == 1 ? 'BGN' : '-'

                    const categoriesMap = getCategoriesMap()

                    let categoryStr = ""
                    for (let c in categoriesMap) {
                        if ((categoriesMap[c].code & data.category) != 0) {
                            categoryStr = categoryStr + " " + c
                        }
                    }

                    document.getElementById('product-name').textContent = data.name;
                    document.getElementById('product-description').textContent = data.description;
                    document.getElementById('product-price').textContent = `${p1},${p2} ${cur}`;
                    document.getElementById('product-quantity').textContent = data.quantity;
                    document.getElementById('product-category').textContent = categoryStr;
                    document.getElementById('product-available').textContent = data.available;
                    document.getElementById('product-rating').textContent = data.rating;
                    document.getElementById('product-ratingsCount').textContent = data.ratingsCount;
                    document.getElementById('product-createdAt').textContent = new Date(data.createdAt).toLocaleString();

                    fetch('/api/v1/users/' + data.userId)
                        .then(response => response.json())
                        .then(data => {
                            document.getElementById('user-name').textContent = data.name;
                            document.getElementById('user-image').src = data.pictureUrl
                        })

                    if (currentUser.id === product.userId) {
                        editButton.style.display = 'block';
                    }

                    fetch('/api/v1/products/' + productId + '/images?limit=10')
                        .then(response => response.json())
                        .then(images => {
                            const imagesDiv = document.getElementById('product-images');
                            images.forEach(function (image) {
                                addImage(image, product, currentUser, productId, imagesDiv);
                            });
                        });

                    return data
                }).then(product => {
                    const addToCartButton = document.getElementById('add-to-cart');
                    const quantityInput = document.getElementById('quantity-input');

                    if (!product.available) {
                        document.getElementById('addToCartDiv').style.display = 'none';
                    } else {
                        addToCartButton.addEventListener('click', () => {
                            fetch('/api/v1/orders?status=1')
                                .then(response => response.json())
                                .then(orders => orders[0])
                                .then(cart => {
                                    fetch(`/api/v1/orders/${cart.id}/items`, {
                                        method: 'POST',
                                        headers: {
                                            'Content-Type': 'application/json',
                                        },
                                        body: JSON.stringify({
                                            productId: product.id,
                                            quantity: Number(quantityInput.value),
                                        }),
                                    }).then(response => {
                                        if (response.ok) {
                                            const p = document.createElement('p')
                                            p.innerHTML = 'Success'
                                            document.getElementById('addToCartDiv').appendChild(p)
                                        }
                                    });
                                })

                        });
                    }
                    return product
                }).then(product => {
                    const submitRatingButton = document.getElementById('submit-rating');
                    const ratingInput = document.getElementById('rating-input');

                    submitRatingButton.addEventListener('click', () => {
                        const rating = Number(ratingInput.value);

                        fetch(`/api/v1/products/${productId}`, {
                            method: 'PATCH',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({
                                rating: rating
                            }),
                        })
                            .then(response => {
                                if (response.ok) {
                                    return response.json()
                                }
                                return product
                            })
                            .then(response => {
                                product.rating = response.rating
                                product.ratingsCount = response.ratingsCount

                                document.getElementById('product-rating').textContent = product.rating;
                                document.getElementById('product-ratingsCount').textContent = product.ratingsCount;
                            })
                    });
                })



            fetch('/api/v1/products/' + productId + '/comments')
                .then(response => response.json())
                .then(comments => {
                    var commentsDiv = document.getElementById('product-comments');
                    comments.forEach(function (comment) {
                        addComment(commentsDiv, comment, comment.user);
                    });
                });

            var modal = document.getElementById('edit-product-modal');
            var updateButton = document.getElementById('update-product');
            var cancelButton = document.getElementById('cancel-edit');
            var confirmButton = document.getElementById('confirm-button');

            var descriptionInput = document.getElementById('description');
            var priceInput = document.getElementById('price');
            var currencyInput = document.getElementById('currency');
            var quantityInput = document.getElementById('quantity');
            var availableCheckBox = document.getElementById('availableCheckBox');
            var categories = document.querySelectorAll('#categorySelect input');

            editButton.addEventListener("click", function () {
                const categoriesMap = getCategoriesMap()

                modal.style.display = 'block';
                descriptionInput.value = product.description;
                priceInput.value = `${p1},${p2}`;
                currencyInput.value = cur;
                quantityInput.value = product.quantity;
                availableCheckBox.checked = product.available;
                for (let category in categoriesMap) {
                    const input = categoriesMap[category]
                    if ((product.category & input.code) != 0) {
                        input.el.checked = true
                    }
                }
            });


            updateButton.addEventListener("click", function () {
                const categoriesMap = getCategoriesMap()

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

                if (category == 0) {
                    category = product.category
                }

                const payload = JSON.stringify({
                    description: description,
                    price: price,
                    quantity: quantity,
                    available: available,
                    category: category
                });

                fetch(`/api/v1/products/${productId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: payload
                }).then(_ => {
                    window.location.href = '/store/products/' + productId;
                });

                cancelButton.click()
            });

            cancelButton.addEventListener("click", function () {
                modal.style.display = 'none';
            });

            var fileInput = document.getElementById("img-file-input");

            fileInput.onchange = function () {
                var files = fileInput.files;

                var imagesDiv = document.getElementById('images');
                imagesDiv.innerHTML = ''
                for (var i = 0; i < files.length; i++) {
                    var img = document.createElement('img');
                    img.src = URL.createObjectURL(files[i]);
                    imagesDiv.appendChild(img);
                }

                confirmButton.style.display = 'block';
            }

            confirmButton.addEventListener("click", function () {
                var files = fileInput.files;

                Array.from(files).forEach(function (file, i) {
                    var reader = new FileReader();

                    reader.onloadend = function () {
                        var base64Image = reader.result;
                        var imageFormat = file.type.split('/')[1];

                        const payload = JSON.stringify({
                            data: base64Image,
                            format: imageFormat
                        });
                        fetch('/api/v1/products/' + productId + '/images', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json'
                            },
                            body: payload
                        })
                            .then(response => response.json())
                            .then(image => {
                                const imagesDiv = document.getElementById('product-images');
                                addImage(image, product, currentUser, productId, imagesDiv);
                            });
                    }

                    reader.readAsDataURL(file);
                });
                cancelButton.click()
            });

            function getCategoriesMap() {
                var categoriesMap = Array.from(categories).reduce((map, input, index) => {
                    map[input.value] = { code: Math.pow(2, index), el: input }
                    return map;
                }, {});
                return categoriesMap
            }

            var addCommentButton = document.getElementById('add-comment');

            const postComment = function () {
                var commentInput = document.getElementById('comment-input');
                var comment = commentInput.value;

                fetch('/api/v1/products/' + productId + '/comments', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        comment: comment
                    })
                })
                    .then(response => response.json())
                    .then(comment => {
                        var commentsDiv = document.getElementById('product-comments');
                        addComment(commentsDiv, comment, comment.user);
                    });

                commentInput.value = '';
            }

            addCommentButton.addEventListener('click', postComment)

            function addComment(commentsDiv, comment, user) {
                let commentDiv = document.createElement('div');
                commentDiv.className = "comment-display-div";

                let userDiv = document.createElement('div');
                userDiv.className = "user-div";

                let userImg = document.createElement('img');
                userImg.src = user.pictureUrl;
                userDiv.appendChild(userImg);

                let userSpan = document.createElement('span');
                userSpan.textContent = user.name;
                userDiv.appendChild(userSpan);

                commentDiv.appendChild(userDiv);

                let commentSpan = document.createElement('span');
                commentSpan.textContent = comment.comment;
                commentSpan.style.width = '75%';
                commentDiv.appendChild(commentSpan);

                if (comment.user.id == currentUser.id) {
                    let deleteButton = document.createElement('button');
                    deleteButton.textContent = 'X';
                    deleteButton.style.backgroundColor = 'red';
                    deleteButton.style.color = 'white';
                    deleteButton.addEventListener('click', function () {
                        fetch('/api/v1/products/' + productId + '/comments/' + comment.id, {
                            method: 'DELETE'
                        })
                            .then(response => {
                                if (response.ok) {
                                    commentDiv.style.display = 'none';
                                }
                            });
                    });
                    commentDiv.appendChild(deleteButton);
                }

                commentsDiv.appendChild(commentDiv);
            }
        });
};

function addImage(image, product, currentUser, productId, imagesDiv) {
    let img = document.createElement('img');
    img.src = image.data;

    let imageContainer = document.createElement('div');
    imageContainer.className = "imageContainerDiv";
    imageContainer.appendChild(img);

    if (product.userId === currentUser.id) {
        let deleteButton = document.createElement('button');
        deleteButton.textContent = 'X';
        deleteButton.style.backgroundColor = 'red';
        deleteButton.style.color = 'white';
        deleteButton.addEventListener('click', function () {
            fetch('/api/v1/products/' + productId + '/images/' + image.id, {
                method: 'DELETE'
            })
                .then(response => {
                    if (response.ok) {
                        imagesDiv.removeChild(imageContainer);
                    }
                });
        });
        imageContainer.appendChild(deleteButton);
    }
    imagesDiv.appendChild(imageContainer);
}
