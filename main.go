package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/nhoc20170861/simple-bank/api"
	db "github.com/nhoc20170861/simple-bank/db/sqlc"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://postgres:hrqZCS24@com@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = ":4000"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
}
