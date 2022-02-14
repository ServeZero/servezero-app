package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"
	"web/config"
	"web/modules/log"

	"github.com/gin-gonic/gin"
)

// システム稼働状況のチェック
func CheckEnv() {
	// 製品のインストールディレクトリが存在するかチェック
	_, err := os.Stat(config.GetEnv().ProductPath)
	if err == nil {
		// Dockerコマンドが実行できるかチェック。インストールはDockerが後になるので注意。
		_, err = exec.Command("docker").Output()
		if err == nil {
			// システムが十分な稼働環境上で動作している
			config.GetEnv().OnProductEnv = true
			log.Info("Processing on PRODUCT environment")
		} else {
			log.Info("Processing on TEST environment. (Docker not running)")
		}
	} else {
		log.Info("Processing on TEST environment.")
	}
}

// Webサービス起動
func Run(router *gin.Engine) {
	log.Info("Start http server listening: " + config.GetEnv().ServerPort)

	srv := &http.Server{
		Addr:    ":" + config.GetEnv().ServerPort,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Info("Server exiting")

	pid := fmt.Sprintf("%d", os.Getpid())
	_, openErr := os.OpenFile("pid", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if openErr == nil {
		//_ = ioutil.WriteFile("pid", []byte(pid), 0)	// ioutilは廃止
		_ = os.WriteFile("pid", []byte(pid), 0)
	}
}
