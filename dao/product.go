package dao

import (
	"database/sql"
	"errors"
	"online-store/model"
)

const (
	minPage     = 1
	minPageSize = 40
	maxPageSize = 80
)

const (
	selectProducts = `
		SELECT id, name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, created_at, user_id
		FROM products
	`

	selectProductsWithPagination = `
		SELECT id, name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, created_at, user_id, (SELECT count(*) FROM products WHERE available AND (category & $3) != 0) as count
		FROM products
		WHERE available AND (category & $3) != 0
		ORDER BY rating
		LIMIT $1 OFFSET $2
	`

	selectProductByID = selectProducts +
		" WHERE id = $1"

	selectProductByName = `
		SELECT id, name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, created_at, user_id, (SELECT count(*) FROM products WHERE name LIKE $1 AND available AND (category & $4) != 0) as count
		FROM products 
		WHERE name LIKE $1 AND available AND (category & $4) != 0
		ORDER BY rating
		LIMIT $2 OFFSET $3
	`

	selectProductsByUserID = `
		SELECT id, name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, created_at, user_id, (SELECT count(*) FROM products WHERE user_id = $1) as count
		FROM products
		WHERE user_id = $1
		ORDER BY rating
		LIMIT $2 OFFSET $3
	`

	insertProduct = `
		INSERT INTO products(name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 0, $8)
		RETURNING id, created_at
	`

	updateProduct = `
		UPDATE products 
		SET description = $1, price_units = $2, price_currency = $3, quantity = $4, category = $5, available = $6
		WHERE id = $7
		RETURNING name, rating, ratings_count, user_id
	`

	getRating = `
		SELECT user_id, product_id, rating
		FROM ratings
		WHERE user_id = $1 AND product_id = $2
	`

	insertRating = `
		INSERT INTO ratings(user_id, product_id, rating)
		VALUES ($1, $2, $3)
	`

	updateRating = `
		UPDATE ratings
		SET rating = $1
		WHERE user_id = $2 AND product_id = $3
	`

	updateProductNewRating = `
		UPDATE products 
		SET rating = (rating * ratings_count + $1) / (ratings_count + 1), ratings_count = ratings_count + 1 
		WHERE id = $2 
		RETURNING id, name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, created_at, user_id
	`

	updateProductExistingRating = `
		UPDATE products 
		SET rating = (rating * ratings_count + $1 - $2) / ratings_count
		WHERE id = $3
		RETURNING id, name, description, price_units, price_currency, quantity, category, available, rating, ratings_count, created_at, user_id
	`
)

type ProductDAO struct {
	dao *DAO
	qe  queryExecutor
}

func NewProductDAO() *ProductDAO {
	return newProductDAO(GetDAO().db)
}

func newProductDAO(qe queryExecutor) *ProductDAO {
	return &ProductDAO{
		dao: GetDAO(),
		qe:  qe,
	}
}

func (p *ProductDAO) GetAll(page, pageSize int, category model.ProductCategory) ([]*model.Product, int64, error) {
	pageSize, offset := getPageSizeAndOffset(pageSize, page)

	var count int64
	products, err := executeMultiRowQuery(p.qe,
		func(row rowScanner) (*model.Product, error) {
			var product model.Product
			return propertyScanner(&product, &product.ID, &product.Name, &product.Description, &product.Price.Units, &product.Price.Currency, &product.Quantity, &product.Category, &product.Available, &product.Rating, &product.RatingsCount, &product.CreatedAt, &product.UserID, &count)(row)
		},
		selectProductsWithPagination,
		pageSize, offset, category)

	return products, count, err
}

func (p *ProductDAO) GetByID(id int64) (*model.Product, error) {
	return executeSingleRowQuery(p.qe,
		scanProduct,
		selectProductByID,
		id)
}

func (p *ProductDAO) GetByNameLike(name string, page, pageSize int, category model.ProductCategory) ([]*model.Product, int64, error) {
	pageSize, offset := getPageSizeAndOffset(pageSize, page)

	var count int64
	products, err := executeMultiRowQuery(p.qe,
		func(row rowScanner) (*model.Product, error) {
			var product model.Product
			return propertyScanner(&product, &product.ID, &product.Name, &product.Description, &product.Price.Units, &product.Price.Currency, &product.Quantity, &product.Category, &product.Available, &product.Rating, &product.RatingsCount, &product.CreatedAt, &product.UserID, &count)(row)
		},
		selectProductByName,
		"%"+name+"%", pageSize, offset, category)

	return products, count, err
}

func (p *ProductDAO) GetByUserID(userID int64, page, pageSize int) ([]*model.Product, int64, error) {
	pageSize, offset := getPageSizeAndOffset(pageSize, page)

	var count int64
	products, err := executeMultiRowQuery(p.qe,
		func(row rowScanner) (*model.Product, error) {
			var product model.Product
			return propertyScanner(&product, &product.ID, &product.Name, &product.Description, &product.Price.Units, &product.Price.Currency, &product.Quantity, &product.Category, &product.Available, &product.Rating, &product.RatingsCount, &product.CreatedAt, &product.UserID, &count)(row)
		},
		selectProductsByUserID,
		userID, pageSize, offset)

	return products, count, err
}

func (p *ProductDAO) Create(product *model.Product) (*model.Product, error) {
	return executeSingleRowQuery(p.qe,
		propertyScanner(product, &product.ID, &product.CreatedAt),
		insertProduct,
		product.Name, product.Description, product.Price.Units, product.Price.Currency, product.Quantity,
		product.Category, product.Available, product.UserID)
}

func (p *ProductDAO) Update(product *model.Product) (*model.Product, error) {
	return executeSingleRowQuery(p.qe,
		propertyScanner(product, &product.Name, &product.Rating, &product.RatingsCount, &product.UserID),
		updateProduct,
		product.Description, product.Price.Units, product.Price.Currency, product.Quantity, product.Category, product.Available, product.ID)
}

func (p *ProductDAO) AddRating(rating *model.Rating) (*model.Product, error) {
	_, err := executeInTransaction(p.dao.db,
		func(tx *sql.Tx) (int, error) {
			var er model.Rating
			existingRating, err := executeSingleRowQuery(tx,
				propertyScanner(&er, &er.UserID, &er.ProductID, &er.Rating),
				getRating,
				rating.UserID, rating.ProductID,
			)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 1, err
			}

			if errors.Is(err, sql.ErrNoRows) {
				err = executeNoRowsQuery(tx, insertRating, rating.UserID, rating.ProductID, rating.Rating)
				if err != nil {
					return 1, err
				}

				err = executeNoRowsQuery(tx, updateProductNewRating, rating.Rating, rating.ProductID)
				if err != nil {
					return 1, err
				}
			} else {
				err = executeNoRowsQuery(tx, updateRating, rating.Rating, rating.UserID, rating.ProductID)
				if err != nil {
					return 1, err
				}

				err = executeNoRowsQuery(tx, updateProductExistingRating, rating.Rating, existingRating.Rating, rating.ProductID)
				if err != nil {
					return 1, err
				}
			}

			return 0, nil
		})

	if err != nil {
		return nil, &DAOError{Query: "Update product rating transaction", Message: "Error while updating the rating", Err: err}
	}

	return p.GetByID(rating.ProductID.Int64)
}

func scanProduct(row rowScanner) (*model.Product, error) {
	var product model.Product
	return propertyScanner(&product, &product.ID, &product.Name, &product.Description, &product.Price.Units, &product.Price.Currency, &product.Quantity, &product.Category, &product.Available, &product.Rating, &product.RatingsCount, &product.CreatedAt, &product.UserID)(row)
}

func getPageSizeAndOffset(pageSize, page int) (int, int) {
	if page < minPage {
		page = minPage
	}

	if pageSize < minPageSize {
		pageSize = minPageSize
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset := (page - 1) * pageSize
	return pageSize, offset
}
