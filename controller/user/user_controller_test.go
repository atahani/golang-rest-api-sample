package user

import (
	"bytes"
	"encoding/json"
	"testing"

	"fmt"
	"net/http"
	"os"

	"github.com/atahani/golang-rest-api-sample/controller/client"
	"github.com/atahani/golang-rest-api-sample/models"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"github.com/atahani/golang-rest-api-sample/util/testhelper"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/test"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var testingProvider testhelper.TestingProvider
var newAppIdStr string
var authResponse models.AuthenticationResponse
var userEmail, userPassword, newPassword string

func TestSignUpNewUser(t *testing.T) {
	//assign the email and password
	userEmail = "ahmad.tahani@gmail.com"
	userPassword = "123456abcz"
	newPassword = "987654321mnbvcx"
	//get copy of db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	userController := NewUserController(session, testhelper.DB_TEST_NAME)
	session.DB(testhelper.DB_TEST_NAME).C(USER_COLLECTION_NAME).DropCollection()
	//define different cases
	reqBodyInvalidAppId := models.SignUpRequest{
		AppId:       "123124",
		FirstName:   "ahmad",
		LastName:    "tahani",
		DisplayName: "Ahmad",
		Email:       "ahmad.tahani@gmail.com",
		Password:    "123456abc",
	}
	reqBodyInvalidAppIdJ, _ := json.Marshal(reqBodyInvalidAppId)
	reqBodyNotHaveSomeField := models.SignUpRequest{
		AppId:     newAppIdStr,
		FirstName: "ahmad",
		Email:     "ahmad.tahani@gmail.com",
		Password:  "123456abc",
	}
	reqBodyNotHaveSomeFieldJ, _ := json.Marshal(reqBodyNotHaveSomeField)
	reqBodyInvalidEmail := models.SignUpRequest{
		AppId:       newAppIdStr,
		FirstName:   "ahmad",
		LastName:    "tahani",
		DisplayName: "Ahmad",
		Email:       "ahmad@tahani@@gmail.com",
		Password:    userPassword,
	}
	reqBodyInvalidEmailJ, _ := json.Marshal(reqBodyInvalidEmail)
	reqBodyInvalidPassword := models.SignUpRequest{
		AppId:       newAppIdStr,
		FirstName:   "ahmad",
		LastName:    "tahani",
		DisplayName: "Ahmad",
		Email:       userEmail,
		Password:    "12",
	}
	reqBodyInvalidPasswordJ, _ := json.Marshal(reqBodyInvalidPassword)
	reqBodyValid := models.SignUpRequest{
		AppId:       newAppIdStr,
		FirstName:   "ahmad",
		LastName:    "tahani",
		DisplayName: "Ahmad",
		Email:       userEmail,
		Password:    userPassword,
	}
	reqBodyValidJ, _ := json.Marshal(reqBodyValid)
	path := "/auth/signup"
	method := echo.POST
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidAppIdJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrNotValidClientInformation,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyNotHaveSomeFieldJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidEmailJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidPasswordJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyValidJ)),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		if err := userController.SignUpNewUser(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
		if c.expectedError == nil {
			//it's mean have successfully sign up in to system
			if err := json.NewDecoder(c.res.Body).Decode(&authResponse); err != nil {
				t.Error("can not get authentication response in sign up request !")
			}
		}
	}
}

func TestRefreshAccessToken(t *testing.T) {
	//get copy db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	userController := NewUserController(session, testhelper.DB_TEST_NAME)
	//define different cases
	refreshToke1 := models.RefreshTokenRequest{
		AppId:        newAppIdStr,
		RefreshToken: authResponse.RefreshToken,
	}
	refreshToke1J, _ := json.Marshal(refreshToke1)
	refreshToken2 := models.RefreshTokenRequest{
		AppId:        "aksdflhjas8",
		RefreshToken: authResponse.RefreshToken,
	}
	refreshToken2J, _ := json.Marshal(refreshToken2)
	refreshToken3 := models.RefreshTokenRequest{
		AppId: newAppIdStr,
	}
	refreshToken3J, _ := json.Marshal(refreshToken3)
	refreshToken4 := models.RefreshTokenRequest{
		AppId:        newAppIdStr,
		RefreshToken: "asdasfly89asfdj4hp",
	}
	refreshToken4J, _ := json.Marshal(refreshToken4)
	path := "/auth/refreshtoken"
	method := echo.POST
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewReader(refreshToke1J)),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(refreshToken2J)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrNotValidClientInformation,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(refreshToken3J)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(refreshToken4J)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrRefreshTokenIsNotValid,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		if err := userController.RefreshAccessToken(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
		if c.expectedError == nil {
			//it's mean have successfully sign in to system
			if err := json.NewDecoder(c.res.Body).Decode(&authResponse); err != nil {
				t.Error("can not get authentication response when refresh token is not valid !")
			}
		}
	}
}

func TestSignIn(t *testing.T) {
	//get copy of db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	userController := NewUserController(session, testhelper.DB_TEST_NAME)
	reqBodyInvalidAppId := models.SignInRequest{
		AppId:    "0981234",
		Email:    userEmail,
		Password: userPassword,
	}
	reqBodyInvalidAppIdJ, _ := json.Marshal(reqBodyInvalidAppId)
	reqBodyInvalidEmailAddress := models.SignInRequest{
		AppId:    newAppIdStr,
		Email:    "ahmad.tahani@@@.com",
		Password: userPassword,
	}
	reqBodyInvalidEmailAddressJ, _ := json.Marshal(reqBodyInvalidEmailAddress)
	reqBodyInvalidCredential := models.SignInRequest{
		AppId:    newAppIdStr,
		Email:    userEmail,
		Password: "091823qwerlimasdop",
	}
	reqBodyInvalidCredentialJ, _ := json.Marshal(reqBodyInvalidCredential)
	reqBodyValidRequest := models.SignInRequest{
		AppId:    newAppIdStr,
		Email:    userEmail,
		Password: userPassword,
	}
	reqBodyValidRequestJ, _ := json.Marshal(reqBodyValidRequest)
	path := "/auth/singin"
	method := echo.POST
	//define different cases
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidAppIdJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrNotValidClientInformation,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidEmailAddressJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidCredentialJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrNotValidCredentialInfo,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyValidRequestJ)),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		if err := userController.SignIn(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
		if c.expectedError == nil {
			//it's mean have successfully sign up in to system
			if err := json.NewDecoder(c.res.Body).Decode(&authResponse); err != nil {
				t.Error("can not get authentication response in sign in request !")
			}
		}
	}
}

func TestJWTAuthenticationMiddleware(t *testing.T) {
	//get copy of db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	//define jwt as handler since we test middleware alone
	jwt := JWTAuthenticationMiddleware(session, testhelper.DB_TEST_NAME)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	//define different case
	//define different cases
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		accessToken   string
		expectedError error
	}{
		{
			req:           test.NewRequest(echo.GET, "/", nil),
			res:           test.NewResponseRecorder(),
			accessToken:   fmt.Sprintf("%s %s", BEARER_AUTHENTICATION_TYPE, authResponse.AccessToken),
			expectedError: nil,
		},
		{
			req:           test.NewRequest(echo.GET, "/", nil),
			res:           test.NewResponseRecorder(),
			accessToken:   authResponse.AccessToken,
			expectedError: specialerror.ErrUnauthorized,
		},
		{
			req:           test.NewRequest(echo.GET, "/", nil),
			res:           test.NewResponseRecorder(),
			accessToken:   fmt.Sprintf("%s %s", BEARER_AUTHENTICATION_TYPE, "invalidaccessToken"),
			expectedError: specialerror.ErrUnauthorized,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderAuthorization, c.accessToken)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		if err := jwt(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestAuthorizeUserByRolesMiddleware(t *testing.T) {
	//get copy of db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	//define authorize role as handler since we test middleware alone
	authorizeRole := AuthorizeUserByRolesMiddleware([]string{"admin", "user"})(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	//define different cases
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		roles         []string
		expectedError error
	}{
		{
			req:           test.NewRequest(echo.GET, "/", nil),
			res:           test.NewResponseRecorder(),
			roles:         []string{"admin", "user"},
			expectedError: nil,
		},
		{
			req:           test.NewRequest(echo.GET, "/", nil),
			res:           test.NewResponseRecorder(),
			roles:         []string{"admin"},
			expectedError: specialerror.ErrCanNotAccessToTheseResource,
		},
		{
			req:           test.NewRequest(echo.GET, "/", nil),
			res:           test.NewResponseRecorder(),
			roles:         []string{"user"},
			expectedError: specialerror.ErrCanNotAccessToTheseResource,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		context.Set(ROLES_KEY, c.roles)
		if err := authorizeRole(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestUpdateUserProfile(t *testing.T) {
	//copy session of db
	session := testingProvider.Session.Copy()
	defer session.Close()
	//get the user_id from JWT token
	to, err := jwt.Parse(authResponse.AccessToken, func(token *jwt.Token) (interface{}, error) {
		//always check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("can't parse the access token !")
		}
		//return the key for validation
		return []byte(JWT_SIGNING_KEY_PHRASE), nil
	})
	if err != nil || to == nil || !to.Valid {
		t.Errorf("the access token is not valid or can't validate the access token !")
	}
	userId, ok := to.Claims["uid"].(string)
	if !ok {
		t.Errorf("can't get user_id from claims !")
	}
	userController := NewUserController(session, testhelper.DB_TEST_NAME)
	//define different cases
	reqBodyValid := models.User{
		FirstName:   "ahmad :)",
		LastName:    "tahani new ",
		DisplayName: ":) :)",
		Email:       userEmail,
	}
	reqBodyValidJ, _ := json.Marshal(reqBodyValid)
	reqBodyInvalidEmailAddress := models.User{
		FirstName:   "ahmad",
		LastName:    "tahani",
		DisplayName: "Ahmad",
		Email:       "me@@@gmail.com",
	}
	reqBodyInvalidEmailAddressJ, _ := json.Marshal(reqBodyInvalidEmailAddress)
	reqBodyInvalidFields := models.User{
		FirstName: "ahmad",
		LastName:  "tahani",
		//don't have displayName and email address
	}
	reqBodyInvalidFieldsJ, _ := json.Marshal(reqBodyInvalidFields)
	path := "/api/user/profile"
	method := echo.PUT
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyValidJ)),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidEmailAddressJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidFieldsJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		//set the user_id for context
		context.Set(USER_ID_KEY, bson.ObjectIdHex(userId))
		if err := userController.UpdateUserProfile(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

func TestChangeUserPassword(t *testing.T) {
	//copy db session
	session := testingProvider.Session.Copy()
	defer session.Close()
	//get the user_id from JWT token
	to, err := jwt.Parse(authResponse.AccessToken, func(token *jwt.Token) (interface{}, error) {
		//always check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("can't parse the access token !")
		}
		//return the key for validation
		return []byte(JWT_SIGNING_KEY_PHRASE), nil
	})
	if err != nil || to == nil || !to.Valid {
		t.Errorf("the access token is not valid or can't validate the access token !")
	}
	userId, ok := to.Claims["uid"].(string)
	if !ok {
		t.Errorf("can't get user_id from claims !")
	}
	userController := NewUserController(session, testhelper.DB_TEST_NAME)
	//define different cases
	reqBodyValid := models.ChangePasswordRequestModel{
		OldPassword: userPassword,
		Password:    newPassword,
	}
	reqBodyValidJ, _ := json.Marshal(reqBodyValid)
	reqBodyInvalidOldPassword := models.ChangePasswordRequestModel{
		OldPassword: "123098123klqwerlku89023",
		Password:    newPassword,
	}
	reqBodyInvalidOldPasswordJ, _ := json.Marshal(reqBodyInvalidOldPassword)
	reqBodyInvalidPassword := models.ChangePasswordRequestModel{
		OldPassword: userPassword,
		Password:    "",
	}
	reqBodyInvalidPasswordJ, _ := json.Marshal(reqBodyInvalidPassword)
	path := "/api/user/password"
	method := echo.POST
	cases := []struct {
		req           engine.Request
		res           *test.ResponseRecorder
		expectedError error
	}{
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyValidJ)),
			res:           test.NewResponseRecorder(),
			expectedError: nil,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidOldPasswordJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrNotValidCredentialInfo,
		},
		{
			req:           test.NewRequest(method, path, bytes.NewReader(reqBodyInvalidPasswordJ)),
			res:           test.NewResponseRecorder(),
			expectedError: specialerror.ErrSomeFieldAreNotValid,
		},
	}
	for _, c := range cases {
		c.req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		context := echo.NewContext(c.req, c.res, testingProvider.Echo)
		//set the user_id for context
		context.Set(USER_ID_KEY, bson.ObjectIdHex(userId))
		if err := userController.ChangeUserPassword(context); err != c.expectedError {
			t.Errorf("Error should %q \t but get %q", c.expectedError, err)
		}
	}
}

//create new client in db just for test
func createNewClientInDB(s *mgo.Session, e *echo.Echo) (*models.Client, error) {
	//copy db session
	session := s.Copy()
	defer session.Close()
	//add new client via function inside the manage_by_admin.go
	clientController := client.NewClientController(session, testhelper.DB_TEST_NAME)
	req := test.NewRequest(echo.POST, "/api/manage/client", bytes.NewBuffer([]byte(`{"name":"new client just for test"}`)))
	req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	res := test.NewResponseRecorder()
	context := echo.NewContext(req, res, e)
	if err := clientController.CreateNewClient(context); err != nil {
		return nil, err
	}
	//decode the body
	client := models.Client{}
	if err := json.NewDecoder(res.Body).Decode(&client); err != nil {
		return nil, err
	}
	//return new client information
	return &client, nil
}

func TestMain(m *testing.M) {
	//start of testing
	testingProvider = testhelper.TestingProvider{}
	testingProvider.StartTesting()
	//create the new client to signUp
	if cli, err := createNewClientInDB(testingProvider.Session, testingProvider.Echo); err != nil {
		fmt.Println("error happend in creating client !\n%s", err)
	} else {
		newAppIdStr = string(cli.AppId.Hex())
	}
	ret := m.Run()
	os.Exit(ret)
}
