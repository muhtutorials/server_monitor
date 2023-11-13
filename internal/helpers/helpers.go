package helpers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"server_monitor/internal/channels"
	"server_monitor/internal/config"
	"time"
)

var app *config.AppConfig
var src = rand.NewSource(time.Now().UnixNano())

const (
	letterBytes     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIndexBits = 6                      // 6 bits to represent a letter index
	letterIndexMask = 1<<letterIndexBits - 1 // all 1-bits, as many as letterIdxBits
	letterIndexMax  = 63 / letterIndexBits   // of letter indices fitting in 63 bits
)

func NewHelpers(a *config.AppConfig) {
	app = a
}

func RandomString(length int) string {
	b := make([]byte, length)

	for i, cache, remain := length-1, src.Int63(), letterIndexMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIndexMax
		}
		if index := int(cache & letterIndexMask); index < len(letterBytes) {
			b[i] = letterBytes[index]
			i--
		}
		cache >>= letterIndexBits
		remain--
	}

	return string(b)
}

func CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, mode)
		if err != nil {
			app.ErrorLog.Println(err)
			return err
		}
	}
	return nil
}

func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "userID")
	return exists
}

func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = log.Output(2, trace)

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	http.ServeFile(w, r, "./static/500.html")
}

func SendEmail(message channels.Email) {
	if message.FromAddress == "" {
		message.FromAddress = app.Preferences["smtp_from_email"]
		message.FromName = app.Preferences["smtp_from_name"]
	}

	app.EmailQueue <- message
}

func BroadcastMessage(channel, eventName string, data map[string]string) {
	err := app.WS.Trigger(channel, eventName, data)
	if err != nil {
		app.ErrorLog.Println(err)
	}
}
