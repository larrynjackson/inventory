package main

import (
	"fmt"
	"github/lnj/inventory/app"
	"github/lnj/inventory/data"
	"github/lnj/inventory/sql"

	"github.com/rs/zerolog/log"
)

type Config struct {
	ListenIp   string `default:"0.0.0.0"`
	ListenPort string `default:"8080"`
	SqlitePath string `default:"./sqlitedb"`
}

var (
	conf Config
)

func main() {
	fmt.Println("Welcom to Inventory")

	conf.ListenIp = "localhost"
	conf.ListenPort = "8000"
	conf.SqlitePath = "./sqlite.db"

	storage, err := setupStorage()
	if err != nil {
		log.Panic().Msgf("error setting up storage - '%s'", err.Error())
	}
	app := app.NewServer(storage, conf.ListenIp+":"+conf.ListenPort)
	app.Run()

}

func setupStorage() (data.StorageReadWrite, error) {
	var store data.StorageReadWrite

	log.Info().Msgf("conf.SqlitePath: '%s'", conf.SqlitePath)
	sqliteStore := sql.NewStoreSQLite(conf.SqlitePath)
	sqliteStore.Open()

	store = sqliteStore

	fmt.Println("store set to:", store)
	return store, nil
}
