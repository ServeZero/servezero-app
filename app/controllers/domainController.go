package controllers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"web/config"
	"web/db"
	"web/modules/log"
	"web/modules/webapp"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"

	"github.com/go-playground/validator/v10"
	"github.com/lithammer/shortuuid"
	"github.com/sethvargo/go-password/password"
)

// ######################################################################
// ドメイン情報コントローラ
// ・ドメイン情報を管理する
// ######################################################################
const (
	SITE_CONF_FILE_FORMAT        = "vhost-%s.conf"        // Nginxサイト定義ファイルフォーマット
	BACKUP_SITE_CONF_FILE_FORMAT = "vhost-%s.conf_backup" // Nginxサイト定義ファイルフォーマット
	SITE_CONF_TEMPLATE           = "site.conf.tmpl"       // Nginxサイト定義ファイルテンプレート
	SITE_CONF_PUBLIC_DIR         = "public_html"
	SITE_HOME_DIR_HEAD           = "vhost-"
	SITE_DB_CREATE_SQL_TEMPLATE  = "createdb.sql.tmpl" // DB、DBユーザ生成用スクリプト
	SITE_DB_DROP_SQL_TEMPLATE    = "dropdb.sql.tmpl"   // DB、DBユーザ破棄用スクリプト
	SITE_DB_NAME_HEAD            = "vhost-"
	MSG_CHANGE_FILENAME          = "ファイル名を変更しました。%s → %s"
	MSG_CREATE_DB                = "データベースを作成しました。DB名=%s, ユーザ=%s"
	MSG_CREATE_DB_FAILED         = "データベース作成に失敗しました。スクリプト=%s"
	MSG_DROP_DB                  = "データベースを削除しました。DB名=%s, ユーザ=%s"
	MSG_DROP_DB_FAILED           = "データベース削除に失敗しました。スクリプト=%s"
	MSG_DELETE_DOMAIN_DIR_FAILED = "ドメインのディレクトリ削除に失敗しました。%s"
)

type DomainController struct{}

type ValidateNewDomain struct {
	Name    string `validate:"required,fqdn"`
	AppType string `validate:"required,apptype"`
}

func (pc *DomainController) Index(c *gin.Context) {
	// パラメータ初期化
	var error, success string // メッセージパラメータ
	domainDb := &db.DomainDb{}

	// 入力値取得
	act := strings.TrimSpace(c.PostForm("act")) // 実行操作

	if act == "add" { // ドメイン追加の場合
		// 入力値取得
		domainName := strings.TrimSpace(c.PostForm("name")) // ドメイン名
		appType := strings.TrimSpace(c.PostForm("apptype")) // Webアプリケーションタイプ

		// 入力値フィールド定義
		newDomain := &ValidateNewDomain{
			Name:    domainName,
			AppType: appType,
		}
		validate := validator.New()

		// Webアプリケーションタイプの値チェック処理追加
		_ = validate.RegisterValidation("apptype", func(fl validator.FieldLevel) bool {
			result := webapp.CheckAppType(fl.Field().String())
			return result
		})

		// 入力値チェック
		err := validate.Struct(newDomain)
		if err != nil {
			// エラーメッセージ設定
			for _, err := range err.(validator.ValidationErrors) {
				fieldName := err.Field() // 入力エラーの構造体変数名を取得

				switch fieldName {
				case "Name":
					switch err.Tag() {
					case "required":
						error = "ドメイン名を入力してください"
					case "fqdn":
						error = "ドメイン名のフォーマットが不正です"
					}
				case "AppType":
					switch err.Tag() {
					case "required":
						error = "Webアプリケーションが選択されていません"
					case "apptype":
						error = "選択できないWebアプリケーションを選択しました"
					}
				}
			}
		}

		if error == "" {
			// ドメイン存在確認
			row := domainDb.GetDomainByName(domainName)
			if row == nil {
				// ドメインハッシュ追加
				domainHash := generateDomainHash(domainName) // ドメインハッシュ作成
				domainId := domainDb.AddDomain(domainName, domainName /*ディレクトリ名*/, domainHash)
				if domainId > 0 { // ドメイン登録成功の場合
					// ########## Webアプリケーションの配置 ##########
					// Webアプリケーションのディレクトリ作成
					siteDirPath := config.GetEnv().NginxVirtualHostHome + "/" + domainName
					err = os.MkdirAll(siteDirPath, 0755)
					if err != nil {
						log.Error(err)
					}

					// Webアプリケーションインストール(公開ディレクトリ)
					webApp, err := webapp.NewWebapp(appType)
					if err == nil {
						installResult := webApp.Install(siteDirPath + "/" + SITE_CONF_PUBLIC_DIR)
						if installResult {
							// Webアプリケーションの公開ディレクトリのオーナーをPHP-FPM用(www-data)に変更
							changePublicDirOwner(domainName)
						}
					} else {
						log.Error(err)
					}

					// ########## DBの作成 ##########
					// パスワード生成
					password, err := password.Generate(8, 8, 0, false, false)
					if err != nil {
						log.Error(err)
					}

					// DB、DBユーザ生成スクリプト実行
					dbName := SITE_DB_NAME_HEAD + domainName
					dbUser := SITE_DB_NAME_HEAD + domainName
					dbResult := createDb(dbName, dbUser, password)
					if dbResult {
						// Webアプリケーション情報を登録
						domainDb.UpdateAppInfo(appType, domainId, dbName, dbUser, password)
					} else {
						// Webアプリケーションタイプのみ登録
						domainDb.UpdateAppInfo(appType, domainId, "", "", "")
					}

					// ########## Nginxの設定 ##########
					// サイト定義ファイル追加
					installSiteConf(domainName, domainHash)

					// サイト定義格納ディレクトリがある場合はNginxを再起動する
					_, err = os.Stat(config.GetEnv().NginxSiteConfPath)
					if err == nil {
						// Nginxサービスに反映
						restartResult := restartNginxService()
						if !restartResult {
							// 再起動不可の場合は定義ファイルを外して再度実行
							recoverSiteConf(domainName)

							restartNginxService()
						}
					}
					success = "ドメインを登録しました"
				} else {
					error = "ドメイン登録に失敗しました"
				}
			} else {
				error = "登録済みのドメインです"
			}
		}
	} else if act == "del" { // ドメイン削除の場合
		// 入力値取得
		domainId, _ := strconv.Atoi(strings.TrimSpace(c.PostForm("id"))) // ドメインID

		// ドメイン情報取得
		row := domainDb.GetDomain(domainId)
		if row == nil {
			error = "ドメイン削除に失敗しました"
		} else {
			domainName := row["name"].(string)

			// Webアプリケーションのディレクトリ削除
			siteDirPath := config.GetEnv().NginxVirtualHostHome + "/" + domainName
			err := os.RemoveAll(siteDirPath)
			if err != nil {
				log.Errorf(MSG_DELETE_DOMAIN_DIR_FAILED, err)
			}

			// DB、DBユーザ破棄スクリプト実行
			dbName := SITE_DB_NAME_HEAD + domainName
			dbUser := SITE_DB_NAME_HEAD + domainName
			dropDb(dbName, dbUser)

			// ドメイン情報削除
			result := domainDb.DelDomain(domainId)
			if result {
				success = fmt.Sprintf("ドメイン「%s」を削除しました", domainName)
			} else {
				error = "ドメイン削除に失敗しました"
			}
		}
	}

	// ドメイン一覧取得
	rows := domainDb.GetDomainList()

	if error == "" { // エラーなしの場合
		// ドメイン一覧表示
		c.HTML(http.StatusOK, "domain.tmpl.html", pongo2.Context{
			"app_name":   config.GetEnv().AppName, // ナビゲーションメニュータイトル
			"page_title": "ドメイン一覧",
			"domainList": rows,
			"appList":    webapp.AppInfoList,
			"success":    success,
		})
	} else {
		c.HTML(http.StatusOK, "domain.tmpl.html", pongo2.Context{
			"app_name":   config.GetEnv().AppName, // ナビゲーションメニュータイトル
			"page_title": "ドメイン一覧",
			"domainList": rows,
			"appList":    webapp.AppInfoList,
			"error":      error,
		})
	}
}

func (pc *DomainController) Detail(c *gin.Context) {
	// パラメータ初期化
	var error string // メッセージパラメータ
	domainDb := &db.DomainDb{}

	domainId, _ := strconv.Atoi(strings.TrimSpace(c.Param("id"))) // ドメインID

	// ドメイン情報取得
	row := domainDb.GetDomain(domainId)
	if row == nil {
		error = "ドメインが見つかりません"
	}

	c.HTML(http.StatusOK, "domain_detail.tmpl.html", pongo2.Context{
		"app_name":    config.GetEnv().AppName, // ナビゲーションメニュータイトル
		"page_title":  "ドメイン詳細",
		"domain_name": row["name"],
		"db_name":     row["db_name"],
		"db_user":     row["db_user"],
		"db_password": row["db_password"],
		"db_host":     config.GetEnv().DockerContainerNameDb,
		"app_type":    row["app_type"],
		"app_dir":     config.GetEnv().NginxVirtualHostHome + "/" + row["dir_name"].(string) + "/" + SITE_CONF_PUBLIC_DIR,
		"created_dt":  row["created_dt"],
		"error":       error,
	})
}

func generateDomainHash(domain string) string {
	domainHash := shortuuid.NewWithNamespace(domain + time.Now().Format("2006-01-02 15:04:05"))
	return domainHash
}

// Nginxのサイト定義ファイルを作成
func installSiteConf(domain string, domainHash string) bool {
	// サイト定義ファイル名生成
	siteConfPath := config.GetEnv().NginxSiteConfPath + "/"
	_, err := os.Stat(siteConfPath)
	if err != nil {
		// デバッグモードの場合はデバッグ用のディレクトリを作成
		if gin.IsDebugging() {
			siteConfPath = "_dest/" + filepath.Base(config.GetEnv().NginxSiteConfPath) + "/"

			// ディレクトリがなければ作成
			os.MkdirAll(siteConfPath, 0755)
		} else {
			log.Error(err)
		}
	}
	siteConfPath += fmt.Sprintf(SITE_CONF_FILE_FORMAT, domain)

	file, err := os.OpenFile(siteConfPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error(err)
		return false
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	siteConfTemplatePath := config.GetEnv().NginxSiteConfTemplateDir + "/" + SITE_CONF_TEMPLATE
	template := pongo2.Must(pongo2.FromFile(siteConfTemplatePath))
	err = template.ExecuteWriter(pongo2.Context{
		"servezero_generated": "ServeZero generate ID: " + domainHash,
		"domain_name":         domain,
		"vhost_path":          config.GetEnv().NginxContainerVirtualHostHome + "/" + domain,
		"public_dir":          SITE_CONF_PUBLIC_DIR,
	}, w)
	if err != nil {
		log.Error(err)
		return false
	}
	w.Flush()
	return true
}

// 問題のあるNginxのサイト定義ファイルを退避
func recoverSiteConf(domain string) {
	// サイト定義ファイルを確認
	siteConfPath := config.GetEnv().NginxSiteConfPath + "/" + fmt.Sprintf(SITE_CONF_FILE_FORMAT, domain)
	_, err := os.Stat(siteConfPath)
	if err == nil {
		// ファイル名変更
		backupConfPath := config.GetEnv().NginxSiteConfPath + "/" + fmt.Sprintf(BACKUP_SITE_CONF_FILE_FORMAT, domain)
		err = os.Rename(siteConfPath, backupConfPath)
		if err == nil {
			log.Info(fmt.Sprintf(MSG_CHANGE_FILENAME, filepath.Base(siteConfPath), filepath.Base(backupConfPath)))
		} else {
			log.Error(err)
		}
	}
}

// Nginxにサイト定義を再読み込みさせる
func restartNginxService() bool {
	_, err := exec.Command("docker", "exec", "nginx", "nginx", "-t").Output()
	if err == nil { // テストOKの場合は設定を再読み込み
		exec.Command("docker", "exec", "nginx", "nginx", "-s", "reload").Output()
		return true
	} else {
		log.Error(err)
		return false
	}
}

// ディレクトリのオーナーを変更
func changePublicDirOwner(domain string) bool {
	// Webアプリケーションの公開ディレクトリ
	appPublicDir := config.GetEnv().NginxContainerVirtualHostHome + "/" + domain + "/" + SITE_CONF_PUBLIC_DIR

	// Dockerコマンドが実行できる場合のみ実行
	if !config.GetEnv().OnProductEnv {
		log.Infof("[TEST-env] Public directory owner not changed. path: %s", appPublicDir)
		return false
	}

	// PHP-FPMのプロセスに書き込み権限を与える
	//_, err := exec.Command("docker", "exec", "nginx", "chown", "-R", config.GetEnv().NginxUser+":"+config.GetEnv().NginxUser, appPublicDir).Output()
	_, err := exec.Command("docker", "exec", "php", "chown", "-R", config.GetEnv().PhpFpmUser+":"+config.GetEnv().PhpFpmUser, appPublicDir).Output()
	if err == nil { // テストOKの場合は設定を再読み込み
		return true
	} else {
		log.Error(err)
		return false
	}
}

// DB、DBユーザを生成
func createDb(dbname string, user string, password string) bool {
	// Dockerコマンドが実行できる場合のみ実行
	if !config.GetEnv().OnProductEnv {
		log.Infof("[TEST-env] DB not created.")
		return false
	}

	// DB作成スクリプトを読み込む
	createDbTemplatePath := config.GetEnv().NginxSiteConfTemplateDir + "/" + SITE_DB_CREATE_SQL_TEMPLATE
	template := pongo2.Must(pongo2.FromFile(createDbTemplatePath))
	createDbScript, err := template.Execute(pongo2.Context{
		"db_name":          dbname,
		"db_user":          user,
		"db_password":      password,
		"db_character_set": config.GetEnv().MariaDbCharacterSet,
		"db_collation":     config.GetEnv().MariaDbCollation,
	})
	if err != nil {
		log.Error(err)
		return false
	}

	// DB、DBユーザを作成
	_, err = exec.Command("docker", "exec", "db", "mysql", "-u", "root", "-p"+config.GetEnv().MariaDbRootPassword, "-e", createDbScript).Output()
	if err == nil {
		log.Info(fmt.Sprintf(MSG_CREATE_DB, dbname, user))
		return true
	} else {
		log.Errorf(MSG_CREATE_DB_FAILED, strings.Replace(createDbScript, "\n", "", -1))
		return false
	}
}

// DB、DBユーザを破棄
func dropDb(dbname string, user string) bool {
	// Dockerコマンドが実行できる場合のみ実行
	if !config.GetEnv().OnProductEnv {
		log.Infof("[TEST-env] DB not dropped.")
		return false
	}

	// DB破棄スクリプトを読み込む
	createDbTemplatePath := config.GetEnv().NginxSiteConfTemplateDir + "/" + SITE_DB_DROP_SQL_TEMPLATE
	template := pongo2.Must(pongo2.FromFile(createDbTemplatePath))
	createDbScript, err := template.Execute(pongo2.Context{
		"db_name": dbname,
		"db_user": user,
	})
	if err != nil {
		log.Error(err)
		return false
	}

	// DB、DBユーザを作成
	_, err = exec.Command("docker", "exec", "db", "mysql", "-u", "root", "-p"+config.GetEnv().MariaDbRootPassword, "-e", createDbScript).Output()
	if err == nil {
		log.Info(fmt.Sprintf(MSG_DROP_DB, dbname, user))
		return true
	} else {
		log.Errorf(MSG_DROP_DB_FAILED, strings.Replace(createDbScript, "\n", "", -1))
		return false
	}
}
