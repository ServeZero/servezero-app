# 概要

ServeZeroの管理用Webアプリケーションを VSCode + Docker + Remote Containers で開発するためのプロジェクトです。

## 動作環境

- Windows10
- Docker Desktop
- VSCode(拡張機能: Go, Remote Containers)

## 使い方

1. Docker Desktopを起動します。完全に起動するまで待ちます。

2. トップディレクトリでVSCodeを起動します。

```
> code .
```

3. 起動直後に以下のメッセージのダイアログが表示されます。「Reopen in Container」ボタンでDockerコンテナを起動します。

```
Folder contains a Dev Container configuration file. Reopen folder to develop in a container (learn more)
```

4. Dockerコンテナ環境が起動すると、VSCodeがリモート接続して、Webサーバ開発ディレクトリ(app)の表示に切り替わります。

5. リモートエクスプローラー画面の「CONTAINERS」から起動したDockerコンテナ環境を選択し、右クックで「Show container Log」を選択すると、
アプリケーションの起動状況がログが表示されます。

6. Webブラウザからアクセスします。

```
http://localhost:8080
```
