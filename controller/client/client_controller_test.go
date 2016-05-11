package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/atahani/golang-rest-api-sample/models"
	"github.com/atahani/golang-rest-api-sample/util/operationresult"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"github.com/atahani/golang-rest-api-sample/util/testhelper"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/test"
	"os"
)

var testingProvider testhelper.TestingProvider
var newAppIdStrForClient string

func TestCreateNewClient(t *testing.T) {
	//get copy of session
	session := testingProvider.Session.Copy()
	defer session.Close()
	clientController := NewClientController(session, testhelper.DB_TEST_NAME)
	//define different case
	path := "/api/manage/client"
	method := echo.POST
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewBuffer([]byte(`{"name":"client for web"}`))),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewBuffer([]byte(`{"platform_type":"android"}`))),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		if err := clientController.CreateNewClient(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
		if c.expectedError == nil {
			//decode response body
			client := models.Client{}
			if err := json.NewDecoder(c.res.Body).Decode(&client); err == nil {
				newAppIdStrForClient = string(client.AppId.Hex())
			}
		}
	}
}

func TestGetClientById(t *testing.T) {
	//since the path have id param should add it to Router
	testingProvider.Router.Add(echo.GET, "/api/manage/client/:id", nil, testingProvider.Echo)
	//get session
	session := testingProvider.Session.Copy()
	defer session.Close()
	clientController := NewClientController(session, testhelper.DB_TEST_NAME)
	//define different cases
	path := fmt.Sprintf("/api/manage/client/%s", newAppIdStrForClient)
	method := echo.GET
	cases := []struct {
		path          string
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			path:          fmt.Sprintf("%ssomeinvalid", path),
			req:           test.NewRequest(method, fmt.Sprintf("%ssomeinvalid", path), nil),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrNotValidItemId,
		},
		{
			path:          path,
			req:           test.NewRequest(method, path, nil),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		testingProvider.Router.Find(echo.GET, c.path, context)
		if err := clientController.GetClientById(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestUpdateClientById(t *testing.T) {
	//since the path have id param should add it to Router
	testingProvider.Router.Add(echo.PUT, "/api/manage/client/:id", nil, testingProvider.Echo)
	//get session
	session := testingProvider.Session.Copy()
	defer session.Close()
	clientController := NewClientController(session, testhelper.DB_TEST_NAME)
	//define different case
	path := fmt.Sprintf("/api/manage/client/%s", newAppIdStrForClient)
	method := echo.PUT
	cases := []struct {
		path          string
		req           engine.Request
		res           *test.ResponseRecorder
		responseBody  *operationresult.OperationResult
		expectedError error
	}{
		{
			path:          fmt.Sprintf("%ssomeinvalid", path),
			req:           test.NewRequest(method, fmt.Sprintf("%ssomeinvalid", path), nil),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.New("", ""),
			expectedError: specialerror.ErrNotValidItemId,
		},
		{
			path: path,
			req: test.NewRequest(method, path,
				bytes.NewBuffer([]byte(`{"name":"updated name"}`))),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.SuccessfullyUpdated,
			expectedError: nil,
		},
		{
			path: path,
			req: test.NewRequest(method, path,
				bytes.NewBuffer([]byte(`{"platform_type":"android"}`))),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.New("", ""),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		testingProvider.Router.Find(echo.PUT, c.path, context)
		if err := clientController.UpdateClientById(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestGetClients(t *testing.T) {
	//get session
	session := testingProvider.Session.Copy()
	defer session.Close()
	clientController := NewClientController(session, testhelper.DB_TEST_NAME)
	path := "/api/manage/client"
	req := test.NewRequest(echo.GET, path, nil)
	res := test.NewResponseRecorder()
	context := echo.NewContext(req, res, testingProvider.Echo)
	if err := clientController.GetClients(context); err != nil {
		t.Errorf("Error should %q \t but get %q", nil, err)
	}
	//check the number of clients
	result := []models.Client{}
	if err := json.NewDecoder(res.Body).Decode(&result); err == nil {
		if len(result) == 0 {
			t.Error("should at least one client in this get clients request !")
		}
	}
}

func TestDeleteClientById(t *testing.T) {
	//since the URL have id param should add it to Router
	testingProvider.Router.Add(echo.DELETE, "/api/manage/client/:id", nil, testingProvider.Echo)
	//copy db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	clientController := NewClientController(session, testhelper.DB_TEST_NAME)
	//define different cases
	path := fmt.Sprintf("/api/manage/client/%s", newAppIdStrForClient)
	method := echo.DELETE
	cases := []struct {
		path          string
		req           engine.Request
		res           *test.ResponseRecorder
		responseBody  *operationresult.OperationResult
		expectedError error
	}{
		{
			path:          fmt.Sprintf("%ssomeinvalid", path),
			req:           test.NewRequest(method, fmt.Sprintf("%ssomeinvalid", path), nil),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.New("", ""),
			expectedError: specialerror.ErrNotValidItemId,
		},
		{
			path:          path,
			req:           test.NewRequest(method, path, nil),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.SuccessfullyRemoved,
			expectedError: nil,
		},
		{
			path:          path,
			req:           test.NewRequest(method, path, nil),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.New("", ""),
			expectedError: specialerror.ErrNotFoundAnyItemWithThisId, //since the last case delete this client now notfound :)
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		testingProvider.Router.Find(echo.DELETE, c.path, context)
		if err := clientController.DeleteClientById(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestMain(m *testing.M) {
	//start of testing
	testingProvider = testhelper.TestingProvider{}
	testingProvider.StartTesting()
	ret := m.Run()
	os.Exit(ret)
}