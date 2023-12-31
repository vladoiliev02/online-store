package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type NullInt64JSON sql.NullInt64
type NullStringJSON sql.NullString
type NullBoolJSON sql.NullBool
type NullFloat64JSON sql.NullFloat64

func (n NullInt64JSON) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int64)
	}
	return json.Marshal(nil)
}

func (n *NullInt64JSON) UnmarshalJSON(data []byte) error {
	var value *int64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		n.Valid = true
		n.Int64 = *value
	} else {
		n.Valid = false
	}

	return nil
}

func (n NullStringJSON) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	}
	return json.Marshal(nil)
}

func (n *NullStringJSON) UnmarshalJSON(data []byte) error {
	var value *string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		n.Valid = true
		n.String = *value
	} else {
		n.Valid = false
	}

	return nil
}

func (n NullBoolJSON) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Bool)
	}
	return json.Marshal(nil)
}

func (n *NullBoolJSON) UnmarshalJSON(data []byte) error {
	var value *bool
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		n.Valid = true
		n.Bool = *value
	} else {
		n.Valid = false
	}

	return nil
}

func (n NullFloat64JSON) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Float64)
	}
	return json.Marshal(nil)
}

func (n *NullFloat64JSON) UnmarshalJSON(data []byte) error {
	var value *float64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		n.Valid = true
		n.Float64 = *value
	} else {
		n.Valid = false
	}

	return nil
}

func (n *NullInt64JSON) Scan(value any) error {
	return ((*sql.NullInt64)(n)).Scan(value)
}

func (n NullInt64JSON) Value() (driver.Value, error) {
	return (sql.NullInt64(n)).Value()
}

func (n *NullStringJSON) Scan(value interface{}) error {
	return ((*sql.NullString)(n)).Scan(value)
}

func (n NullStringJSON) Value() (driver.Value, error) {
	return (sql.NullString(n)).Value()
}

func (n *NullBoolJSON) Scan(value interface{}) error {
	return ((*sql.NullBool)(n)).Scan(value)
}

func (n NullBoolJSON) Value() (driver.Value, error) {
	return (sql.NullBool(n)).Value()
}

func (n *NullFloat64JSON) Scan(value interface{}) error {
	return ((*sql.NullFloat64)(n)).Scan(value)
}

func (n NullFloat64JSON) Value() (driver.Value, error) {
	return (sql.NullFloat64(n)).Value()
}
