package db_repo

import (
	"database/sql"
	"server_monitor/internal/config"
	"server_monitor/internal/database"
)

var app *config.AppConfig

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(a *config.AppConfig, conn *sql.DB) database.Repo {
	app = a
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}
