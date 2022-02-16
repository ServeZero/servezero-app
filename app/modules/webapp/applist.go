package webapp

type AppInfo struct {
	Type            string
	Name            string
	IconName        string
	Description     string
	DownloadUrl     string
	DownloadVersion string
}

var AppInfoList []AppInfo

func init() {
	AppInfoList = []AppInfo{
		{WordPressWebAppType, "WordPress", "wordpress.png", "", "https://ja.wordpress.org/latest-ja.tar.gz", "5.9"},
		{JoomlaWebAppType, "Joomla!", "joomla.png", "", "https://downloads.joomla.org/cms/joomla4/4-1-0/Joomla_4-1-0-Stable-Full_Package.tar.gz", "4.1.0"},
	}
}

// Webアプリケーションタイプが登録済み値かチェック
func CheckAppType(appType string) bool {
	for _, appInfo := range AppInfoList {
		if appInfo.Type == appType {
			return true
		}
	}
	return false
}

// WebアプリケーションのダウンロードURLを取得
func getDownloadUrl(appType string) string {
	for _, appInfo := range AppInfoList {
		if appInfo.Type == appType {
			return appInfo.DownloadUrl
		}
	}
	return ""
}
