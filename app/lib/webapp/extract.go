package webapp

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// アーカイブファイルの解凍
// srcFile: ソースファイルパス
// filename: HTTPレスポンスヘッダから取得した正式なファイル名
// destDir: 解凍先ディレクトリ
// 返り値: 解凍後生成されたディレクトリ, エラーステータス
func extract(srcFile string, filename string, destDir string) (string, error) {
	var err error
	var extractedDir string

	// 拡張子取得
	ext := getExt(filename)

	// アーカイブタイプに応じて解凍
	switch ext {
	case ".tar.gz":
		extractedDir, err = extractTarGz(srcFile, destDir)
	default:
		return "", fmt.Errorf("extract: unsupported archive type: %s", ext)
	}
	return extractedDir, err
}

func getExt(filename string) string {
	base := filepath.Base(filename)
	ext := strings.ToLower(filepath.Ext(base))
	nonExt := filename[:len(filename)-len(ext)]
	if strings.ToLower(filepath.Ext(nonExt)) == ".tar" {
		ext = ".tar" + ext
	}
	return ext
}

func extractTarGz(srcFile string, destDir string) (string, error) {
	gzipStream, err := os.Open(srcFile)
	if err != nil {
		return "", fmt.Errorf("extractTarGz: Open() failed: %s", err.Error())
	}
	defer gzipStream.Close()

	// gzip解凍
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return "", fmt.Errorf("extractTarGz: NewReader() failed: %s", err.Error())
	}

	var isFirst = true
	var extractedDir = ""
	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", fmt.Errorf("extractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(filepath.Join(destDir, header.Name), 0755); err != nil {
				return "", fmt.Errorf("extractTarGz: Mkdir() failed: %s", err.Error())
			}
			if isFirst {
				extractedDir = filepath.Join(destDir, header.Name)
			}
		case tar.TypeReg:
			if isFirst {
				// ルートがディレクトリでない場合はディレクトリを作成
				addDirName := strings.TrimSuffix(filepath.Base(srcFile), ".tar.gz")
				destDir = filepath.Join(destDir, addDirName) // 解凍先ルートディレクトリを変更

				// ルートディレクトリを更新
				if err := os.Mkdir(destDir, 0755); err != nil {
					return "", fmt.Errorf("extractTarGz: Mkdir() failed: %s", err.Error())
				}
				extractedDir = destDir
			}
			outFile, err := os.Create(filepath.Join(destDir, header.Name))
			if err != nil {
				return "", fmt.Errorf("extractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", fmt.Errorf("extractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			return "", fmt.Errorf("extractTarGz: uknown type: %d in %s", header.Typeflag, header.Name)
		}

		isFirst = false
	}
	return extractedDir, nil
}
