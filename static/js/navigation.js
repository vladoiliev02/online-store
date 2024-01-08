function initNavigation() {
    let storeHeader = document.createElement('div')
    storeHeader.id = 'storeHeader'

    storeHeader.innerHTML = `
        <span id="storeHeaderSpan">Online Store</span>
        <div id="userNav">
            <button id="cartButton">
                <i class="fa fa-shopping-cart">Cart</i>
            </button>

            <button id="profile">
                <img id="profilePic" src="" alt="Profile Picture">
            </button>

            <button id="logoutButton">Logout</button>
        </div>
    `

    const errorModal = document.createElement('div');
    errorModal.id = 'errorModal';
    errorModal.innerHTML = `
            <p id="errorMessage"></p>
            <button id="closeErrorModalButton">Close</button>
    `
    errorModal.style.display = 'none';

    let body = document.querySelectorAll('body')[0];
    body.prepend(storeHeader)
    body.appendChild(errorModal);

    document.getElementById('storeHeaderSpan').addEventListener('click', function () {
        window.location.href = '/store/index.html';
    });

    document.getElementById('logoutButton').addEventListener('click', function () {
        window.location.href = `/logout`;
    });

    fetchWithStatusCheck('/api/v1/users/me')
        .then(response => response.json())
        .then(currentUser => {
            document.getElementById('profilePic').src = currentUser.pictureUrl;
            let profileBtn = document.getElementById('profile')
            profileBtn.innerHTML = profileBtn.innerHTML + currentUser.firstName
            return currentUser
        })
        .then(currentUser => {
            document.getElementById('profile').addEventListener('click', function () {
                window.location.href = '/store/users/' + currentUser.id;
            });

            document.getElementById('cartButton').addEventListener('click', function () {
                fetchWithStatusCheck(`/api/v1/orders?status=1`)
                    .then(data => data.json())
                    .then(orders => {
                        window.location.href = '/store/orders/' + orders[0].id;
                    })
            });
        });
}

function fetchWithStatusCheck(input, init = null, displayError = true) {
    let promise = fetch(input, init)
        .then(response => {
            if (!response.ok) {
                return response.json().then(error => {
                    if (displayError) {
                        handleError(error);
                    }
                    throw error;
                });
            }
            return response;
        })

    return {
        then: function (callback) {
            promise = promise.then(data => {
                if (callback) {
                    try {
                        return callback(data);
                    } catch (error) {
                        if (displayError) {
                            handleError(error);
                        }
                        throw error;
                    }
                }
            });
            return this;
        },
        catch: function (onRejected) {
            promise = promise.catch(error => {
                if (displayError) {
                    handleError(error)
                };
                if (onRejected) {
                    onRejected(error)
                };
            });
            return this;
        }
    }
}

function handleError(error) {
    console.error(error);
    const errorModal = document.getElementById('errorModal');
    const errorMessage = document.getElementById('errorMessage');
    const closeErrorModalButton = document.getElementById('closeErrorModalButton');

    errorMessage.textContent = error.message;
    errorModal.style.display = 'flex';
    errorModal.style["flex-direction"] = 'column';
    errorModal.style["align-items"] = 'center';
    errorModal.style["justify-content"] = 'center';

    closeErrorModalButton.addEventListener('click', () => {
        errorModal.style.display = 'none';
    });
}