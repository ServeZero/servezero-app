package drivers

import (
	"net/http"
	"os"
	"web/config"
	"web/modules/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var store *sessions.FilesystemStore

type fileAuthManager struct {
	name string
}

func init() {
	store = sessions.NewFilesystemStore(os.TempDir(), securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))

	store.Options = &sessions.Options{
		// Domain:   "localhost", // Set this when we have a domain name.
		Path:     "/",
		MaxAge:   3600 * 0.5, // 有効期間30分
		HttpOnly: true,
		// Secure:   true, // Set this when TLS is set up.
	}
}

func NewFileAuthDriver() *fileAuthManager {
	return &fileAuthManager{
		name: config.GetCookieConfig().NAME,
	}
}

func (fileAuth *fileAuthManager) Check(c *gin.Context) bool {
	session, err := store.Get(c.Request, fileAuth.name)
	if err != nil {
		log.Error(err) // securecookie: the value is not valid
		return false
	}
	if session == nil {
		return false
	}
	if session.Values == nil {
		return false
	}
	if session.Values["id"] == nil {
		return false
	}

	// クッキーの有効日時を更新
	session.Save(c.Request, c.Writer)
	return true
}

func (fileAuth *fileAuthManager) User(c *gin.Context) map[string]interface{} {
	session, err := store.Get(c.Request, fileAuth.name)
	if err != nil {
		log.Error(err)
		return nil
	}
	/*if session == nil {
		log.Error("session not found")
		return nil
	}*/

	// マップ文字列型に変換
	user := make(map[string]interface{})
	for key, value := range session.Values {
		switch key := key.(type) {
		case string:
			//switch value := value.(type) {
			//case string:
			user[key] = value
			//}
		}
	}
	return user
}

func (fileAuth *fileAuthManager) Login(http *http.Request, w http.ResponseWriter, user map[string]interface{}) interface{} {
	session, err := store.Get(http, fileAuth.name)
	if err != nil {
		log.Error(err)
		return false
	}
	session.Values["id"] = user["id"]
	_ = session.Save(http, w)
	return true
}

func (fileAuth *fileAuthManager) Logout(http *http.Request, w http.ResponseWriter) bool {
	session, err := store.Get(http, fileAuth.name)
	if err != nil {
		log.Error(err)
		return false
	}

	// セッションクッキーとセッションデータの保存先ファイルを削除
	session.Options.MaxAge = -1
	_ = session.Save(http, w)
	return true
}
