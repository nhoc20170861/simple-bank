package api

import (
	"github.com/go-playground/validator/v10"

	util "github.com/nhoc20170861/simple-bank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
