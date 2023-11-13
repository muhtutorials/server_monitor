package main

import (
	"flag"
	"fmt"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go/v5"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
	"server_monitor/internal/channels"
	"server_monitor/internal/config"
	"server_monitor/internal/database"
	"server_monitor/internal/handlers"
	"server_monitor/internal/helpers"
	"time"
)

func setupApp() (string, error) {
	insecurePort := flag.String("port", ":8000", "port to listen on")
	identifier := flag.String("identifier", "server_monitor", "unique identifier")
	domain := flag.String("domain", "localhost", "domain name (e.g. example.com)")
	inProduction := flag.Bool("production", false, "application is in production")
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbPort := flag.String("dbport", "5432", "database port")
	dbUser := flag.String("dbuser", "postgres", "database user")
	dbPassword := flag.String("dbpassword", "postgres", "database password")
	dbName := flag.String("dbname", "server_monitor", "database name")
	dbSSL := flag.String("dbssl", "disable", "database SSL setting")
	pusherHost := flag.String("pusherHost", "", "pusher host")
	pusherPort := flag.String("pusherPort", "443", "pusher port")
	pusherAppID := flag.String("pusherAppID", "1690069", "pusher app id")
	pusherKey := flag.String("pusherKey", "ca911d430350756d3260", "pusher key")
	pusherSecret := flag.String("pusherSecret", "c0c4f43474c4696217a8", "pusher secret")
	pusherCluster := flag.String("pusherCluster", "eu", "pusher cluster")
	pusherSecure := flag.Bool("pusherSecure", false, "pusher server uses SSL (true or false)")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app = config.AppConfig{
		InfoLog:      infoLog,
		ErrorLog:     errorLog,
		InProduction: *inProduction,
		Domain:       *domain,
		Identifier:   *identifier,
		Version:      serverMonitorVersion,
	}

	app.InfoLog.Println("Connecting to database...")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
		*dbHost,
		*dbPort,
		*dbUser,
		*dbPassword,
		*dbName,
		*dbSSL,
	)

	db, err := database.ConnectToPostgres(dsn)
	if err != nil {
		app.ErrorLog.Fatal("Cannot connect to database!", err)
	}
	app.DB = db

	app.InfoLog.Println("Initializing session manager...")
	session = scs.New()
	session.Store = postgresstore.New(db.Conn)
	session.Lifetime = 24 * time.Hour
	session.Cookie.Name = fmt.Sprintf("%s_session_id", *identifier)
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = *inProduction
	app.Session = session

	app.InfoLog.Println("Starting email dispatcher...")
	emailQueue := make(chan channels.Email, maxEmailQueueSize)
	app.EmailQueue = emailQueue
	dispatcher := NewDispatcher(emailQueue, maxWorkerPoolSize)
	dispatcher.run()

	repo = handlers.NewPostgresHandlers(&app, db)
	handlers.NewHandlers(&app, repo)

	preferences = make(map[string]string)
	prefs, err := repo.DB.GetAllPreferences()
	if err != nil {
		app.ErrorLog.Fatal("Couldn't read preferences:", err)
	}
	for _, p := range prefs {
		preferences[p.Name] = string(p.Value)
	}

	preferences["pusherHost"] = *pusherHost
	preferences["pusherPort"] = *pusherPort
	preferences["pusherKey"] = *pusherKey
	preferences["identifier"] = *identifier
	preferences["version"] = serverMonitorVersion
	app.Preferences = preferences

	ws = pusher.Client{
		AppID:   *pusherAppID,
		Key:     *pusherKey,
		Secret:  *pusherSecret,
		Cluster: *pusherCluster,
		Secure:  *pusherSecure,
	}
	app.InfoLog.Println("Host", fmt.Sprintf("%s:%s", *pusherHost, *pusherPort))
	app.InfoLog.Println("Secure:", *pusherSecure)

	app.WS = ws

	localZone, _ := time.LoadLocation("Local")
	scheduler := cron.New(cron.WithLocation(localZone), cron.WithChain(
		cron.DelayIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))
	app.Scheduler = scheduler

	app.MonitorEntries = make(map[int]cron.EntryID)

	go handlers.StartMonitoring()

	helpers.NewHelpers(&app)

	return *insecurePort, nil
}
