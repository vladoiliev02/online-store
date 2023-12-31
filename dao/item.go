package dao

import "online-store/model"

const (
	selectItems = `
		SELECT id, product_id, order_id, quantity, price_units, price_currency
		FROM items
	`

	selectItemsByOrderID = selectItems + " WHERE order_id = $1"

	selectItemByOrderIDAndProductID = selectItems + " WHERE order_id = $1 AND product_id = $2"

	insertItem = `
		INSERT INTO items(product_id, order_id, quantity, price_units, price_currency)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	updateItem = `
		UPDATE items
		SET order_id = $1, quantity = $2, price_units = $3, price_currency = $4
		WHERE id = $6
		RETURNING id, product_id, order_id, quantity, price_units, price_currency
	`

	deleteItem = `
		DELETE FROM items
		WHERE id = $1
	`
)

type ItemDAO struct {
	dao *DAO
	qe  queryExecutor
}

func NewItemDAO() *ItemDAO {
	return newItemDAO(GetDAO().db)
}

func newItemDAO(qe queryExecutor) *ItemDAO {
	return &ItemDAO{
		dao: GetDAO(),
		qe:  qe,
	}
}

func (i *ItemDAO) GetByOrderIDAndProductID(orderID, productID int64) (*model.Item, error) {
	return executeSingleRowQuery(i.qe, scanItem,
		selectItemByOrderIDAndProductID, orderID, productID)
}

func (i *ItemDAO) GetByOrderID(orderID int64) ([]*model.Item, error) {
	return executeMultiRowQuery(i.qe, scanItem,
		selectItemsByOrderID, orderID)
}

func (i *ItemDAO) Create(item *model.Item) (*model.Item, error) {
	return executeSingleRowQuery(i.qe, propertyScanner(item, &item.ID),
		insertItem, item.ProductID, item.OrderID, item.Quantity, item.Price.Units, item.Price.Currency)

}

func (i *ItemDAO) Update(item *model.Item) (*model.Item, error) {
	return executeSingleRowQuery(i.qe, scanItem,
		updateItem, item.OrderID, item.Quantity, item.Price.Units, item.Price.Currency, item.ID)
}

func (i *ItemDAO) Delete(id int64) error {
	return executeNoRowsQuery(i.qe, deleteItem, id)
}

func scanItem(row rowScanner) (*model.Item, error) {
	var item model.Item
	return propertyScanner(&item, &item.ID, &item.ProductID, &item.OrderID, &item.Quantity, &item.Price.Units, &item.Price.Currency)(row)
}
