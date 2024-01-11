window.onload = function () {
    initNavigation()

    const orderStatus = {
        1: 'In Cart',
        2: 'In Progress',
        3: 'Completed',
        4: 'Canceled',
        5: 'Invalid Order Status'
    }

    fetchWithStatusCheck('/api/v1/users/me')
        .then(response => response.json())
        .then(currentUser => {
            const orderId = window.location.pathname.split('/').pop();

            fetchWithStatusCheck(`/api/v1/orders/${orderId}/items`)
                .then(response => response.json())
                .then(order => {
                    if (order.userId === currentUser.id) {
                        if (order.status != 1) {
                            fetchWithStatusCheck(`/api/v1/orders/${orderId}/invoice`)
                                .then(response => response.json())
                                .then(invoice => {
                                    document.getElementById('details').style = `
                                        display: flex;
                                        flex-direction: column;
                                    `;
                                    document.getElementById('orderId').textContent = order.id;
                                    document.getElementById('orderStatus').textContent = orderStatus[order.status];
                                    document.getElementById('orderAddress').textContent = `${order.address.city}, ${order.address.country}, ${order.address.address}, ${order.address.postalCode}`;
                                    document.getElementById('orderCreatedAt').textContent = new Date(order.createdAt).toLocaleString();
                                    document.getElementById('orderLatestUpdate').textContent = new Date(order.latestUpdate).toLocaleString();

                                    document.getElementById('invoiceId').textContent = invoice.id;
                                    document.getElementById('totalPrice').textContent = priceToString(invoice.totalPrice);
                                });
                        }

                        const productsDiv = document.getElementById('productsDiv');

                        order.products.forEach(product => {
                            fetchWithStatusCheck(`/api/v1/products/${product.productId}`)
                                .then(response => response.json())
                                .then(productDetails => {
                                    const productDiv = document.createElement('div');
                                    productDiv.className = 'productItem'

                                    const nameParagraph = document.createElement('p');
                                    nameParagraph.textContent = `Name: ${productDetails.name}`;
                                    productDiv.appendChild(nameParagraph);

                                    const quantityParagraph = document.createElement('p');
                                    quantityParagraph.textContent = `Quantity: ${product.quantity}`;
                                    productDiv.appendChild(quantityParagraph);

                                    const priceParagraph = document.createElement('p');
                                    priceParagraph.textContent = `Price: ${priceToString(product.price)}`;
                                    productDiv.appendChild(priceParagraph);

                                    productDiv.addEventListener('click', () => {
                                        window.location.href = `/store/products/${product.productId}`;
                                    });

                                    if (order.status == 1) {
                                        const deleteButton = document.createElement('button');
                                        deleteButton.textContent = 'x';
                                        deleteButton.style.backgroundColor = 'red';
                                        deleteButton.style.color = 'white'
                                        deleteButton.addEventListener('click', (event) => {
                                            event.stopPropagation();

                                            fetchWithStatusCheck(`/api/v1/orders/${order.id}/items/${product.id}`, {
                                                method: 'DELETE',
                                            }).then(response => {
                                                if (response.ok) {
                                                    productsDiv.removeChild(productDiv);
                                                }
                                            });
                                        });
                                        productDiv.appendChild(deleteButton);
                                    }

                                    productsDiv.appendChild(productDiv);
                                });
                        });

                        const buyButton = document.getElementById('buy');

                        if (order.status != 1) {
                            buyButton.style.display = 'none';
                        }else {
                            const purchaseButton = document.getElementById('purchase');
                            const purchaseModal = document.getElementById('purchaseModal');
                            const cancelButton = document.getElementById('cancel');
                            const closeButton = document.querySelector('.close');

                            cancelButton.addEventListener('click', () => {
                                purchaseModal.style.display = 'none';
                            });

                            closeButton.addEventListener('click', () => {
                                purchaseModal.style.display = 'none';
                            });

                            buyButton.addEventListener('click', () => {
                                purchaseModal.style.display = 'block';
                            });

                            purchaseButton.addEventListener('click', () => {
                                const city = document.getElementById('city').value;
                                const country = document.getElementById('country').value;
                                const address = document.getElementById('address').value;
                                const postalCode = document.getElementById('postalCode').value;

                                let addressObject = {
                                    city: city,
                                    country: country,
                                    address: address,
                                    postalCode: postalCode,
                                }

                                order.address = addressObject
                                order.status = 2

                                fetchWithStatusCheck(`/api/v1/orders/${orderId}`, {
                                    method: 'PUT',
                                    headers: {
                                        'Content-Type': 'application/json',
                                    },
                                    body: JSON.stringify(order),
                                })
                                    .then(response =>  response.json())
                                    .then(order => {
                                        purchaseModal.style.display = 'none';
                                        window.location.href = `/store/orders/${order.id}`
                                    });
                            });
                        }
                    }
                });
        });

    function priceToString(price) {
        p1 = Math.floor(price.units / 100)
        p2 = price.units % 100
        cur = price.currency == 1 ? 'BGN' : '-'
        return `${p1},${p2} ${cur}`
    }
}
