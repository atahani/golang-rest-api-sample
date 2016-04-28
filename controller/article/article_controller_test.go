package article

import (
	"github.com/atahani/golang-rest-api-sample/util/testhelper"
	"testing"
	"os"
	"github.com/labstack/echo"
	"github.com/atahani/golang-rest-api-sample/controller/user"
	"github.com/labstack/echo/test"
	"github.com/atahani/golang-rest-api-sample/models"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"github.com/labstack/echo/engine"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"bytes"
	"fmt"
	"github.com/atahani/golang-rest-api-sample/util/operationresult"
)

var testingProvider testhelper.TestingProvider
var newArticleIdStr string
var userIdObj bson.ObjectId

func TestCreateArticle(t *testing.T) {
	//get copy of session
	session := testingProvider.Session.Copy()
	defer session.Close()
	articleController := NewArticleController(session, testhelper.DB_TEST_NAME)
	//define different cases
	path := "/api/article"
	method := echo.POST
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewBuffer([]byte(`{"title":"new article"}`))),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewBuffer([]byte(`{"title":"new article","content":"some content ..."}`))),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		//set user_id for context
		context.Set(user.USER_ID_KEY, userIdObj)
		if err := articleController.CreateArticle(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
		if c.expectedError == nil {
			//decode response body
			article := models.Article{}
			if err := json.NewDecoder(c.res.Body).Decode(&article); err == nil {
				newArticleIdStr = string(article.Id.Hex())
			}
		}
	}
}

func TestGetArticleById(t *testing.T) {
	//since the path have id param should add it to routeer
	testingProvider.Router.Add(echo.GET, "/api/article/:id", nil, testingProvider.Echo)
	//get copy of db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	articleController := NewArticleController(session, testhelper.DB_TEST_NAME)
	path := fmt.Sprintf("/api/article/%s", newArticleIdStr)
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
		//set the user_id for context
		context.Set(user.USER_ID_KEY, userIdObj)
		if err := articleController.GetArticleById(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestUpdateArticleById(t *testing.T) {
	//add path with id to router
	testingProvider.Router.Add(echo.PUT, "/api/article/:id", nil, testingProvider.Echo)
	session := testingProvider.Session.Copy()
	defer session.Close()
	articleController := NewArticleController(session, testhelper.DB_TEST_NAME)
	//define different cases
	path := fmt.Sprintf("/api/article/%s", newArticleIdStr)
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
			req:           test.NewRequest(method, path, bytes.NewBuffer([]byte(`{"title":"updated article","content":"some content updated ..."}`))),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.SuccessfullyUpdated,
			expectedError: nil,
		},
		{
			path: path,
			req:           test.NewRequest(method, path, bytes.NewBuffer([]byte(`{"title":"new article"}`))),
			res:           test.NewResponseRecorder(),
			responseBody:  operationresult.New("", ""),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		testingProvider.Router.Find(method, c.path, context)
		//set the user_id for context
		context.Set(user.USER_ID_KEY, userIdObj)
		if err := articleController.UpdateArticleById(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestGetArticlesOfUser(t *testing.T) {
	session := testingProvider.Session.Copy()
	defer session.Close()
	articleController := NewArticleController(session, testhelper.DB_TEST_NAME)
	path := "/api/article"
	req := test.NewRequest(echo.GET, path, nil)
	res := test.NewResponseRecorder()
	context := echo.NewContext(req, res, testingProvider.Echo)
	//set the user_id for context
	context.Set(user.USER_ID_KEY, userIdObj)
	if err := articleController.GetArticlesOfUser(context); err != nil {
		t.Errorf("Error should %q \t but get %q", nil, err)
	} else {
		//check have at least one article
		result := [] models.Article{}
		if err := json.NewDecoder(res.Body).Decode(&result); err == nil {
			if len(result) == 0 {
				t.Error("should at least one article in get articles request")
			}
		}
	}
}

func TestDeleteArticleById(t *testing.T) {
	//since the path have id param should add it to routeer
	testingProvider.Router.Add(echo.DELETE, "/api/article/:id", nil, testingProvider.Echo)
	//get copy of db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	articleController := NewArticleController(session, testhelper.DB_TEST_NAME)
	path := fmt.Sprintf("/api/article/%s", newArticleIdStr)
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
			responseBody: operationresult.New("", ""),
			expectedError: specialerror.ErrNotValidItemId,
		},
		{
			path:          path,
			req:           test.NewRequest(method, path, nil),
			res:           test.NewResponseRecorder(),
			responseBody: operationresult.SuccessfullyRemoved,
			expectedError: nil,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		testingProvider.Router.Find(method, c.path, context)
		//set the user_id for context
		context.Set(user.USER_ID_KEY, userIdObj)
		if err := articleController.DeleteArticleById(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestMain(m *testing.M) {
	//start of testing
	testingProvider = testhelper.TestingProvider{}
	testingProvider.StartTesting()
	userIdObj = bson.NewObjectId()
	ret := m.Run()
	os.Exit(ret)
}