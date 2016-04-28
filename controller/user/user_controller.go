package user

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/atahani/golang-rest-api-sample/controller/client"
	"github.com/atahani/golang-rest-api-sample/models"
	"github.com/atahani/golang-rest-api-sample/util"
	"github.com/atahani/golang-rest-api-sample/util/operationresult"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mcuadros/go-defaults.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	USER_COLLECTION_NAME = "users"
	ACCESS_TOKEN_COLLECTION_NAME = "accessTokens"
	JWT_SIGNING_KEY_PHRASE = "sectet_keys_to_hash_jwt_token"
	BEARER_AUTHENTICATION_TYPE = "Bearer"
	ROLES_KEY = "roles"
	USER_ID_KEY = "user_id"
	TOKEN_ID_KEY = "token_id"
	TRUSTED_APP_ID_KEY = "trusted_app_id"
)

type UserController struct {
	Session *mgo.Session
	DBName  string
}

func NewUserController(s *mgo.Session, dbName string) *UserController {
	return &UserController{s, dbName}
}

//echo middleware for checking JWT token is valid and authorize request
func JWTAuthenticationMiddleware(s *mgo.Session, dbName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header().Get(echo.HeaderAuthorization)
			l := len(BEARER_AUTHENTICATION_TYPE)
			he := specialerror.ErrUnauthorized
			if len(authHeader) > l + 1 && authHeader[:l] == BEARER_AUTHENTICATION_TYPE {
				token := string(authHeader[l + 1:])
				t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					//always check the signing method
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, he
					}
					//return the key for validation
					return []byte(JWT_SIGNING_KEY_PHRASE), nil
				})
				if err == nil && t.Valid {
					//get copy of database session
					session := s.Copy()
					defer session.Close()
					accessToken := models.AccessToken{}
					if err := session.DB(dbName).C(ACCESS_TOKEN_COLLECTION_NAME).Find(bson.M{"token": token}).One(&accessToken); err != nil {
						if err == mgo.ErrNotFound {
							return he
						} else {
							return specialerror.ErrInternalServerError
						}
					}
					//get the user and check is enable or not
					user := models.User{}
					if err := session.DB(dbName).C(USER_COLLECTION_NAME).FindId(accessToken.UserId).One(&user); err != nil {
						return specialerror.ErrInternalServerError
					}
					if user.IsEnable {
						//set some information that need in routes handler
						c.Set(USER_ID_KEY, user.Id)
						c.Set(ROLES_KEY, user.Roles)
						c.Set(TOKEN_ID_KEY, accessToken.Id)
						c.Set(TRUSTED_APP_ID_KEY, accessToken.TrustedAppId)
						//process the next and finish this middleware
						return next(c)
					} else {
						return specialerror.ErrUserIsDisable
					}
				}
			}
			return he
		}
	}
}

//echo middleware to check is this user have role
func AuthorizeUserByRolesMiddleware(roles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			data := c.Get(ROLES_KEY)
			if roleSlice, ok := data.([]string); ok {
				//check is have these role or not
				for _, v := range roles {
					if !util.IsStringInSlice(v, roleSlice) {
						return specialerror.ErrCanNotAccessToTheseResource
					}
				}
				//it's mean the user have all of the roles
				//process the next and finish this middleware
				return next(c)
			} else {
				return specialerror.ErrInternalServerError
			}
		}
	}
}

func (uc UserController) SignUpNewUser(c echo.Context) error {
	//create the user model
	signUpModel := models.SignUpRequest{}
	if err := c.Bind(&signUpModel); err != nil {
		return err
	}
	//get copy of session
	session := uc.Session.Copy()
	defer session.Copy()
	//TODO : please NOTE should check is't new email for this user ? if yes ? send verification email to this email
	//first check is already have user with this email address > unique or not
	if count, err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).Find(bson.M{"email": strings.ToLower(signUpModel.Email)}).Count(); err != nil {
		return specialerror.ErrInternalServerError
	} else if count != 0 {
		return specialerror.ErrAlreadyHaveUserWithThisEmailAddress
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpModel.Password), bcrypt.DefaultCost)
	if err != nil {
		return specialerror.ErrInternalServerError
	}
	//check client information is valid or not
	cliController := client.NewClientController(uc.Session, uc.DBName)
	if isWebClient, err := cliController.ClientAuthorization(uc.Session, uc.DBName, signUpModel.AppId, signUpModel.AppKey); err != nil {
		return err
	} else {
		//create new user and assign attributes
		u := models.User{
			Id:             bson.NewObjectId(),
			FirstName:      signUpModel.FirstName,
			LastName:       signUpModel.LastName,
			DisplayName:    signUpModel.DisplayName,
			Email:          strings.ToLower(signUpModel.Email),
			HashedPassword: string(hashedPassword),
			Roles:          []string{"user"},
			JoinedAt:       time.Now(),
			UpdatedAt:      time.Now(),
		}
		//set defaults values for user model
		defaults.SetDefaults(&u)
		//store new user into database
		if err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).Insert(u); err != nil {
			return specialerror.ErrInternalServerError
		}
		//should generate access token and send it
		if authResponse, err := generateAccessToken(session, uc.DBName, &u, bson.ObjectIdHex(signUpModel.AppId), bson.NewObjectId(), signUpModel.DeviceModel, false, isWebClient); err != nil {
			return err
		} else {
			//return the authentication response
			c.JSON(http.StatusOK, &authResponse)
			return nil
		}
	}
}

//authenticate the user with credential information email/username with password and generate access token
func (uc UserController) SignIn(c echo.Context) error {
	//get copy of database session
	session := uc.Session.Copy()
	defer session.Close()
	//check the credential information if it's valid
	signInRequest := models.SignInRequest{}
	//populate user credential information from request
	//the binder check if struct is not valid return err
	if err := c.Bind(&signInRequest); err != nil {
		return err
	}
	//check client information is valid or not
	cliController := client.NewClientController(uc.Session, uc.DBName)
	if isWebClient, err := cliController.ClientAuthorization(uc.Session, uc.DBName, signInRequest.AppId, signInRequest.AppKey); err != nil {
		return err
	} else {
		var user models.User
		//get user by email
		if err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).Find(bson.M{"email": strings.ToLower(signInRequest.Email)}).One(&user); err != nil {
			if err == mgo.ErrNotFound {
				return specialerror.ErrNotValidCredentialInfo
			} else {
				return specialerror.ErrInternalServerError
			}
		}
		//check user password with hashed password in db
		if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(signInRequest.Password)); err != nil {
			return specialerror.ErrNotValidCredentialInfo
		}
		//it's mean the credential information is valid so should generate JWT token as send it as JSON
		if authResponse, err := generateAccessToken(session, uc.DBName, &user, bson.ObjectIdHex(signInRequest.AppId), bson.NewObjectId(), signInRequest.DeviceModel, false, isWebClient); err != nil {
			return err
		} else {
			//return the authentication response
			c.JSON(http.StatusOK, &authResponse)
			return nil
		}
	}
}

//refresh the user access token
func (ac UserController) RefreshAccessToken(c echo.Context) error {
	//get copy of database session
	session := ac.Session.Copy()
	defer session.Close()
	refreshTokenRequest := models.RefreshTokenRequest{}
	//the binder check if struct is not valid return err
	if err := c.Bind(&refreshTokenRequest); err != nil {
		return err
	}
	cliController := client.NewClientController(ac.Session, ac.DBName)
	//check client information is valid or not
	if isWebClient, err := cliController.ClientAuthorization(session, ac.DBName, refreshTokenRequest.AppId, refreshTokenRequest.AppKey); err != nil {
		return err
	} else {
		//check is refresh token valid or not
		user := models.User{}
		if err := session.DB(ac.DBName).C(USER_COLLECTION_NAME).Find(bson.M{"trusted_apps._client": bson.ObjectIdHex(refreshTokenRequest.AppId), "trusted_apps.refresh_token": refreshTokenRequest.RefreshToken}).One(&user); err != nil {
			if err == mgo.ErrNotFound {
				return specialerror.ErrRefreshTokenIsNotValid
			} else {
				return specialerror.ErrInternalServerError
			}
		}
		//get the trusted app
		var trustedAppId bson.ObjectId
		for _, trustedApp := range user.TrustedApps {
			if trustedApp.RefreshToken == refreshTokenRequest.RefreshToken {
				trustedAppId = trustedApp.Id
			}
		}
		//generate the access token
		if authResponse, err := generateAccessToken(session, ac.DBName, &user, bson.ObjectIdHex(refreshTokenRequest.AppId), trustedAppId, "", true, isWebClient); err != nil {
			return err
		} else {
			//return the authentication response
			c.JSON(http.StatusOK, &authResponse)
			return nil
		}
	}
}

//update user model such as email , first, last, display name
//TODO : should implement user change image profile, get uploaded image resize it and save it as user image profile
func (uc UserController) UpdateUserProfile(c echo.Context) error {
	userId, ok := c.Get(USER_ID_KEY).(bson.ObjectId)
	if !ok {
		return specialerror.ErrInternalServerError
	}
	u := models.User{}
	//get some fields such as email, first,last,display name from payload
	if err := c.Bind(&u); err != nil {
		return err
	}
	//copy session db
	session := uc.Session.Copy()
	defer session.Close()
	//TODO : please NOTE should check is't new email for this user ? if yes ? send verification email to this email
	//check is email address unique or not
	if count, err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).Find(bson.M{"email": strings.ToLower(u.Email), "_id": bson.M{"$ne": userId}}).Count(); err != nil {
		fmt.Println(err)
		return specialerror.ErrInternalServerError
	} else if count != 0 {
		return specialerror.ErrAlreadyHaveUserWithThisEmailAddress
	}
	userUpdateSet := bson.M{
		"first_name":   u.FirstName,
		"last_name":    u.LastName,
		"display_name": u.DisplayName,
		"email":        u.Email,
		"updated_at":   time.Now(),
	}
	//update the user profile in one query
	if err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).UpdateId(userId, bson.M{"$set": userUpdateSet}); err != nil {
		return specialerror.ErrInternalServerError
	}
	c.JSON(http.StatusOK, operationresult.SuccessfullyUpdated)
	return nil
}

//change password with authorized user NOT reset password
//TODO : should implement reset password flow like send reset password link to user by email
func (uc UserController) ChangeUserPassword(c echo.Context) error {
	chPasswordReqModel := models.ChangePasswordRequestModel{}
	if err := c.Bind(&chPasswordReqModel); err != nil {
		return err
	}
	//get user id from c
	userId, ok := c.Get(USER_ID_KEY).(bson.ObjectId)
	if !ok {
		return specialerror.ErrInternalServerError
	}
	//copy db session
	session := uc.Session.Copy()
	defer session.Close()
	//get user model from db to check password and update it
	u := models.User{}
	if err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).FindId(userId).One(&u); err != nil {
		return specialerror.ErrInternalServerError
	}
	//check old password is valid or not
	if err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(chPasswordReqModel.OldPassword)); err != nil {
		return specialerror.ErrNotValidCredentialInfo
	}
	//hash new password and save it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(chPasswordReqModel.Password), bcrypt.DefaultCost)
	if err != nil {
		return specialerror.ErrInternalServerError
	}
	u.HashedPassword = string(hashedPassword)
	u.UpdatedAt = time.Now()
	if err := session.DB(uc.DBName).C(USER_COLLECTION_NAME).UpdateId(userId, &u); err != nil {
		return specialerror.ErrInternalServerError
	}
	//inform user the password successfully changed
	c.JSON(http.StatusOK, operationresult.PasswordSuccessfullyChanged)
	return nil
}

func generateAccessToken(s *mgo.Session, dbName string, u *models.User, clientId, trustedAppId bson.ObjectId, deviceModel string, isRefreshToken, isWebClient bool) (*models.AuthenticationResponse, error) {
	//define access token model
	accessToken := models.AccessToken{
		Id:           bson.NewObjectId(),
		UserId:       u.Id,
		TrustedAppId: trustedAppId,
	}
	expireIn := time.Hour * time.Duration(util.GenerateRandomNumber(24, 72))
	accessToken.ExpireAt = time.Now().Add(expireIn)
	//new JWT
	token := jwt.New(jwt.SigningMethodHS256)
	//set headers
	token.Header["type"] = "JWT"
	token.Claims["exp"] = accessToken.ExpireAt.Unix()
	token.Claims["uid"] = u.Id.Hex()
	token.Claims["roles"] = u.Roles
	token.Claims["name"] = u.DisplayName
	token.Claims["img"] = u.ImageFileName
	token.Claims["aid"] = trustedAppId.Hex()
	token.Claims["tid"] = accessToken.Id.Hex()
	if sToken, err := token.SignedString([]byte(JWT_SIGNING_KEY_PHRASE)); err != nil {
		return nil, specialerror.ErrInternalServerError
	} else {
		//assign access Token
		accessToken.Token = sToken
		//generate the refresh token
		if refreshToken, err := util.GenerateNewRefreshToken(); err != nil {
			return nil, specialerror.ErrInternalServerError
		} else {
			//save trusted app for this user and save the access token in database
			//check is web client
			if isRefreshToken {
				//so already have trusted app for this user
				foundIt := false
				for _, trustedApp := range u.TrustedApps {
					if trustedApp.Id == trustedAppId {
						foundIt = true
						trustedApp.RefreshToken = refreshToken
					}
				}
				if !foundIt {
					return nil, specialerror.ErrInternalServerError
				}
			} else {
				if isWebClient {
					//search in trustedApps have already this webClient
					foundIt := false
					for _, trustedApp := range u.TrustedApps {
						if trustedApp.ClientId == clientId {
							foundIt = true
							trustedApp.RefreshToken = refreshToken
						}
					}
					if !foundIt {
						//so should create new trusted app
						newTrustedAppForWebClient := models.TrustedApp{
							Id:           trustedAppId,
							ClientId:     clientId,
							RefreshToken: refreshToken,
							GrantedAt:    time.Now(),
						}
						//assign this trusted app for web client
						u.TrustedApps = append(u.TrustedApps, newTrustedAppForWebClient)
					}
				} else {
					//first check is already have any trusted app with this client and device model
					foundIt := false
					for _, trustedApp := range u.TrustedApps {
						if trustedApp.ClientId == clientId && trustedApp.DeviceModel == deviceModel {
							foundIt = true
							trustedApp.RefreshToken = refreshToken
						}
					}
					if !foundIt {
						//create new trusted app
						newTrustedApp := models.TrustedApp{
							Id:           trustedAppId,
							ClientId:     clientId,
							RefreshToken: refreshToken,
							DeviceModel:  deviceModel,
							GrantedAt:    time.Now(),
						}
						//assign this trusted app to user
						u.TrustedApps = append(u.TrustedApps, newTrustedApp)
					}
				}
			}
			var err1, err2 error
			waitGroup := sync.WaitGroup{}
			waitGroup.Add(2)
			go func() {
				defer waitGroup.Done()
				//copy the db session
				session1 := s.Copy()
				defer session1.Close()
				//now should save the user
				err1 = session1.DB(dbName).C(USER_COLLECTION_NAME).UpdateId(u.Id, &u)
			}()
			go func() {
				defer waitGroup.Done()
				//copy the db session
				session2 := s.Copy()
				defer session2.Close()
				//save access token to db
				err2 = session2.DB(dbName).C(ACCESS_TOKEN_COLLECTION_NAME).Insert(&accessToken)
			}()
			waitGroup.Wait()
			if err1 != nil || err2 != nil {
				return nil, specialerror.ErrInternalServerError
			}
			AuthResponse := models.AuthenticationResponse{
				TokenType:    BEARER_AUTHENTICATION_TYPE,
				AccessToken:  accessToken.Token,
				ExpiresInMin: expireIn.Minutes(),
				RefreshToken: refreshToken,
			}
			return &AuthResponse, nil
		}
	}
}
