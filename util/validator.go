package util

import (
	"github.com/go-playground/validator/v10"
	"strconv"
)

func IsValidAmount(fl validator.FieldLevel) bool {
	amount, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	floatAmt, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return false
	}

	return floatAmt > 0
}

func IsValidCurrency(fl validator.FieldLevel) bool {
	currency, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return IsSupportedCurrency(currency)
}
