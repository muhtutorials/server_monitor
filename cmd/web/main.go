package main

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go/v5"
	"net/http"
	"runtime"
	"server_monitor/internal/config"
	"server_monitor/internal/handlers"
	"server_monitor/internal/models"
	"time"
)

const (
	serverMonitorVersion = "1.0.0"
	maxEmailQueueSize    = 5
	maxWorkerPoolSize    = 5
)

var (
	app         config.AppConfig
	session     *scs.SessionManager
	repo        *handlers.DBRepo
	preferences map[string]string
	ws          pusher.Client
)

func init() {
	// Behind the scenes SCS uses gob encoding to store session data,
	// so if you want to store custom types in the session data
	// they must be registered with the encoding/gob package first.
	// Struct fields of custom types must also be exported so that they are visible to the encoding/gob package.
	gob.Register(models.User{})
}

func main() {
	insecurePort, err := setupApp()
	if err != nil {
		app.ErrorLog.Fatal(err)
	}

	defer func() {
		close(app.EmailQueue)
		err = app.DB.Conn.Close()
		if err != nil {
			app.ErrorLog.Println(err)
		}
	}()

	app.InfoLog.Printf("******************************************")
	app.InfoLog.Printf("** %sServer monitor%s v%s built in %s", "\033[31m", "\033[0m", serverMonitorVersion, runtime.Version())
	app.InfoLog.Printf("**----------------------------------------")
	app.InfoLog.Printf("** Running with %d processors", runtime.NumCPU())
	app.InfoLog.Printf("** Running on %s", runtime.GOOS)
	app.InfoLog.Printf("******************************************")

	srv := &http.Server{
		Addr:              insecurePort,
		Handler:           routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.InfoLog.Println("Starting HTTP server on port", insecurePort)
	err = srv.ListenAndServe()
	if err != nil {
		app.ErrorLog.Fatal(err)
	}
}
