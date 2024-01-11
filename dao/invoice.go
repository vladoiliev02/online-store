package dao

import "github.com/vladoiliev02/online-store/model"

const (
	selectInvoices = `
		SELECT i.id, i.user_id, i.total_price_units, i.total_price_currency, i.created_at,
			o.id, o.user_id, o.status, o.created_at,
			a.id, a.city, a.country, a.address, a.postal_code
		FROM invoices i
		JOIN orders o ON o.id = i.order_id
		LEFT JOIN addresses a ON a.id = o.address_id
	`

	selectInvoiceByID = selectInvoices + " WHERE id = $1"

	selectInvoicesByUserID = selectInvoices + " WHERE user_id = $1"

	selectInvoicesByOrderID = selectInvoices + " WHERE order_id = $1"

	insertInvoice = `
		INSERT INTO invoices(user_id, order_id, total_price_units, total_price_currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
)

type InvoiceDAO struct {
	dao *DAO
	qe  queryExecutor
}

func NewInvoiceDAO() *InvoiceDAO {
	return newInvoiceDAO(GetDAO().db)
}

func newInvoiceDAO(qe queryExecutor) *InvoiceDAO {
	return &InvoiceDAO{
		dao: GetDAO(),
		qe:  qe,
	}
}

func (i *InvoiceDAO) GetByID(id int64) (*model.Invoice, error) {
	return executeSingleRowQuery(i.qe, scanInvoice,
		selectInvoiceByID, id)
}

func (i *InvoiceDAO) GetByUserID(userID int64) ([]*model.Invoice, error) {
	return executeMultiRowQuery(i.qe, scanInvoice,
		selectInvoicesByUserID, userID)
}

func (i *InvoiceDAO) GetByOrderID(orderID int64) (*model.Invoice, error) {
	return executeSingleRowQuery(i.qe, scanInvoice,
		selectInvoicesByOrderID, orderID)
}

func (i *InvoiceDAO) Create(invoice *model.Invoice) (*model.Invoice, error) {
	return executeSingleRowQuery(i.qe, propertyScanner(invoice, &invoice.ID, &invoice.CreatedAt),
		insertInvoice, invoice.UserID, invoice.Order.ID, invoice.TotalPrice.Units, invoice.TotalPrice.Currency)
}

func scanInvoice(row rowScanner) (*model.Invoice, error) {
	var invoice model.Invoice
	return propertyScanner(&invoice,
		&invoice.ID, &invoice.UserID, &invoice.TotalPrice.Units, &invoice.TotalPrice.Currency, &invoice.CreatedAt,
		&invoice.Order.ID, &invoice.Order.UserID, &invoice.Order.Status, &invoice.Order.CreatedAt,
		&invoice.Order.Address.ID, &invoice.Order.Address.City, &invoice.Order.Address.Country, &invoice.Order.Address.Address, &invoice.Order.Address.PostalCode)(row)
}
