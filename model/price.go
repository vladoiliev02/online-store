package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Currency int

const (
	BGN Currency = iota + 1
	// USD
	// EUR
	InvalidCurrency
)

type Price struct {
	Units    int64    `json:"units,omitempty"`
	Currency Currency `json:"currency,omitempty"`
}

func NewPrice(units int64, currency Currency) Price {
	return Price{
		Units:    units,
		Currency: currency,
	}
}

func (p Price) Add(other Price) Price {
	if p.Currency != other.Currency {
		panic("Cannot add prices with different currencies")
	}

	p.Units += other.Units
	return p
}

func (p Price) Subtract(other Price) Price {
	if p.Currency != other.Currency {
		panic("Cannot subtract prices with different currencies")
	}

	p.Units -= other.Units
	return p
}

func (p Price) Multiply(factor float64) Price {
	p.Units = int64(float64(p.Units) * factor)
	return p
}

func (p Price) MultiplyInt(factor int) Price {
	p.Units *= int64(factor)
	return p
}

func (p Price) ToString() string {
	return fmt.Sprintf("%d.%d %d", p.Units/100, p.Units%100, p.Currency)
}

func FromString(s string) (Price, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return Price{}, fmt.Errorf("invalid price string format")
	}

	currency, err := strconv.Atoi(parts[1])
	if err != nil {
		return Price{}, fmt.Errorf("invalid units format: %v", err)
	}

	amountParts := strings.Split(parts[0], ".")
	if len(amountParts) != 2 {
		return Price{}, errors.New("invalid price string format")
	}

	units, err := strconv.ParseInt(amountParts[0], 10, 64)
	if err != nil {
		return Price{}, fmt.Errorf("invalid units format: %v", err)
	}

	subunits, err := strconv.ParseInt(amountParts[1], 10, 64)
	if err != nil {
		return Price{}, fmt.Errorf("invalid subunits format: %v", err)
	}

	return Price{
		Units:    units*100 + subunits,
		Currency: Currency(currency),
	}, nil
}
