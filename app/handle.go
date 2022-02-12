package main

import (
	"fmt"
	"net/http"
	"runtime"
	"web/modules/log"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
)

func handleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// ランタイムエラーメッセージ出力
				var callStack string
				for depth := 0; ; depth++ {
					pc, src, line, ok := runtime.Caller(depth)
					if !ok {
						break
					}
					callStack += fmt.Sprintf(" -> %d: %s: %s(%d)\n", depth, runtime.FuncForPC(pc).Name(), src, line)
				}
				log.Errorf("runtime error: %s:\n%s", err, callStack)

				// エラー画面表示
				c.HTML(http.StatusInternalServerError, "500.tmpl.html", pongo2.Context{})
				// var (
				// 	errMsg     string
				// 	mysqlError *mysql.MySQLError
				// 	ok         bool
				// )
				// if errMsg, ok = err.(string); ok {
				// 	c.JSON(http.StatusInternalServerError, pongo2.Context{
				// 		"code": 500,
				// 		"msg":  "system error, " + errMsg,
				// 	})
				// 	return
				// } else if mysqlError, ok = err.(*mysql.MySQLError); ok {
				// 	c.JSON(http.StatusInternalServerError, pongo2.Context{
				// 		"code": 500,
				// 		"msg":  "system error, " + mysqlError.Error(),
				// 	})
				// 	return
				// } else {
				// 	c.JSON(http.StatusInternalServerError, pongo2.Context{
				// 		"code": 500,
				// 		"msg":  "system error",
				// 	})
				// 	return
				// }
			}
		}()
		c.Next()
	}
}
