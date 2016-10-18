package article

import (
	"time"
	"net/http"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/labstack/echo"

	"github.com/atahani/golang-rest-api-sample/models"
	"github.com/atahani/golang-rest-api-sample/util/specialerror"
	"github.com/atahani/golang-rest-api-sample/util/operationresult"
	"github.com/atahani/golang-rest-api-sample/controller/user"
)

const (
	ARTICLE_COLLECTION_NAME = "articles"
)

type ArticleController struct {
	Session *mgo.Session
	DBName  string
}

func NewArticleController(s *mgo.Session, dbName string) *ArticleController {
	return &ArticleController{s, dbName}
}

func (ac ArticleController) CreateArticle(c echo.Context) error {
	//get user_id from context
	userId, ok := c.Get(user.USER_ID_KEY).(bson.ObjectId)
	if !ok {
		return specialerror.ErrInternalServerError
	}
	//get article from request
	article := models.Article{
		Id:bson.NewObjectId(),
		UserId:userId,
		CreatedAt:time.Now(),
		UpdatedAt:time.Now(),
	}
	if err := c.Bind(&article); err != nil {
		return err
	}
	//get copy of db session
	session := ac.Session.Copy()
	defer session.Close()
	if err := session.DB(ac.DBName).C(ARTICLE_COLLECTION_NAME).Insert(&article); err != nil {
		return specialerror.ErrInternalServerError
	}
	//return the new article
	c.JSON(http.StatusCreated, article)
	return nil
}

func (ac ArticleController) GetArticleById(c echo.Context) error {
	//first check is id valid or not
	if !bson.IsObjectIdHex(c.Param("id")) {
		return specialerror.ErrNotValidItemId
	}
	//get copy of db session
	session := ac.Session.Copy()
	defer session.Close()
	article := models.Article{}
	if err := session.DB(ac.DBName).C(ARTICLE_COLLECTION_NAME).FindId(bson.ObjectIdHex(c.Param("id"))).One(&article); err != nil {
		if err == mgo.ErrNotFound {
			return specialerror.ErrNotFoundAnyItemWithThisId
		}
		return specialerror.ErrInternalServerError
	}
	//send the article
	c.JSON(http.StatusOK, article)
	return nil
}

func (ac ArticleController) DeleteArticleById(c echo.Context) error {
	//get the userId from context
	userId, ok := c.Get(user.USER_ID_KEY).(bson.ObjectId)
	if !ok {
		return specialerror.ErrInternalServerError
	}
	//first check is id valid or not
	if !bson.IsObjectIdHex(c.Param("id")) {
		return specialerror.ErrNotValidItemId
	}
	//get copy of db session
	session := ac.Session.Copy()
	defer session.Close()
	if err := session.DB(ac.DBName).C(ARTICLE_COLLECTION_NAME).Remove(bson.M{"_id":bson.ObjectIdHex(c.Param("id")), "user_id":userId}); err != nil {
		if err == mgo.ErrNotFound {
			return specialerror.ErrNotFoundAnyItemWithThisId
		}
		return specialerror.ErrInternalServerError
	}
	//inform user that this article removed successfully
	c.JSON(http.StatusOK, operationresult.SuccessfullyRemoved)
	return nil
}

func (ac ArticleController) UpdateArticleById(c echo.Context) error {
	//get the userId from context
	userId, ok := c.Get(user.USER_ID_KEY).(bson.ObjectId)
	if !ok {
		return specialerror.ErrInternalServerError
	}
	//first check is id valid or not
	if !bson.IsObjectIdHex(c.Param("id")) {
		return specialerror.ErrNotValidItemId
	}
	updatedArticle := models.Article{}
	//the binder check if struct is not valid return err
	if err := c.Bind(&updatedArticle); err != nil {
		return err
	}
	articleUpdateSet := bson.M{
		"title": updatedArticle.Title,
		"content":updatedArticle.Content,
		"updated_at":time.Now(),
	}
	//get copy of db session
	session := ac.Session.Copy()
	defer session.Close()
	//NOTE: since we want to update article with one query we don't check is article own by this user separately
	if err := session.DB(ac.DBName).C(ARTICLE_COLLECTION_NAME).Update(bson.M{"_id":bson.ObjectIdHex(c.Param("id")), "user_id":userId}, bson.M{"$set":articleUpdateSet}); err != nil {
		if err == mgo.ErrNotFound {
			return specialerror.ErrCanNotAccessToTheseResource
		}
		return specialerror.ErrInternalServerError
	}
	//inform user this article update successfully
	c.JSON(http.StatusOK, operationresult.SuccessfullyUpdated)
	return nil
}

func (ac ArticleController) GetArticlesOfUser(c echo.Context) error {
	//get copy of db session
	session := ac.Session.Copy()
	defer session.Close()
	//get user_id from Context
	userId, ok := c.Get(user.USER_ID_KEY).(bson.ObjectId);
	if !ok {
		return specialerror.ErrInternalServerError
	}
	result := [] models.Article{}
	if err := session.DB(ac.DBName).C(ARTICLE_COLLECTION_NAME).Find(bson.M{"user_id":userId}).All(&result); err != nil {
		return specialerror.ErrInternalServerError
	}
	//send articles
	c.JSON(http.StatusOK, result)
	return nil
}
