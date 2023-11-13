package database

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"time"
)

type DB struct {
	Conn *sql.DB
}

var dbConn = &DB{}

const (
	maxOpenDBConns    = 25
	maxIdleDBConns    = 25
	dbConnMaxLifetime = 5 * time.Minute
)

func ConnectToPostgres(dsn string) (*DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	conn.SetMaxOpenConns(maxOpenDBConns)
	conn.SetMaxIdleConns(maxIdleDBConns)
	conn.SetConnMaxLifetime(dbConnMaxLifetime)

	err = conn.Ping()
	if err != nil {
		log.Println("DB error:", err)
	} else {
		log.Println("Successfully connected to DB")
	}

	dbConn.Conn = conn

	return dbConn, err
}
