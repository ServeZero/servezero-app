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
		{"wordpress", "WordPress", "wordpress.png", "", "https://ja.wordpress.org/latest-ja.tar.gz", "5.9"},
		{"joomla", "Joomla!", "joomla.png", "", "https://downloads.joomla.org/cms/joomla4/4-1-0/Joomla_4-1-0-Stable-Full_Package.tar.gz", "4.1.0"},
	}
}
