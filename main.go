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

	dbConn, err := sql.Open(config.DBDriver, config.DBUrl)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(dbConn)
	server, err := api.NewServer(store, &config)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
