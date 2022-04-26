package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"web/config"
	"web/modules/fileutil"
	"web/modules/log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type BaseDb struct {
	*sqlx.DB
}

// DB接続のインスタンス
var sqlxDb *sqlx.DB

func init() {
	// インストール済みDBファイルの存在確認
	dbPath := config.GetEnv().DatabasePath + "/" + config.GetEnv().DatabaseName
	_, err := os.Stat(dbPath)
	if err != nil && gin.IsDebugging() {
		// インストール済みのDBファイルがない場合はローカルのDBに接続(デバッグモード起動時のみ)
		dbPath = "_" + config.GetEnv().DatabaseName
		_, err := os.Stat(dbPath)
		if err != nil { // ファイルがない場合はコピー
			fileutil.CopyFile("install/init.sqlite3", dbPath)
		}
	}

	// DBコネクション取得
	sqlxDb, err = sqlx.Connect("sqlite3", dbPath)
	checkErr(err, dbPath)

	// DB接続メッセージ出力
	path, _ := filepath.Abs(dbPath)
	log.Info("DB connected: " + path)
}
func checkErr(err error, path string) {
	if err != nil {
		// 異常時は終了
		absPath, _ := filepath.Abs(path)
		log.Error("error path: " + absPath)
		log.FatalStack(err) // スタックトレースも出力
	}
}

func (db *BaseDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := sqlxDb.Query(query, args...)

	return rows, err
}

func (db *BaseDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	rows, err := sqlxDb.Exec(query, args...)

	return rows, err
}

// クエリーを実行し複数のレコードを取得
// 連想配列の配列でデータを取得する。エラーの場合はnilが返る。
func (db *BaseDb) QueryRows(query string, args ...interface{}) []map[string]interface{} {
	rs, err := sqlxDb.Query(query, args...)
	if err != nil {
		log.ErrorStack(err)
		return nil
	}

	defer rs.Close()

	col, colErr := rs.Columns()
	if colErr != nil {
		log.ErrorStack(colErr)
		return nil
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		log.ErrorStack(err)
		return nil
	}

	results := make([]map[string]interface{}, 0)

	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			setColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}

		if scanErr := rs.Scan(colVar...); scanErr != nil {
			log.ErrorStack(err)
			return nil
		}

		result := make(map[string]interface{})
		for j := 0; j < len(col); j++ {
			setResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		log.ErrorStack(err)
		return nil
	}

	return results
}

// クエリーを実行し1行レコードを取得
// 連想配列でデータを取得する。エラーあるいはデータなしの場合はnil(Mapの空の値)が返る。
func (db *BaseDb) QueryRow(query string, args ...interface{}) map[string]interface{} {
	rs, err := sqlxDb.Query(query, args...)
	if err != nil {
		log.ErrorStack(err)
		return nil
	}

	defer rs.Close()

	col, colErr := rs.Columns()
	if colErr != nil {
		log.ErrorStack(colErr)
		return nil
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		log.ErrorStack(err)
		return nil
	}

	var result map[string]interface{}

	if rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			setColVarType(&colVar, i, typeVal[i].DatabaseTypeName())
		}

		if scanErr := rs.Scan(colVar...); scanErr != nil {
			log.ErrorStack(err)
			return nil
		}

		result = make(map[string]interface{})
		for j := 0; j < len(col); j++ {
			setResultValue(&result, col[j], colVar[j], typeVal[j].DatabaseTypeName())
		}
	}
	if err := rs.Err(); err != nil {
		log.ErrorStack(err)
		return nil
	}

	return result
}

func setColVarType(colVar *[]interface{}, i int, typeName string) {
	switch typeName {
	case "INTEGER":
		var s sql.NullInt64
		(*colVar)[i] = &s
	case "REAL":
		var s sql.NullFloat64
		(*colVar)[i] = &s
	case "TEXT":
		var s sql.NullString
		(*colVar)[i] = &s
	case "BLOB":
		var s sql.NullString
		(*colVar)[i] = &s
	default:
		var s interface{}
		(*colVar)[i] = &s
	}
}

func setResultValue(result *map[string]interface{}, index string, colVar interface{}, typeName string) {
	switch typeName {
	case "INTEGER":
		temp := *(colVar.(*sql.NullInt64))
		if temp.Valid {
			(*result)[index] = temp.Int64
		} else {
			(*result)[index] = nil
		}
	case "REAL":
		temp := *(colVar.(*sql.NullFloat64))
		if temp.Valid {
			(*result)[index] = temp.Float64
		} else {
			(*result)[index] = nil
		}
	case "TEXT":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	case "BLOB":
		temp := *(colVar.(*sql.NullString))
		if temp.Valid {
			(*result)[index] = temp.String
		} else {
			(*result)[index] = nil
		}
	default:
		(*result)[index] = colVar
	}
}
