package dao

import (
	"database/sql"

	"github.com/vladoiliev02/online-store/model"
)

const (
	selectAllUsers = `
		SELECT u.id, u.name, u.first_name, u.last_name, u.picture_url, u.email, u.created_at, 
			a.id, a.city, a.country, a.address, a.postal_code
		FROM users u
		LEFT JOIN addresses a ON u.address_id = a.id
	`

	selectUserByID = selectAllUsers +
		"WHERE u.id = $1;"

	selectUserByEmail = selectAllUsers +
		"WHERE u.email = $1;"

	insertUser = `
		INSERT INTO users(name, first_name, last_name, picture_url, email, address_id, created_at)
		VALUES($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING id, created_at;
	`

	updateUsersAddress = `
		UPDATE users
		SET address_id=$1
		WHERE id=$2;
	`

	deleteUser = `
		DELETE FROM users
		WHERE id=$1;
	`
)

type UserDAO struct {
	dao *DAO
	qe  queryExecutor
}

func NewUserDAO() *UserDAO {
	return newUserDAO(GetDAO().db)
}

func newUserDAO(qe queryExecutor) *UserDAO {
	return &UserDAO{
		dao: GetDAO(),
		qe:  qe,
	}
}

func (u *UserDAO) GetAll() ([]*model.User, error) {
	return executeMultiRowQuery(u.qe, u.scanUser,
		selectAllUsers)
}

func (u *UserDAO) GetByEmail(email string) (*model.User, error) {
	return executeSingleRowQuery(u.qe, u.scanUser,
		selectUserByEmail, email)
}

func (u *UserDAO) GetByID(id int64) (*model.User, error) {
	return executeSingleRowQuery(u.qe, u.scanUser,
		selectUserByID, id)
}

func (u *UserDAO) Create(user *model.User) (*model.User, error) {
	if user == nil {
		return nil, &DAOError{Query: insertUser, Message: "Nil User"}
	}

	return executeInTransaction(u.dao.db,
		func(tx *sql.Tx) (*model.User, error) {
			if (user.Address != model.Address{}) {
				addressTx := newAddressDAO(tx)
				address, err := addressTx.CreateAddress(&user.Address)
				if err != nil {
					return nil, err
				}
				user.Address = *address
			}

			return executeSingleRowQuery(tx, propertyScanner(user, &user.ID, &user.CreatedAt),
				insertUser, user.Name, user.FirstName, user.LastName, user.PictureURL, user.Email, user.Address.ID)
		})
}

func (u *UserDAO) Update(user *model.User) (*model.User, error) {
	if user == nil {
		return nil, &DAOError{Query: updateUsersAddress, Message: "Nil User"}
	}

	return executeInTransaction(u.dao.db,
		func(tx *sql.Tx) (*model.User, error) {
			addressTx := newAddressDAO(tx)
			address, err := addressTx.CreateAddress(&user.Address)
			if err != nil {
				return nil, err
			}

			err = executeNoRowsQuery(tx, updateUsersAddress, address.ID, user.ID)
			if err != nil {
				return nil, err
			}
			user.Address = *address

			return user, nil
		})
}

func (u *UserDAO) Delete(id int64) error {
	return executeNoRowsQuery(u.dao.db, deleteUser, id)
}

func (u *UserDAO) scanUser(row rowScanner) (*model.User, error) {
	var user model.User
	return propertyScanner(&user, &user.ID, &user.Name, &user.FirstName, &user.LastName, &user.PictureURL, &user.Email, &user.CreatedAt, &user.Address.ID, &user.Address.City, &user.Address.Country, &user.Address.Address, &user.Address.PostalCode)(row)
}
