package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"

	"github.com/atahani/golang-rest-api-sample/controller/article"
	"github.com/atahani/golang-rest-api-sample/controller/client"
	"github.com/atahani/golang-rest-api-sample/controller/user"
	"github.com/atahani/golang-rest-api-sample/util"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
)

func main() {
	//Echo instance
	app := echo.New()

	//set custom binder to validate payloads
	bi := util.NewCustomBinderWithValidation()
	app.SetBinder(bi)

	//set custom error handler
	app.SetHTTPErrorHandler(specialerror.CustomErrorHandler)

	//set the port listener
	port := "8090"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	//Configs in different app environment mode
	var applicationEnv string
	var mongoDBDialInfo *mgo.DialInfo
	switch os.Getenv("APP_ENV") {
	case "development":
		applicationEnv = "development"
		mongoDBDialInfo = &mgo.DialInfo{
			Addrs:    []string{"localhost:27017"},
			Timeout:  60 * time.Second,
			Database: "golang_sample_dev",
		}
		app.SetDebug(true)
		//Custom logger for console
		app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "${method}-${status} at > ${uri} < in ${response_time} - ${response_size} bytes\n",
		}))
	case "production":
		applicationEnv = "production"
		mongoDBDialInfo = &mgo.DialInfo{
			Addrs:    []string{"localhost"},
			Timeout:  60 * time.Second,
			Database: "golang_sample",
		}
		app.Use(middleware.Recover())
		app.SetDebug(false)
		app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 5,
		}))
	default:
		applicationEnv = "development"
		mongoDBDialInfo = &mgo.DialInfo{
			Addrs:    []string{"localhost:27017"},
			Timeout:  60 * time.Second,
			Database: "golang_sample_dev",
		}
		app.SetDebug(true)
		//Custom logger for console
		app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "${method}-${status} at > ${uri} < in ${response_time} - ${response_size} bytes\n",
		}))
	}

	//create a session with maintains a pool of socket connections to out mongodb
	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		fmt.Printf("connection %s\n", err)
	}
	//check and ensure database indexes
	mongoSession.DB(mongoDBDialInfo.Database).C(user.ACCESS_TOKEN_COLLECTION_NAME).EnsureIndex(mgo.Index{
		Key:         []string{"expire_at"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Second * 1,
	})

	clientController := client.NewClientController(mongoSession, mongoDBDialInfo.Database)
	userController := user.NewUserController(mongoSession, mongoDBDialInfo.Database)
	articleController := article.NewArticleController(mongoSession, mongoDBDialInfo.Database)
	//auth endpoint
	app.Post("/auth/signup", userController.SignUpNewUser)
	app.Post("/auth/singin", userController.SignIn)
	app.Post("/auth/token/refresh", userController.RefreshAccessToken)

	//manage endpoint for client
	apiAdmin := app.Group("/api/manage", user.JWTAuthenticationMiddleware(mongoSession, mongoDBDialInfo.Database), user.AuthorizeUserByRolesMiddleware([]string{"admin"}))
	//manage clients
	apiAdmin.Get("/client", clientController.GetClients)
	apiAdmin.Post("/client", clientController.CreateNewClient)
	apiAdmin.Get("/client/:id", clientController.GetClientById)
	apiAdmin.Put("/client/:id", clientController.UpdateClientById)
	apiAdmin.Delete("/client/:id", clientController.DeleteClientById)

	apiUser := app.Group("/api", user.JWTAuthenticationMiddleware(mongoSession, mongoDBDialInfo.Database), user.AuthorizeUserByRolesMiddleware([]string{"user"}))
	//user profile
	apiUser.Put("/user/profile", userController.UpdateUserProfile)
	apiUser.Put("/user/password", userController.ChangeUserPassword)
	//article
	apiUser.Get("/article", articleController.GetArticlesOfUser)
	apiUser.Post("/article", articleController.CreateArticle)
	apiUser.Get("/article/:id", articleController.GetArticleById)
	apiUser.Put("/article/:id", articleController.UpdateArticleById)
	apiUser.Delete("/article/:id", articleController.DeleteArticleById)

	//start server
	fmt.Printf("API Management Listen to %s port in %s\n", port, applicationEnv)
	app.Run(standard.New(fmt.Sprint(":", port)))
}
