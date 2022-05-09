package auth

import (
	"net/http"
	"web/config"
	"web/lib/filters/auth/drivers"
	"web/lib/log"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
)

// 認証ドライバータイプ
const (
	FileAuthDriverKey = "file"
)

// 格納データタイプ
const (
	DataTypeUserInfo = "userinfo" // ユーザ情報
)

var driverList = map[string]Auth{
	FileAuthDriverKey: drivers.NewFileAuthDriver(),
}

type Auth interface {
	Check(c *gin.Context) bool
	User(c *gin.Context) map[string]interface{}
	Login(http *http.Request, w http.ResponseWriter, user map[string]interface{}) interface{}
	Logout(http *http.Request, w http.ResponseWriter) bool
}

func RegisterGlobalAuthDriver(authKey string, key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(key, _getAuthDriver(authKey))
		c.Next()
	}
}

func Middleware(authKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !_getAuthDriver(authKey).Check(c) {
			c.HTML(http.StatusOK, "login.tmpl.html", pongo2.Context{
				"app_name": config.GetEnv().AppName,
			})
			c.Abort()
		}
		c.Next()
	}
}

func GetDefaultUser(c *gin.Context) map[string]interface{} {
	return GetUser(c, DataTypeUserInfo)
}

func GetUser(c *gin.Context, key string) map[string]interface{} {
	authDriver, _ := c.MustGet(key).(Auth)
	return authDriver.User(c)
}

func _getAuthDriver(string string) Auth {
	auth := driverList[string]
	if auth == nil {
		log.Errorf("Auth driver not found: %v", string)
	}
	return auth
}
