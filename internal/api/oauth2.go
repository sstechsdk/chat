package api

import (
	"fmt"
	"github.com/OpenIMSDK/chat/pkg/common/apicall"
	"github.com/OpenIMSDK/chat/pkg/common/config"
	"github.com/OpenIMSDK/chat/pkg/common/constant"
	model "github.com/OpenIMSDK/chat/pkg/common/db/model/oauth"
	"github.com/OpenIMSDK/chat/pkg/common/mctx"
	"github.com/OpenIMSDK/chat/pkg/proto/chat"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/apiresp"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/models"
	"google.golang.org/grpc"
	"strconv"

	//"github.com/go-oauth2/mongo"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	oredis "github.com/go-oauth2/redis/v4"
	"github.com/go-redis/redis/v8"
	"github.com/go-session/session"
	"log"
	"net/http"
	"net/url"
	"time"
)

func NewOauth2(chatConn grpc.ClientConnInterface) *Oauth2Api {
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
		//gManage.MapTokenStorage(
		//	mongo.NewTokenStore(mongo.NewConfig(
		//		"mongodb://"+config.Config.Mongo.Username+":"+config.Config.Mongo.Password+"@"+config.Config.Mongo.Address[0]+"/"+config.Config.Mongo.Database,
		//		config.Config.Mongo.Database,
		//	)),
		//)
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

	return &Oauth2Api{
		gServer:     gServer,
		gClient:     gClient,
		gManage:     gManage,
		imApiCaller: apicall.NewCallerInterface(),
		chatClient:  chat.NewChatClient(chatConn)}
}

type Oauth2Api struct {
	gServer     *server.Server
	gClient     *store.ClientStore
	gManage     *manage.Manager
	imApiCaller apicall.CallerInterface
	chatClient  chat.ChatClient
}

func (o *Oauth2Api) TokenRequest(c *gin.Context) {
	o.gServer.HandleTokenRequest(c.Writer, c.Request)
}

func (o *Oauth2Api) CodeRequest(c *gin.Context) {
	//http://localhost:10008/auth/code?client_id=1726574631793594368&redirect_uri=&response_type=code&scope=all&state=code
	w := c.Writer
	r := c.Request

	//获取登陆的用户
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
	// 获取 access token
	access_token, ok := o.gServer.BearerAuth(c.Request)
	if !ok {
		log.Println("Failed to get access token from request")
		return
	}

	// 从 access token 中获取 信息
	tokenInfo, err := o.gServer.Manager.LoadAccessToken(c, access_token)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	// 获取当前时间的毫秒级时间戳
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// 将时间戳转换为字符串
	timestampStr := fmt.Sprintf("%d", timestamp)

	// 添加了operationID的處理邏輯
	if operationID, _ := c.Value(constant.RpcOperationID).(string); operationID == "" {
		c.Set(constant.RpcOperationID, timestampStr)
	}
	// 获取 user id
	userId := tokenInfo.GetUserID()
	if opUserID, _ := c.Value(constant.RpcOpUserID).(string); opUserID == "" {
		c.Set(constant.RpcOpUserID, userId)
		c.Set(constant.RpcOpUserType, []string{strconv.Itoa(constant.NormalUser)})
		c.Set(constant.RpcCustomHeader, []string{constant.RpcOpUserType})
	}
	//grant_scope := tokenInfo.GetScope()

	req := chat.SearchUserFullInfoReq{
		Keyword: userId,
		Pagination: &sdkws.RequestPagination{
			PageNumber: 1,
			ShowNumber: 1,
		},
	}

	data, err := chat.ChatClient.SearchUserFullInfo(o.chatClient, c, &req)
	if err != nil {
		apiresp.GinError(c, err) // RPC调用失败
		return
	}
	c.JSON(200, data.Users[0])
	//apiresp.GinSuccess(c, data) // 成功

	// 根据 grant scope 决定获取哪些用户信息
	//if grant_scope != "read_user_info" {
	//	log.Println("invalid grant scope")
	//	w.Write([]byte("invalid grant scope"))
	//	return
	//}
	//
	//user_info = user_info_map[user_id]
	//resp, err := json.Marshal(user_info)
	//w.Write(resp)
	//return
}

func (o *Oauth2Api) Credentials(c *gin.Context) {
	// 获取接口数据，并缓存起来
	r := c.Request
	clientID := r.FormValue("client_id")
	//clientSecret := ""
	client, _ := o.gClient.GetByID(c, clientID)
	if client == nil {
		respRegisterUser, err := o.imApiCaller.GetThirdApp(c, clientID)
		if err != nil {
			apiresp.GinError(c, err)
			return
		}
		//clientSecret = respRegisterUser.AppSecret
		u, err := url.Parse(respRegisterUser.CallbackUrl)
		if err != nil {
			fmt.Println("URL parsing error:", err)
			return
		}

		// 获取 host 和 port
		host := u.Hostname()
		port := u.Port()

		// 组合 host 和 port
		hostWithPort := "http://" + host
		if port != "" {
			hostWithPort += ":" + port
		}

		client = &models.Client{
			ID:               clientID,
			Secret:           respRegisterUser.AppSecret,
			Domain:           hostWithPort,
			CallbackUrl:      respRegisterUser.CallbackUrl,
			ServerPublicKey:  respRegisterUser.ServerPublicKey,
			ServerPrivateKey: respRegisterUser.ServerPrivateKey,
			ClientPublicKey:  respRegisterUser.ClientPublicKey,
			AppName:          respRegisterUser.AppName,
		}
		err = o.gClient.Set(clientID, client)
		if err != nil {
			baseResponse := &model.Base{}
			baseResponse.Code = 1000
			baseResponse.Message = err.Error()
			c.JSON(500, baseResponse)
			c.Abort()
		}
	}
	credentialsResponse := &model.Credential{}
	credentialsResponse.Code = 1
	credentialsResponse.Message = "success"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	appUrl := fmt.Sprint("http://127.0.0.1:10008/auth/code?client_id=", clientID, "&redirect_uri=", client.GetCallbackUrl(), "&response_type=token&scope=all&state=", timestamp)
	//credentialsResponse.ClientSecret = clientSecret

	apiresp.GinSuccess(c, appUrl)
	//c.JSON(200, appUrl)
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
