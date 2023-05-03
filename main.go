package main

import (
	"database/sql"
	"github.com/gaggudeep/bank_go/api"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/util"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	dbConn, err := sql.Open(config.DBConfig.Driver, config.DBConfig.URL)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(dbConn)
	server := api.NewServer(store)

	err = server.Start(config.ServerConfig.Addr)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
