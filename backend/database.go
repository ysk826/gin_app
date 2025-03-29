package main

import (
	"database/sql"
	"fmt"
	"gin_app/config"

	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var err error

const tableDiary = "diary"

func init() {
	// データベースのディレクトリの作成
	dbDir := config.Config.DbDir
	if err = os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalln("データベースのディレクトリの作成に失敗しました:", err)
	}

	// DB接続
	db, err = sql.Open(config.Config.SQLdriver, config.Config.DbName)
	if err != nil {
		log.Fatalln("DB接続に失敗しました:", err)
	}

	// DB接続のテスト
	if err = db.Ping(); err != nil {
		log.Fatalln("DB接続テストに失敗しました:", err)
	}

	// テーブルの作成
	cmdD := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(

			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time DATETIME,
			milk INTEGER,
			urine INTEGER DEFAULT 0 CHECK (urine IN (0, 1)),
			poop INTEGER DEFAULT 0 CHECK (poop IN (0, 1)),
			created_at DATETIME)`, tableDiary)

	_, err = db.Exec(cmdD)

	if err != nil {
		log.Fatalln("テーブルの作成に失敗しました:", err)
	}

}
