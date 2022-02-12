package fileutil

import (
	"os"
	"web/modules/log"
)

func CopyFile(src string, dest string) bool {
	//bytesRead, err := ioutil.ReadFile(src)	// ioutilは廃止
	bytesRead, err := os.ReadFile(src)
	if err != nil {
		log.Error(err)
		return false
	}

	//err = ioutil.WriteFile(dest, bytesRead, 0644)// ioutilは廃止
	err = os.WriteFile(dest, bytesRead, 0644)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}
