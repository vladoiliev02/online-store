package dao

import (
	"database/sql"
	"errors"
	"online-store/model"
)

const (
	selectAddressByID = `
		SELECT id, city, country, address, postal_code
		FROM addresses
		WHERE id = $1`

	selectAddress = `
		SELECT id, city, country, address, postal_code
		FROM addresses
		WHERE city = $1
			AND country = $2
			AND address = $3
			AND postal_code = $4`

	insertAddress = `
		INSERT INTO addresses(city, country, address, postal_code)
		VALUES($1, $2, $3, $4)
		RETURNING id`
)

type AddressDAO struct {
	dao *DAO
	qe  queryExecutor
}

func NewAddressDAO() *AddressDAO {
	return newAddressDAO(GetDAO().db)
}

func newAddressDAO(qe queryExecutor) *AddressDAO {
	return &AddressDAO{
		dao: GetDAO(),
		qe:  qe,
	}
}

func (a *AddressDAO) GetAddressByID(id int) (*model.Address, error) {
	return executeSingleRowQuery(a.qe, scanAddress,
		selectAddressByID, id)
}

func (a *AddressDAO) CreateAddress(address *model.Address) (*model.Address, error) {
	existingAddress, err := a.GetAddress(address)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, &DAOError{Query: insertAddress, Message: "Failed to check if address already exists", Err: err}
	}

	if existingAddress != nil {
		return existingAddress, nil
	}

	return executeSingleRowQuery(a.qe, propertyScanner(address, &address.ID),
		insertAddress, address.City, address.Country, address.Address, address.PostalCode)
}

func (a *AddressDAO) GetAddress(address *model.Address) (*model.Address, error) {
	return executeSingleRowQuery(a.qe, scanAddress,
		selectAddress, address.City, address.Country, address.Address, address.PostalCode)
}

func scanAddress(row rowScanner) (*model.Address, error) {
	var address model.Address
	return propertyScanner(&address, &address.ID, &address.City, &address.Country, &address.Address, &address.PostalCode)(row)
}
