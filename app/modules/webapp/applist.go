package webapp

type AppInfo struct {
	Type        string
	Name        string
	IconName    string
	Description string
}

var AppInfoList []AppInfo

func init() {
	AppInfoList = []AppInfo{{"wordpress", "WordPress", "", ""}, {"joomla", "Joomla!", "", ""}}
}
