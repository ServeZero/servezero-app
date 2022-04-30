package config

const PRODUCT_PATH = "/usr/local/servezero" // 製品インストールディレクトリ

var env = Env{
	// アプリケーション
	AppName:         "ServeZero",
	AppFilename:     "servezero",
	AppSecret:       "something-very-secret",
	DefaultLanguage: "ja",

	// システム環境
	ProductPath:    PRODUCT_PATH, // 製品インストールディレクトリ
	ServerPort:     "8080",
	TemplatePath:   "templates", // テンプレートディレクトリ
	DebugOutputDir: "_dest",     // デバッグ用
	OnProductEnv:   false,       // 製品環境で稼働しているかどうか

	// データベース
	DatabaseName: "zero.sqlite3",
	DatabasePath: PRODUCT_PATH + "/config",

	// ログ機能
	AccessLog:          true,
	AccessLogPath:      "log/access.log",
	ErrorLog:           true,
	ErrorLogPath:       "log/error.log",
	SecurityLog:        true,
	SecurityLogPath:    "log/security.log",
	DebugLog:           true,
	DebugLogPath:       "log/debug.log",
	LogTimestampFormat: "2006-01-02 15:04:05",

	// Dockerコンテナ
	DockerContainerNameDb: "db", // コンテナ名(DB)

	// Nginx設定ファイル
	NginxSiteConfPath:         PRODUCT_PATH + "/volumes/nginx/sites-available", // サイト定義ファイル格納用
	NginxSiteConfTemplateDir:  "conf-templates",
	NginxVirtualHostHome:      PRODUCT_PATH + "/volumes/vhost", // Webサイトホームディレクトリ
	NginxVirtualHostHomeAlias: "/var/www/vhost",                // Webサイトホームディレクトリ
	// 設定ファイル定義用
	NginxContainerVirtualHostHome: "/var/www/vhost",

	// PHP設定ファイル
	PhpFpmUser: "www-data", // php-fpmプロセス実行ユーザ

	// MariaDb設定
	MariaDbRootPassword: "root_password", // ルートパスワード
	MariaDbCharacterSet: "utf8mb4",
	MariaDbCollation:    "utf8mb4_general_ci",
}
