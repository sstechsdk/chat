package api

import (
	"github.com/OpenIMSDK/chat/pkg/common/config"
	model "github.com/OpenIMSDK/chat/pkg/common/db/model/oauth"
	"github.com/OpenIMSDK/chat/pkg/common/mctx"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/go-session/session"
	"github.com/google/uuid"
	"gopkg.in/go-oauth2/mongo.v3"
	oredis "gopkg.in/go-oauth2/redis.v3"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
	"log"
	"net/http"
	"net/url"
	"time"
)

func NewOauth2() *Oauth2Api {
	storeType := config.Config.Oauth.TokenStore
	accessTokenExp := time.Duration(config.Config.Oauth.AccessTokenExp)
	refreshTokenExp := time.Duration(config.Config.Oauth.RefreshTokenExp)
	var isGenerateRefresh bool
	if config.Config.Oauth.IsGenerateRefresh == 1 {
		isGenerateRefresh = true
	} else {
		isGenerateRefresh = false
	}
	gManage := manage.NewDefaultManager()
	switch storeType {
	case "mongo":
		gManage.MapTokenStorage(
			mongo.NewTokenStore(mongo.NewConfig(
				"mongodb://"+config.Config.Mongo.Username+":"+config.Config.Mongo.Password+"@"+config.Config.Mongo.Address[0]+"/"+config.Config.Mongo.Database,
				config.Config.Mongo.Database,
			)),
		)
	case "memory":
		gManage.MustTokenStorage(store.NewMemoryTokenStore())
	case "redis":
		gManage.MapTokenStorage(oredis.NewRedisStore(&redis.Options{
			Addr:     (*config.Config.Redis.Address)[0],
			DB:       *config.Config.Redis.DB,
			Password: *config.Config.Redis.Password,
		}))
	case "file":
		gManage.MustTokenStorage(store.NewFileTokenStore("token.db"))
	}

	gClient := store.NewClientStore()
	gManage.MapClientStorage(gClient)
	gServer := server.NewDefaultServer(gManage)
	gServer.SetAllowGetAccessRequest(true)
	gServer.SetClientInfoHandler(server.ClientFormHandler)
	var cfg = &manage.Config{
		AccessTokenExp:    time.Hour * accessTokenExp,
		RefreshTokenExp:   time.Hour * 24 * refreshTokenExp,
		IsGenerateRefresh: isGenerateRefresh,
	}
	gManage.SetAuthorizeCodeTokenCfg(cfg)
	gManage.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)
	gManage.SetClientTokenCfg(cfg)
	gServer.SetExtensionFieldsHandler(func(ti oauth2.TokenInfo) map[string]interface{} {
		data := map[string]interface{}{
			"code":    1,
			"message": "success",
		}
		return data
	})

	gServer.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		opUserID := r.Header.Get("opUserID")
		return opUserID, nil
	})

	gServer.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	gServer.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	return &Oauth2Api{gServer: gServer, gClient: gClient, gManage: gManage}
}

type Oauth2Api struct {
	gServer *server.Server
	gClient *store.ClientStore
	gManage *manage.Manager
}

func (o *Oauth2Api) TokenRequest(c *gin.Context) {
	o.gServer.HandleTokenRequest(c.Writer, c.Request)
}

func (o *Oauth2Api) CodeRequest(c *gin.Context) {
	w := c.Writer
	r := c.Request
	opUserID, _, err := mctx.Check(c)
	if err != nil {
		opUserID = "admin"
		//return
	}
	r.Header.Set("opUserID", opUserID)
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
	}
	r.Form = form

	store.Delete("ReturnUri")
	store.Save()
	err = o.gServer.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (o *Oauth2Api) GetUserInfo(c *gin.Context) {
	// 创建一个空的动态 JSON 对象
	data := make(map[string]interface{})
	data["name"] = "admin"
	data["age"] = 20
	c.JSON(200, data)
}

func (o *Oauth2Api) Credentials(c *gin.Context) {
	clientId := uuid.New().String()[:16]
	clientSecret := uuid.New().String()[:16]
	err := o.gClient.Set(clientId, &models.Client{
		ID:     clientId,
		Secret: clientSecret,
		Domain: "http://localhost:2048",
	})
	if err != nil {
		baseResponse := &model.Base{}
		baseResponse.Code = 1000
		baseResponse.Message = err.Error()
		c.JSON(500, baseResponse)
		c.Abort()
	}
	credentialsResponse := &model.Credential{}
	credentialsResponse.Code = 1
	credentialsResponse.Message = "success"
	credentialsResponse.ClientId = clientId
	credentialsResponse.ClientSecret = clientSecret
	c.JSON(200, credentialsResponse)
}

/*
*
权限验证中间件
*/
func (o *Oauth2Api) AuthValidate(c *gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := o.gServer.ValidationBearerToken(c.Request)
		if err != nil {
			baseResponse := &model.Base{}
			baseResponse.Code = 1001
			baseResponse.Message = err.Error()
			c.JSON(401, baseResponse)
			c.Abort()
			return
		} else {
			c.Next()
		}

	}
}
