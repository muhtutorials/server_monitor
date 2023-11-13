package handlers

import (
	"fmt"
	"github.com/pusher/pusher-http-go/v5"
	"io"
	"net/http"
	"strconv"
)

func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	userID := repo.App.Session.GetInt(r.Context(), "userID")

	user, _ := repo.DB.GetUserByID(userID)

	params, _ := io.ReadAll(r.Body)

	presenceData := pusher.MemberData{
		UserID: strconv.Itoa(userID),
		UserInfo: map[string]string{
			"id":   strconv.Itoa(userID),
			"name": user.FirstName,
		},
	}

	res, err := app.WS.AuthorizePresenceChannel(params, presenceData)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(res)
}

func (repo *DBRepo) SendPrivateMessage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	message := r.URL.Query().Get("message")

	data := make(map[string]string)
	data["message"] = message

	_ = repo.App.WS.Trigger(fmt.Sprintf("private-channel-%s", id), "private-message", data)
}
