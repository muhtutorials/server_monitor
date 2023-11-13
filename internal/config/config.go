package config

import (
	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go/v5"
	"github.com/robfig/cron/v3"
	"html/template"
	"log"
	"server_monitor/internal/channels"
	"server_monitor/internal/database"
)

type AppConfig struct {
	InfoLog        *log.Logger
	ErrorLog       *log.Logger
	DB             *database.DB
	Session        *scs.SessionManager
	EmailQueue     chan channels.Email
	WS             pusher.Client
	Scheduler      *cron.Cron
	MonitorEntries map[int]cron.EntryID
	Preferences    map[string]string
	TemplateCache  map[string]*template.Template
	InProduction   bool
	Domain         string
	Identifier     string
	Version        string
}
