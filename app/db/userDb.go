package db

import (
	database "web/modules/database/sqlite"
)

type UserDb struct {
	*database.BaseDb
}

// ユーザIDでユーザ情報取得
func (db *UserDb) GetUser(id int) map[string]interface{} {
	row := db.QueryRow(
		`SELECT id, account, password FROM user WHERE id = ?`,
		id,
	)
	return row
}

// サインインアカウントでユーザ情報取得
func (db *UserDb) GetUserByAccount(account string) map[string]interface{} {
	row := db.QueryRow(
		`SELECT id, account, password FROM user WHERE account = ?`,
		account,
	)
	return row
}

// ユーザ情報更新
func (db *UserDb) UpdateUserInfo(account string, password string, id int) bool {
	_, err := db.Exec(
		`UPDATE user SET account = ?, password = ? WHERE id = ?`,
		account,
		password,
		id,
	)
	if err == nil {
		return true
	} else {
		return false
	}
}
