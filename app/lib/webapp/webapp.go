package webapp

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"web/lib/log"
)

const (
	DOWNLOAD_FILE_HEAD = "download-"
	DOWNLOAD_DIR_HEAD  = "download-dir-"
)

const (
	WordPressWebAppType = "wordpress"
	JoomlaWebAppType    = "joomla"
)

type Webapp interface {
	Install(path string) bool
	Backup(path string) bool
}

func NewWebapp(appType string) (Webapp, error) {
	// Webアプリケーションに合わせてインスタンス作成
	switch appType {
	case WordPressWebAppType:
		return &wordpressApp{AppType: WordPressWebAppType}, nil
	case JoomlaWebAppType:
		return &joomlaApp{AppType: JoomlaWebAppType}, nil
	}
	return nil, fmt.Errorf("wrong webapp type: %s", appType)
}

func downloadFile(destPath string, downloadUrl string) (string, error) {
	log.Infof("Download web application package. url: %s", downloadUrl)

	// URLからファイルをダウンロード
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// HTTPレスポンスをチェック
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloadFile: Bad response status: %s", resp.Status)
	}

	// HTTPレスポンスヘッダから正式なファイル名を取得
	var filename string
	filenameCont := resp.Header.Get("Content-Disposition")
	mediaType, params, err := mime.ParseMediaType(filenameCont)
	if err == nil {
		if mediaType != "attachment" {
			return "", fmt.Errorf("downloadFile: ParseMediaType() mediaType not match 'attachment': %s", mediaType)
		}
		filename = params["filename"]
	} else {
		// return "", fmt.Errorf("downloadFile: ParseMediaType() failed: %s", err.Error())
		// MediaTypeが取得できない場合はURLのファイル名を取得
		parsedUrl, err := url.Parse(downloadUrl)
		if err != nil {
			return "", fmt.Errorf("downloadFile: Parse() error: %s", err)
		}
		filename = filepath.Base(parsedUrl.Path)
	}

	// ダウンロードデータをファイルに保存
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return filename, err
}
