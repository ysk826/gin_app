package config

import (
	"log"

	"gopkg.in/go-ini/ini.v1"
)

type ConfigList struct {
	SQLdriver string
	DbName    string
	DbDir     string
}

// グローバル変数で定義
var Config ConfigList

func init() {
	loadConfig()
}

func loadConfig() {
	cfg, err := ini.Load("config/config.ini")
	if err != nil {
		log.Fatalln("設定ファイルの読み込みに失敗しました:", err)
	}
	Config = ConfigList{
		SQLdriver: cfg.Section("db").Key("driver").String(),
		DbName:    cfg.Section("db").Key("name").String(),
		DbDir:     cfg.Section("db").Key("dir").String(),
	}
}
