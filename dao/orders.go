package dao

import (
	"database/sql"
	"errors"
	"online-store/model"
)

const (
	selectOrders = `
		SELECT o.id, o.user_id, o.status, o.created_at,
			a.id, a.city, a.country, a.address, a.postal_code
		FROM orders o
		JOIN addresses a ON a.id = o.address_id
	`

	selectOrderByID = selectOrders + " WHERE id = $1"

	selectOrdersByUserID = selectOrders + " WHERE user_id = $1"

	selectOrdersByUserIDAndStatus = selectOrders + " WHERE user_id = $1 AND status = $2"

	insertOrder = `
		INSERT INTO orders(user_id, status, address_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	updateOrder = `
		UPDATE orders
		SET status = $1, address_id = $2, latest_update = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, created_at, latest_update
	`
)

type OrderDAO struct {
	dao     *DAO
	qe      queryExecutor
	itemDAO *ItemDAO
}

func NewOrderDAO() *OrderDAO {
	return newOrderDAO(GetDAO().db)
}

func newOrderDAO(qe queryExecutor) *OrderDAO {
	return &OrderDAO{
		dao:     GetDAO(),
		qe:      qe,
		itemDAO: NewItemDAO(),
	}
}

func (o *OrderDAO) GetByID(id int64) (*model.Order, error) {
	return executeSingleRowQuery(o.qe, scanOrder, selectOrderByID, id)
}

func (o *OrderDAO) GetByUserID(userID int64) ([]*model.Order, error) {
	return executeMultiRowQuery(o.qe, scanOrder,
		selectOrdersByUserID, userID)
}

func (o *OrderDAO) GetByUserIDAndStatus(userID int64, status model.OrderStatus) ([]*model.Order, error) {
	return executeMultiRowQuery(o.qe, scanOrder,
		selectOrdersByUserIDAndStatus, userID, status)
}

func (o *OrderDAO) Create(order *model.Order) (*model.Order, error) {
	return executeInTransaction(o.dao.db,
		func(tx *sql.Tx) (*model.Order, error) {
			if (order.Address != model.Address{}) {
				addressTx := newAddressDAO(tx)
				address, err := addressTx.CreateAddress(&order.Address)
				if err != nil {
					return nil, err
				}

				order.Address = *address
			}

			return executeSingleRowQuery(o.qe, scanIDAndTimestamps(order),
				insertOrder, order.UserID, order.Status, order.Address.ID)
		})

}

func (o *OrderDAO) Update(order *model.Order) (*model.Order, error) {
	return executeInTransaction(o.dao.db,
		func(tx *sql.Tx) (*model.Order, error) {
			orderTx := newOrderDAO(tx)
			existingOrder, err := orderTx.GetByID(order.ID.Int64)
			if err != nil {
				return nil, err
			}

			if (order.Status == model.Canceled && existingOrder.Status == model.InCart) ||
				(order.Status != model.Canceled && order.Status-existingOrder.Status != 1) {
				return nil, &DAOError{Query: updateOrder, Message: "Invalid order status update", Err: err}
			}

			if existingOrder.Status == model.InCart && order.Status != model.InCart {
				if err := model.ValidateAddress(&order.Address); err != nil {
					return nil, &DAOError{Query: updateOrder, Message: "Invalid order address", Err: err}
				}

				orderTx.Create(&model.Order{
					UserID: existingOrder.UserID,
					Status: model.InCart,
				})

				orderPrice, err := orderTx.calculatePrice(order)
				if err != nil {
					return nil, &DAOError{Query: updateOrder, Message: "Error calculating order price", Err: err}
				}

				invoiceTx := newInvoiceDAO(tx)
				invoiceTx.Create(&model.Invoice{
					UserID:     existingOrder.UserID,
					Order:      *order,
					TotalPrice: orderPrice,
				})
			}

			if (order.Address != model.Address{}) {
				addressTx := newAddressDAO(tx)
				address, err := addressTx.CreateAddress(&order.Address)
				if err != nil {
					return nil, err
				}

				order.Address = *address
			} else {
				order.Address.ID = existingOrder.Address.ID
			}

			return executeSingleRowQuery(tx,
				scanIDAndTimestamps(order),
				updateOrder,
				order.Status, order.Address.ID, order.ID)
		})
}

func (o *OrderDAO) calculatePrice(order *model.Order) (model.Price, error) {
	var err error
	order, err = o.LoadItems(order)
	if err != nil {
		return model.Price{}, err
	}

	if len(order.Products) < 1 {
		return model.Price{}, errors.New("no products for order")
	}

	price := model.NewPrice(0, order.Products[0].Price.Currency)
	for _, item := range order.Products {
		price = price.Add(item.Price.MultiplyInt(int(item.Quantity.Int64)))
	}

	return price, nil
}

func (o *OrderDAO) LoadItems(order *model.Order) (*model.Order, error) {
	items, err := o.itemDAO.GetByOrderID(order.ID.Int64)
	if err != nil {
		return nil, err
	}

	order.Products = items
	return order, nil
}

func (o *OrderDAO) AddItem(userID int64, item *model.Item) (*model.Item, error) {
	return executeInTransaction(o.dao.db,
		func(tx *sql.Tx) (*model.Item, error) {
			orderTx := newOrderDAO(tx)
			order, err := orderTx.GetCart(userID)
			if err != nil {
				return nil, err
			}
			item.OrderID = order.ID

			order, err = orderTx.LoadItems(order)
			if err != nil {
				return nil, err
			}

			for _, p := range order.Products {
				if p.ProductID == item.ProductID {
					item.ID = p.ID
					item.Quantity.Int64 += p.Quantity.Int64
					break
				}
			}

			productTx := newProductDAO(tx)
			product, err := productTx.GetByID(item.ProductID.Int64)
			if err != nil {
				return nil, err
			}
			item.Price = product.Price

			itemDAO := newItemDAO(tx)
			if item.ID.Valid {
				item, err = itemDAO.Update(item)
			} else {
				item, err = itemDAO.Create(item)
			}

			if err != nil {
				return nil, err
			}

			return item, nil
		})
}

func (o *OrderDAO) RemoveItem(userID int64, itemID int64) error {
	_, err := executeInTransaction(o.dao.db,
		func(tx *sql.Tx) (*model.Item, error) {
			orderTx := newOrderDAO(tx)
			order, err := orderTx.GetCart(userID)
			if err != nil {
				return nil, err
			}

			order, err = orderTx.LoadItems(order)
			if err != nil {
				return nil, err
			}

			itemTx := newItemDAO(tx)
			for _, p := range order.Products {
				if p.ID.Int64 == itemID {
					itemTx.Delete(itemID)
					break
				}
			}

			return nil, nil
		})

	return err
}

func (o *OrderDAO) GetCart(userID int64) (*model.Order, error) {
	orders, err := o.GetByUserIDAndStatus(userID, model.InCart)
	if err != nil {
		return nil, err
	}
	if len(orders) > 1 {
		return nil, &DAOError{Query: "Get Cart", Message: "Error finding users cart", Err: nil}
	}
	if len(orders) == 0 {
		order := &model.Order{}
		order.UserID.Scan(userID)
		order.Status = model.InCart
		o.Create(order)
	}

	return orders[0], nil
}

func scanOrder(row rowScanner) (*model.Order, error) {
	var order model.Order
	return propertyScanner(&order,
		&order.ID, &order.UserID, &order.Status, &order.CreatedAt,
		&order.Address.ID, &order.Address.City, &order.Address.Country, &order.Address.Address, &order.Address.PostalCode)(row)
}

func scanIDAndTimestamps(order *model.Order) func(rowScanner) (*model.Order, error) {
	return propertyScanner(order, &order.ID, &order.CreatedAt, &order.LatestUpdate)
}
