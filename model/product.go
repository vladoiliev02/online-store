package model

type ProductCategory int

const (
	Home ProductCategory = 1 << iota
	Clothing
	Shoes
	Sport
	Appliances
	Technology
	Entertainment
	Books
	Cars
	ProductCategoryMask = 1<<iota - 1
)

type OrderStatus int

const (
	InCart OrderStatus = iota + 1
	InProgress
	Completed
	Canceled
	InvalidOrderStatus
)

type User struct {
	ID         NullInt64JSON  `json:"id"`
	Name       NullStringJSON `json:"name"`
	FirstName  NullStringJSON `json:"firstName"`
	LastName   NullStringJSON `json:"LastName"`
	PictureURL NullStringJSON `json:"pictureUrl"`
	Email      NullStringJSON `json:"email"`
	Address    Address        `json:"address"`
	CreatedAt  NullStringJSON `json:"createdAt"`
}

type Product struct {
	ID           NullInt64JSON   `json:"id"`
	Name         NullStringJSON  `json:"name"`
	Description  NullStringJSON  `json:"description"`
	Price        Price           `json:"price"`
	Quantity     NullInt64JSON   `json:"quantity"`
	Category     ProductCategory `json:"category"`
	Available    NullBoolJSON    `json:"available"`
	Comments     []*Comment      `json:"comments"`
	Rating       NullFloat64JSON `json:"rating"`
	RatingsCount NullInt64JSON   `json:"ratingsCount"`
	Ratings      []*Rating       `json:"-"`
	CreatedAt    NullStringJSON  `json:"createdAt"`
	UserID       NullInt64JSON   `json:"userId"`
}

type Image struct {
	ID        NullInt64JSON `json:"id"`
	ProductID NullInt64JSON `json:"productId"`
	Data      []byte        `json:"data"`
}

type Item struct {
	ID        NullInt64JSON `json:"id"`
	ProductID NullInt64JSON `json:"productId"`
	OrderID   NullInt64JSON `json:"orderId"`
	Quantity  NullInt64JSON `json:"quantity"`
	Price     Price         `json:"price"`
}

type Order struct {
	ID           NullInt64JSON  `json:"id"`
	UserID       NullInt64JSON  `json:"userId"`
	Products     []*Item        `json:"products"`
	Status       OrderStatus    `json:"status"`
	Address      Address        `json:"address"`
	CreatedAt    NullStringJSON `json:"createdAt"`
	LatestUpdate NullStringJSON `json:"latestUpdate"`
}

type Invoice struct {
	ID         NullInt64JSON  `json:"id"`
	UserID     NullInt64JSON  `json:"userId"`
	Order      Order          `json:"order"`
	TotalPrice Price          `json:"totalPrice"`
	CreatedAt  NullStringJSON `json:"createdAt"`
}

type Address struct {
	ID         NullInt64JSON  `json:"id"`
	City       NullStringJSON `json:"city"`
	Country    NullStringJSON `json:"country"`
	Address    NullStringJSON `json:"address"`
	PostalCode NullStringJSON `json:"postalCode"`
}

type Comment struct {
	ID        NullInt64JSON  `json:"id"`
	User      User           `json:"user"`
	ProductID NullInt64JSON  `json:"productId"`
	Comment   NullStringJSON `json:"comment"`
	CreatedAt NullStringJSON `json:"createdAt"`
}

type Rating struct {
	UserID    NullInt64JSON `json:"userId"`
	ProductID NullInt64JSON `json:"productId"`
	Rating    NullInt64JSON `json:"rating"`
}

var (
	ProductCategories = map[ProductCategory]string{
		Home:          "Home",
		Clothing:      "Clothing",
		Shoes:         "Shoes",
		Sport:         "Sport",
		Appliances:    "Appliances",
		Technology:    "Technology",
		Entertainment: "Entertainment",
		Books:         "Books",
		Cars:          "Cars",
	}
)
