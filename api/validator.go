package api

import (
	"github.com/go-playground/validator/v10"
	"tutorial.sqlc.dev/app/utils"
)

var validCurrencies validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return utils.IsValidCurrency(currency)
	}
	return false
}
