package testhelper

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/labstack/echo"

	"github.com/atahani/golang-rest-api-sample/util"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
)

const (
	DB_TEST_NAME = "golang_sample_test"
)

//utilities for testing used in unit testing
type TestingProvider struct {
	Session *mgo.Session
	Echo    *echo.Echo
	Router  *echo.Router
}

func (provider *TestingProvider) StartTesting() {
	provider.Session = getSession()
	//create new echo server
	provider.Echo = echo.New()
	provider.Router = provider.Echo.Router()
	//set custom binder to validate model
	bi := util.NewCustomBinderWithValidation()
	provider.Echo.SetBinder(bi)
	//set Custom Error Handler
	provider.Echo.SetHTTPErrorHandler(specialerror.CustomErrorHandler)
	provider.Echo.SetDebug(true)
}

func getSession() *mgo.Session {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Timeout:  60 * time.Second,
		Database: DB_TEST_NAME,
	}
	//first should set path the fake database
	//dbServer.SetPath(folderPath)
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		fmt.Printf("connection %s\n", err)
	}
	return session
}
