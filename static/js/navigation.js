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

    document.querySelectorAll('body')[0].prepend(storeHeader)

    document.getElementById('storeHeaderSpan').addEventListener('click', function () {
        window.location.href = '/store/index.html';
    });

    document.getElementById('logoutButton').addEventListener('click', function () {
        window.location.href = `/logout`;
    });

    fetch('/api/v1/users/me')
        .then(response => response.json())
        .then(data => {
            document.getElementById('profilePic').src = data.pictureUrl;
            profileBtn = document.getElementById('profile')
            profileBtn.innerHTML = profileBtn.innerHTML + data.firstName
            return data
        })
        .then(currentUser => {
            document.getElementById('profile').addEventListener('click', function () {
                window.location.href = '/store/users/' + currentUser.id;
            });

            document.getElementById('cartButton').addEventListener('click', function () {
                fetch(`/api/v1/orders?status=1`)
                    .then(data => data.json())
                    .then(orders => {
                        window.location.href = '/store/orders/' + orders[0].id;
                    })
            });
        });
}