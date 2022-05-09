package webapp

import (
	"os"
	"web/config"
	"web/lib/log"
)

type joomlaApp struct {
	Webapp
	AppType string
}

// Webアプリケーションのソースパッケージをダウンロードし、指定のパスに展開
// path=解凍したディレクトリの配置パス
func (joomlaApp *joomlaApp) Install(path string) bool {
	// テンポラリファイル作成
	file, err := os.CreateTemp("", config.GetEnv().AppFilename+"-"+DOWNLOAD_FILE_HEAD)
	if err != nil {
		log.Error(err)
		return false
	}
	tempFile := file.Name()
	file.Close() // 一旦ファイル閉じる
	defer os.Remove(tempFile)
	//defer file.Close()

	// WordPressのソースアーカイブをダウンロード。HTTPリクエストヘッダから正式なファイル名を取得。
	filename, err := downloadFile(tempFile, getDownloadUrl(joomlaApp.AppType))
	if err != nil {
		log.Error(err)
		return false
	}

	// 解凍用の一時ディレクトリ作成
	destDir, err := os.MkdirTemp("", config.GetEnv().AppFilename+"-"+DOWNLOAD_DIR_HEAD)
	if err != nil {
		log.Error(err)
		return false
	}
	defer os.RemoveAll(destDir)

	// ソースを解凍
	extractedDir, err := extract(tempFile, filename, destDir)
	if err != nil {
		log.Error(err)
		return false
	}

	// 指定の位置にディレクトリを移動
	err = os.Rename(extractedDir, path)
	if err == nil {
		log.Infof("Web application installed. type: %s path: %s", WordPressWebAppType, path)
		return true
	} else {
		log.Error(err)
		return false
	}
}
func (joomlaApp *joomlaApp) Backup(path string) bool {
	return true
}
