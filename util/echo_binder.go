package util

import (
	"strings"
	"reflect"
	"encoding/json"

	"github.com/labstack/echo"
	"github.com/asaskevich/govalidator"

	"github.com/atahani/golang-rest-api-sample/util/specialerror"
)

//this is custom bind function for echo to validate struct

type customBinderWithValidation struct {
}

func NewCustomBinderWithValidation() *customBinderWithValidation {
	return &customBinderWithValidation{}
}

func (customBinderWithValidation) Bind(i interface{}, c echo.Context) error {
	rq := c.Request()
	ct := rq.Header().Get(echo.HeaderContentType)
	//first check the require fields
	if !strings.HasPrefix(ct, echo.MIMEApplicationJSON) {
		return echo.ErrUnsupportedMediaType
	}
	if err := json.NewDecoder(rq.Body()).Decode(i); err != nil {
		return specialerror.ErrSomeFieldAreNotValid
	}
	//data decoded now should check validation if it's struct
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		if isValid, err2 := govalidator.ValidateStruct(i); !isValid || err2 != nil {
			return specialerror.ErrSomeFieldAreNotValid
		}
	}
	return nil
}
