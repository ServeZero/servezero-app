package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"web/config"
	"web/db"
	"web/modules/filters/auth"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ######################################################################
// ページコントローラ
// ・GETまたはPOSTでHTTPリクエストを受信し、WEB画面を作成して、HTMLデータを返す
// ######################################################################
type LoginController struct{}

func (pc *LoginController) Index(c *gin.Context) {
	// ドメイン一覧表示
	c.HTML(http.StatusOK, "domain.tmpl.html", pongo2.Context{
		"app_name":   config.GetEnv().AppName, // ナビゲーションメニュータイトル
		"page_title": "ドメイン一覧",
	})
}

func (pc *LoginController) Login(c *gin.Context) {
	account := strings.TrimSpace(c.PostForm("account"))
	password := c.PostForm("password")

	// ユーザ情報を取得
	userDb := &db.UserDb{}
	row := userDb.GetUserByAccount(account)
	if row != nil {
		authChecked := false

		// パスワードチェック
		passByte := []byte(fmt.Sprintf("%v", row["password"]))
		err := bcrypt.CompareHashAndPassword(passByte, []byte(password)) // cost=10
		if err == nil {                                                  // パスワードチェックOKの場合
			authChecked = true
		}

		if authChecked {
			// ### ユーザ認証に成功 ###
			// セッションにサインイン情報を登録
			authDriver, _ := c.MustGet(auth.DataTypeUserInfo /*格納データ(ユーザ情報)*/).(auth.Auth)
			authDriver.Login(c.Request, c.Writer, map[string]interface{}{"id": row["id"]})

			c.Redirect(http.StatusFound, "/")
			return
		}
	}

	c.HTML(http.StatusOK, "login.tmpl.html", pongo2.Context{
		"app_name": config.GetEnv().AppName,
		"error":    "サインインに失敗しました",
	})
}

func (pc *LoginController) Logout(c *gin.Context) {
	// セッションからサインイン情報を削除
	authDriver, _ := c.MustGet(auth.DataTypeUserInfo /*格納データ(ユーザ情報)*/).(auth.Auth)
	authDriver.Logout(c.Request, c.Writer)

	c.Redirect(http.StatusFound, "/")
}
