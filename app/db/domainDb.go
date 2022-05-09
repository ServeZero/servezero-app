package db

import (
	database "web/lib/database/sqlite"
)

type DomainDb struct {
	*database.BaseDb
}

// ドメイン一覧取得
func (db *DomainDb) GetDomainList() []map[string]interface{} {
	rows := db.QueryRows(
		`SELECT id, name, app_type, dir_name, hash, db_name, db_user, db_password, created_dt FROM domain ORDER BY id`,
	)
	return rows
}

// IDでドメイン取得
func (db *DomainDb) GetDomain(id int) map[string]interface{} {
	row := db.QueryRow(
		`SELECT id, name, app_type, dir_name, hash, db_name, db_user, db_password, created_dt FROM domain WHERE id = ?`,
		id,
	)
	return row
}

// ドメイン取得
func (db *DomainDb) GetDomainByName(name string) map[string]interface{} {
	row := db.QueryRow(
		`SELECT id, name, app_type, dir_name, hash, db_name, db_user, db_password, created_dt FROM domain WHERE name = ?`,
		name,
	)
	return row
}

// ドメイン追加
// 返り値: 新規ドメインID(1以上=成功、0=失敗)
func (db *DomainDb) AddDomain(name string, dir string, domainHash string) int {
	_, err := db.Exec(
		`INSERT INTO domain (name, dir_name, hash, created_dt) VALUES (?, ?, ?, datetime('now', 'localtime'))`,
		name,
		dir,
		domainHash,
	)
	if err == nil {
		row := db.QueryRow(
			`SELECT id FROM domain WHERE name = ?`,
			name,
		)
		if row != nil {
			return int(row["id"].(int64))
		}
	}
	return 0
}

// ドメイン削除
func (db *DomainDb) DelDomain(id int) bool {
	_, err := db.Exec(
		`DELETE FROM domain WHERE id = ?`,
		id,
	)
	if err == nil {
		return true
	} else {
		return false
	}
}

// Webアプリケーション情報更新
func (db *DomainDb) UpdateAppInfo(appType string, id int, dbName string, user string, password string) bool {
	_, err := db.Exec(
		`UPDATE domain SET app_type = ?, db_name = ?, db_user = ?, db_password = ? WHERE id = ?`,
		appType,
		dbName,
		user,
		password,
		id,
	)
	if err == nil {
		return true
	} else {
		return false
	}
}
