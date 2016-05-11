package client

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"github.com/atahani/golang-rest-api-sample/models"
	"github.com/labstack/echo"
	"gopkg.in/mcuadros/go-defaults.v1"
	"time"
	"github.com/atahani/golang-rest-api-sample/util"
	"net/http"
	"github.com/atahani/golang-rest-api-sample/util/operationresult"
)

const (
	CLIENT_COLLECTION_NAME = "clients"
	WEB_PLATFORM_TYPE = "web"
)

type ClientController struct {
	Session *mgo.Session
	DBName  string
}

func NewClientController(s *mgo.Session, dbName string) *ClientController {
	return &ClientController{s, dbName}
}

func (cc ClientController) ClientAuthorization(s *mgo.Session, dbName, appId, appKey string) (bool, error) {
	//get copy of db session
	session := s.Copy()
	defer session.Close()
	//check the client id format
	if !bson.IsObjectIdHex(appId) {
		return false, specialerror.ErrNotValidClientInformation
	}
	//find client with this appId
	client := models.Client{}
	if err := session.DB(cc.DBName).C(CLIENT_COLLECTION_NAME).FindId(bson.ObjectIdHex(appId)).One(&client); err != nil {
		if err == mgo.ErrNotFound {
			return false, specialerror.ErrClientIsNotValidToCommunicate
		}
		return false, specialerror.ErrInternalServerError
	}
	//first check is client enable or not
	if !client.IsEnable {
		return false, specialerror.ErrClientIsNotValidToCommunicate
	}
	//if it's not web platform check the AppKey is valid or not
	if client.PlatformType != WEB_PLATFORM_TYPE {
		//check the AppKey is valid or not
		if client.HashedAppKey() != appKey {
			return false, specialerror.ErrClientIsNotValidToCommunicate
		}
		return true, nil
	}
	return true, nil
}

func (cc ClientController) CreateNewClient(c echo.Context) error {
	//get copy of db session
	session := cc.Session.Copy()
	defer session.Close()
	appKey := util.NewAppKey()
	client := models.Client{
		AppId: bson.NewObjectId(),
		AppKey: appKey,
	}
	defaults.SetDefaults(&client)
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()
	//the binder check if struct is not valid return err
	if err := c.Bind(&client); err != nil {
		return err
	}
	//save the client to DB
	if err := session.DB(cc.DBName).C(CLIENT_COLLECTION_NAME).Insert(&client); err != nil {
		return specialerror.ErrInternalServerError
	}
	//replace the hashed App Key
	client.HashedAppKey()
	c.JSON(http.StatusCreated, client)
	return nil
}

func (cc ClientController) UpdateClientById(c echo.Context) error {
	//first check is id valid or not
	if !bson.IsObjectIdHex(c.Param("id")) {
		return specialerror.ErrNotValidItemId
	}
	updatedClient := models.Client{}
	defaults.SetDefaults(&updatedClient)
	//the binder check if struct is not valid return err
	if err := c.Bind(&updatedClient); err != nil {
		return err
	}
	//get copy of session
	session := cc.Session.Copy()
	defer session.Close()
	clientUpdateSet := bson.M{
		"name":updatedClient.Name,
		"description":updatedClient.Description,
		"is_enable":updatedClient.IsEnable,
		"platform_type": updatedClient.PlatformType,
		"updated_at": time.Now(),
	}
	//update the client information by one query
	if err := session.DB(cc.DBName).C(CLIENT_COLLECTION_NAME).UpdateId(bson.ObjectIdHex(c.Param("id")), clientUpdateSet); err != nil {
		if err == mgo.ErrNotFound {
			return specialerror.ErrNotFoundAnyItemWithThisId
		}
		return specialerror.ErrInternalServerError
	}
	//inform the item successfully updated
	c.JSON(http.StatusOK, operationresult.SuccessfullyUpdated)
	return nil
}

func (cc ClientController) GetClientById(c echo.Context) error {
	//first check is id valid or not
	if !bson.IsObjectIdHex(c.Param("id")) {
		return specialerror.ErrNotValidItemId
	}
	//get copy of session
	session := cc.Session.Copy()
	defer session.Close()
	client := models.Client{}
	if err := session.DB(cc.DBName).C(CLIENT_COLLECTION_NAME).FindId(bson.ObjectIdHex(c.Param("id"))).One(&client); err != nil {
		if err == mgo.ErrNotFound {
			return specialerror.ErrNotFoundAnyItemWithThisId
		}
		return specialerror.ErrInternalServerError
	}
	//replace the hashed app key
	client.HashedAppKey()
	c.JSON(http.StatusOK, client)
	return nil
}

func (cc ClientController) GetClients(c echo.Context) error {
	//get copy of session
	session := cc.Session.Copy()
	defer session.Close()
	result := [] models.Client{}
	//get client from database
	if err := session.DB(cc.DBName).C(CLIENT_COLLECTION_NAME).Find(nil).All(&result); err != nil {
		return specialerror.ErrInternalServerError
	}
	//should replace the hashed app key
	for i, cli := range result {
		result[i].AppKey = cli.HashedAppKey()
	}
	c.JSON(http.StatusOK, result)
	return nil
}

func (cc ClientController) DeleteClientById(c echo.Context) error {
	//first check is id valid or not
	if !bson.IsObjectIdHex(c.Param("id")) {
		return specialerror.ErrNotValidItemId
	}
	//get copy of session
	session := cc.Session.Copy()
	defer session.Close()
	if err := session.DB(cc.DBName).C(CLIENT_COLLECTION_NAME).RemoveId(bson.ObjectIdHex(c.Param("id"))); err != nil {
		if err == mgo.ErrNotFound {
			return specialerror.ErrNotFoundAnyItemWithThisId
		}
		return specialerror.ErrInternalServerError
	}
	//inform this item successfully removed
	c.JSON(http.StatusOK, operationresult.SuccessfullyRemoved)
	return nil
}