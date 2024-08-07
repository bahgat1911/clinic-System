package config

import (
	"database/sql"
	"log"
)

func Connect() *sql.DB {
	dbDriver := "mysql"
	dbUser := "admin"
	dbPass := "bahgat"
	dbName := "clinic"

	dsn := dbUser + ":" + dbPass + "@tcp(mydatabase:3306)/" + dbName

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
