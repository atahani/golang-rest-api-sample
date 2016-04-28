package util

import (
	"github.com/labstack/echo"
	"strings"
	"encoding/json"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"reflect"
	"github.com/asaskevich/govalidator"
)

//this is custom bind function for echo to validate struct

type customBinderWithValidation struct {
}

func NewCustomBinderWithValidation() *customBinderWithValidation {
	return &customBinderWithValidation{}
}

func (customBinderWithValidation) Bind(i interface{}, c echo.Context) (err error) {
	rq := c.Request()
	ct := rq.Header().Get(echo.HeaderContentType)
	err = echo.ErrUnsupportedMediaType
	//first check the require fields
	if strings.HasPrefix(ct, echo.MIMEApplicationJSON) {
		if err = json.NewDecoder(rq.Body()).Decode(i); err != nil {
			err = specialerror.ErrSomeFieldAreNotValid
		}else {
			//data decoded now should check validation if it's struct
			val := reflect.ValueOf(i)
			if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				if isValid, err2 := govalidator.ValidateStruct(i); !isValid || err2 != nil {
					err = specialerror.ErrSomeFieldAreNotValid
				}
			}
		}
	}
	return
}