package main

import (
	"net/http"
	"time"
	"web/config"
	"web/controllers"
	"web/modules/filters/auth"

	//"github.com/gin-contrib/pprof"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"github.com/stnc/pongo2gin"
)

func initRouter() *gin.Engine {
	// システムのデフォルトのミドルウェア(ログ機能、リカバリー機能)は組み込まない
	router := gin.New()

	// テンプレートエンジン設定
	router.HTMLRender = pongo2gin.TemplatePath("templates")
	_setGlobalTemplateData() // グローバルデータ初期化

	// Javascriptファイル、CSSファイル、画像ファイルを公開
	router.Static("/assets", "public/assets")

	// 処理時間計測用
	// if config.GetEnv().Debug {
	// 	pprof.Register(router)
	// }

	// ミドルウェアの設定
	router.Use(gin.Logger())
	//router.Use(gin.Recovery()) // リカバリー機能(gin)
	router.Use(handleErrors()) // リカバリー機能(カスタム)
	router.Use(auth.RegisterGlobalAuthDriver(auth.FileAuthDriverKey /*ドライバータイプ(ファイル保存型セッション)*/, auth.DataTypeUserInfo /*格納データ(ユーザ情報)*/))

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tmpl.html",
			pongo2.Context{
				"language": config.GetEnv().DefaultLanguage,
				"title":    "404 - ページが見つかりません",
				"message":  "ページが見つかりません",
			},
		)
	})

	router.NoMethod(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tmpl.html",
			pongo2.Context{
				"language": config.GetEnv().DefaultLanguage,
				"title":    "404 - ページが見つかりません",
				"message":  "ページが見つかりません",
			},
		)
	})

	// ルーティング設定
	registerApiRouter(router)
	registerPageRouter(router)

	return router
}

// ######################################################################
// Webページ用ルーティング設定
// ・GETまたはPOSTでHTTPリクエストを受信し、WEB画面を作成して、HTMLデータを返す
// ######################################################################
func registerPageRouter(router *gin.Engine) {
	// コントローラ作成
	loginController := &controllers.LoginController{}
	domainController := &controllers.DomainController{}
	accountController := &controllers.AccountController{}

	router.POST("/login", loginController.Login)
	router.GET("/logout", loginController.Logout)

	// その他のページは認証後アクセス可能
	router.Use(auth.Middleware(auth.FileAuthDriverKey))
	{
		router.GET("/", domainController.Index)
		router.POST("/", domainController.Index)
		router.GET("/domain/:id", domainController.Detail)
		router.GET("/account", accountController.Index)
		router.POST("/account", accountController.Index)
	}
}

// ######################################################################
// API用ルーティング設定
// ・JSONデータを受信し、JSONデータを返す
// ######################################################################
func registerApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(auth.Middleware(auth.FileAuthDriverKey))
	{
		controller := &controllers.ApiController{}
		apiRouter.GET("/index", controller.Index)
	}
}

// テンプレート用のグローバルデータ初期化
func _setGlobalTemplateData() {
	now := time.Now()
	pongo2.Globals["g_year"] = now.Year()
}
