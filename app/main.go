// コマンド起動で最初に呼ばれるmain処理です。
// リリース時は環境変数GIN_MODEにreleaseを設定してリリースモードでコマンド実行します。リリースモードではデバッグ用のコンソール出力が抑止されます。
// export GIN_MODE=release
package main

import (
	"runtime"
	_ "web/modules/database/sqlite" // DB初期処理
	"web/modules/server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// システム稼働状況のチェック
	server.CheckEnv()

	// ルーティング設定
	router := initRouter()

	// Webサーバ起動
	server.Run(router)
}
