package model

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	maxCommentLength     = 512
	maxProductNameLength = 255
	maxEmailLength       = 255
	maxUsernameLength    = 255
)

type ValidationError struct {
	Message string `json:"message"`
	Err     error  `json:"error"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func ValidatePrice(price *Price) error {
	if price == nil {
		return &ValidationError{"Price: is nil", nil}
	}

	if price.Units <= 0 {
		return &ValidationError{"Price: units should be positive", nil}
	}

	if price.Currency <= 0 || price.Currency >= InvalidCurrency {
		return &ValidationError{"Price: currency is not valid", nil}
	}

	return nil
}

func ValidateProduct(product *Product, exists bool) error {
	if product == nil {
		return &ValidationError{"Product: is nil", nil}
	}

	if (exists && !product.ID.Valid) || (!exists && product.ID.Valid) {
		return &ValidationError{"Product: invalid ID", nil}
	}

	if product.Name.Valid {
		product.Name.String = strings.TrimSpace(product.Name.String)
		if product.Name.String == "" || len(product.Name.String) > maxProductNameLength {
			return &ValidationError{"Product: name cannot be empty", nil}
		}
	}

	if product.Quantity.Int64 <= 0 {
		return &ValidationError{"Product: quantity should be positive", nil}
	}

	if err := validateBitSet(int(product.Category), ProductCategoryMask, "product category"); err != nil {
		return &ValidationError{"Product: invalid category", nil}
	}

	if err := ValidatePrice(&product.Price); err != nil {
		return &ValidationError{"Product: invalid price", nil}
	}

	return nil
}

func ValidateUser(user *User, exists bool) error {
	if user == nil {
		return &ValidationError{"User: is nil", nil}
	}

	if (exists && !user.ID.Valid) || (!exists && user.ID.Valid) {
		return &ValidationError{"User: invalid ID", nil}
	}

	if user.FirstName.Valid {
		user.FirstName.String = strings.TrimSpace(user.FirstName.String)
		if user.FirstName.String == "" || len(user.FirstName.String) > maxUsernameLength {
			return &ValidationError{"User: first name cannot be empty", nil}
		}
	}

	if user.LastName.Valid {
		user.LastName.String = strings.TrimSpace(user.LastName.String)
		if user.LastName.String == "" || len(user.LastName.String) > maxUsernameLength {
			return &ValidationError{"User: last name cannot be empty", nil}
		}
	}

	if user.Name.Valid {
		user.Name.String = strings.TrimSpace(user.Name.String)
		if user.Name.String == "" || len(user.Name.String) > maxUsernameLength {
			return &ValidationError{"User: name cannot be empty", nil}
		}
	}

	if user.PictureURL.Valid {
		if _, err := url.Parse(user.PictureURL.String); err != nil {
			return &ValidationError{"User: picture url should be a valid url", err}
		}
	}

	if user.Email.Valid {
		user.Email.String = strings.TrimSpace(user.Email.String)
		if len(user.Email.String) > maxEmailLength || !isValidEmail(user.Email.String) {
			return &ValidationError{"User: invalid email address", nil}
		}
	}

	if err := ValidateAddress(&user.Address); err != nil {
		return &ValidationError{"User: invalid address", err}
	}

	return nil
}

func ValidateItem(item *Item, exists bool) error {
	if item == nil {
		return &ValidationError{"Item: is nil", nil}
	}

	if (exists && !item.ID.Valid) || (!exists && item.ID.Valid) {
		return &ValidationError{"Item: invalid ID", nil}
	}

	if !item.ProductID.Valid {
		return &ValidationError{"Item: invalid product ID", nil}
	}

	if !item.OrderID.Valid {
		return &ValidationError{"Item: invalid order ID", nil}
	}

	if !item.Quantity.Valid {
		return &ValidationError{"Item: invalid quantity", nil}
	}

	return nil
}

func ValidateImage(image *Image, exists bool) error {
	if image == nil {
		return &ValidationError{"Image: is nil", nil}
	}

	if (exists && !image.ID.Valid) || (!exists && image.ID.Valid) {
		return &ValidationError{"Image: invalid ID", nil}
	}

	if !image.ProductID.Valid {
		return &ValidationError{"Image: invalid product ID", nil}
	}

	if image.Data == "" || len(image.Data) == 0 {
		return &ValidationError{"Image: invalid data", nil}
	}

	return nil
}

func ValidateOrder(order *Order, exists bool) error {
	if order == nil {
		return &ValidationError{"Order: is nil", nil}
	}

	if (exists && !order.ID.Valid) || (!exists && order.ID.Valid) {
		return &ValidationError{"Order: invalid ID", nil}
	}

	if !order.UserID.Valid {
		return &ValidationError{"Order: invalid user ID", nil}
	}

	if !IsValidOrderStatus(order.Status) {
		return &ValidationError{"Order: invalid status", nil}
	}

	if err := ValidateAddress(&order.Address); err != nil {
		return &ValidationError{"Order: invalid address", err}
	}

	return nil
}

func IsValidOrderStatus(status OrderStatus) bool {
	return !(status <= 0 || status >= InvalidOrderStatus)
}

func ValidateInvoice(invoice *Invoice, exists bool) error {
	if invoice == nil {
		return &ValidationError{"Invoice: is nil", nil}
	}

	if (exists && !invoice.ID.Valid) || (!exists && invoice.ID.Valid) {
		return &ValidationError{"Invoice: invalid ID", nil}
	}

	if !invoice.UserID.Valid {
		return &ValidationError{"Invoice: invalid user ID", nil}
	}

	if err := ValidateOrder(&invoice.Order, exists); err != nil {
		return &ValidationError{"Invoice: invalid order", err}
	}

	if err := ValidatePrice(&invoice.TotalPrice); err != nil {
		return &ValidationError{"Invoice: invalid total price", err}
	}

	return nil
}

func ValidateAddress(address *Address) error {
	if address == nil {
		return &ValidationError{"Address: is nil", nil}
	}

	if address.ID.Valid && address.ID.Int64 < 0 {
		return &ValidationError{"Address: invalid ID", nil}
	}

	if address.City.Valid {
		address.City.String = strings.TrimSpace(address.City.String)
		if address.City.String == "" || len(address.City.String) > 255 {
			return &ValidationError{"Address: city cannot be empty", nil}
		}
	}

	if address.Country.Valid {
		address.Country.String = strings.TrimSpace(address.Country.String)
		if address.Country.String == "" || len(address.Country.String) > 255 {
			return &ValidationError{"Address: country cannot be empty", nil}
		}
	}

	if address.Address.Valid {
		address.Address.String = strings.TrimSpace(address.Address.String)
		if address.Address.String == "" || len(address.Address.String) > 255 {
			return &ValidationError{"Address: address cannot be empty", nil}
		}
	}

	if address.PostalCode.Valid {
		address.PostalCode.String = strings.TrimSpace(address.PostalCode.String)
		if address.PostalCode.String == "" || len(address.PostalCode.String) > 10 {
			return &ValidationError{"Address: postal code cannot be empty", nil}
		}
	}

	return nil
}

func ValidateComment(comment *Comment, exists bool) error {
	if comment == nil {
		return &ValidationError{"Comment: is nil", nil}
	}

	if (exists && !comment.ID.Valid) || (!exists && comment.ID.Valid) {
		return &ValidationError{"Comment: invalid ID", nil}
	}

	if err := ValidateUser(&comment.User, true); err != nil {
		return &ValidationError{"Comment: invalid user", err}
	}

	if !comment.ProductID.Valid {
		return &ValidationError{"Comment: invalid product ID", nil}
	}

	if comment.Comment.Valid {
		comment.Comment.String = strings.TrimSpace(comment.Comment.String)
		if comment.Comment.String == "" || len(comment.Comment.String) >= maxCommentLength {
			return &ValidationError{"Comment: invalid comment", nil}
		}
	}

	return nil
}

func ValidateRating(rating *Rating) error {
	if rating == nil {
		return &ValidationError{"Rating: is nil", nil}
	}

	if !rating.UserID.Valid {
		return &ValidationError{"Rating: invalid user ID", nil}
	}

	if !rating.ProductID.Valid {
		return &ValidationError{"Rating: invalid product ID", nil}
	}

	if rating.Rating.Int64 < 0 || rating.Rating.Int64 > 5 {
		return &ValidationError{"Rating: invalid rating value", nil}
	}

	return nil
}

func validateBitSet(bitset, mask int, name string) error {
	if bitset <= 0 || bitset&mask != bitset {
		return &ValidationError{
			Message: "Invalid " + name + " bitset",
		}
	}
	return nil
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return email == "" || regexp.MustCompile(emailRegex).MatchString(email)
}
