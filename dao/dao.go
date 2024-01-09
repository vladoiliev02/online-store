package dao

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var (
	options *DBOptions
	dao     *DAO
)

type DBOptions struct {
	DriverName string
	ConnStr    string
}

type DAO struct {
	db *sql.DB
}

func Init(dbOptions *DBOptions) {
	if options != nil {
		panic("Database is already configured")
	}

	options = dbOptions
}

func GetDAO() *DAO {
	if dao == nil {
		db, err := sql.Open(options.DriverName, options.ConnStr)

		if err != nil {
			panic(err.Error())
		}

		dao = &DAO{
			db: db,
		}

		_, err = db.Exec("SELECT 1")
		if !dao.IsReady() {
			panic(err.Error())
		}
	}

	return dao
}

func (d *DAO) IsReady() bool {
	_, err := dao.db.Exec("SELECT 1")
	return err != nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

type queryExecutor interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

func executeMultiRowQuery[T any](db queryExecutor, rowScanningFunc func(rowScanner) (T, error), query string, queryArgs ...any) ([]T, error) {
	log.Println("Executing:", query)

	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		return nil, &DAOError{Query: query, Message: "Error executing multi row query", Err: err}
	}
	defer rows.Close()

	objects := make([]T, 0)
	for rows.Next() {
		object, err := rowScanningFunc(rows)
		if err != nil {
			return nil, &DAOError{Query: query, Message: "Error scanning orders by user ID", Err: err}
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func executeSingleRowQuery[T any](db queryExecutor, rowScanningFunc func(rowScanner) (T, error), query string, queryArgs ...any) (T, error) {
	row := db.QueryRow(query, queryArgs...)

	object, err := rowScanningFunc(row)
	if err != nil {
		return object, &DAOError{Query: query, Message: "Error querying single row", Err: err}
	}

	return object, nil
}

func executeNoRowsQuery(db queryExecutor, query string, queryArgs ...any) error {
	_, err := db.Exec(query, queryArgs...)
	if err != nil {
		return &DAOError{Query: query, Message: "Error executing query returning no rows", Err: err}
	}

	return nil
}

func executeInTransaction[T any](db *sql.DB, transactionalFunc func(*sql.Tx) (T, error)) (T, error) {
	tx, err := db.Begin()
	if err != nil {
		var invalid T
		return invalid, &DAOError{Query: "BEGIN", Message: "Error starting transaction", Err: err}
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := transactionalFunc(tx)
	if err != nil {
		return result, &DAOError{Query: "transactionalFunction", Message: "Error executing the provided function in a transaction", Err: err}
	}

	err = tx.Commit()
	if err != nil {
		return result, &DAOError{Query: "COMMIT", Message: "Error committing transaction", Err: err}
	}

	return result, nil
}

func propertyScanner[T any](obj T, args ...any) func(rowScanner) (T, error) {
	return func(row rowScanner) (T, error) {
		err := row.Scan(args...)
		if err != nil {
			return obj, err
		}

		return obj, nil
	}
}

type DAOError struct {
	Query   string `json:"-"`
	Message string `json:"-"`
	Err     error  `json:"-"`
}

func (e *DAOError) Error() string {
	return fmt.Sprintf("DAOError: %s, query: %s, caused by: %s", e.Message, e.Query, e.Err.Error())
}

func (e *DAOError) Unwrap() error {
	return e.Err
}
