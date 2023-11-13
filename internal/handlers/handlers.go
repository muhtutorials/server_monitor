package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/go-chi/chi/v5"
	"net/http"
	"server_monitor/internal/config"
	"server_monitor/internal/database"
	"server_monitor/internal/database/db_repo"
	"server_monitor/internal/helpers"
	"server_monitor/internal/models"
	"strconv"
)

var (
	app  *config.AppConfig
	Repo *DBRepo
)

type DBRepo struct {
	App *config.AppConfig
	DB  database.Repo
}

func NewPostgresHandlers(app *config.AppConfig, db *database.DB) *DBRepo {
	return &DBRepo{
		App: app,
		DB:  db_repo.NewPostgresRepo(app, db.Conn),
	}
}

func NewHandlers(a *config.AppConfig, r *DBRepo) {
	app = a
	Repo = r
}

func (repo *DBRepo) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	healthy, warning, problem, pending, err := repo.DB.GetServiceStatusCounts()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("healthy", healthy)
	vars.Set("warning", warning)
	vars.Set("problem", problem)
	vars.Set("pending", pending)

	hosts, err := repo.DB.GetAllHosts()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars.Set("hosts", hosts)

	err = helpers.RenderPage(w, r, "dashboard", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) Events(w http.ResponseWriter, r *http.Request) {
	events, err := repo.DB.GetAllEvents()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("events", events)

	err = helpers.RenderPage(w, r, "events", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) Settings(w http.ResponseWriter, r *http.Request) {
	err := helpers.RenderPage(w, r, "settings", nil, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) SaveSettings(w http.ResponseWriter, r *http.Request) {
	preferences := make(map[string]string)
	preferences["site_url"] = r.Form.Get("site_url")
	preferences["notify_name"] = r.Form.Get("notify_name")
	preferences["notify_email"] = r.Form.Get("notify_email")
	preferences["smtp_server"] = r.Form.Get("smtp_server")
	preferences["smtp_port"] = r.Form.Get("smtp_port")
	preferences["smtp_user"] = r.Form.Get("smtp_user")
	preferences["smtp_password"] = r.Form.Get("smtp_password")
	preferences["sms_enabled"] = r.Form.Get("sms_enabled")
	preferences["sms_provider"] = r.Form.Get("sms_provider")
	preferences["twilio_phone_number"] = r.Form.Get("twilio_phone_number")
	preferences["twilio_sid"] = r.Form.Get("twilio_sid")
	preferences["twilio_auth_token"] = r.Form.Get("twilio_auth_token")
	preferences["smtp_from_email"] = r.Form.Get("smtp_from_email")
	preferences["smtp_from_name"] = r.Form.Get("smtp_from_name")
	preferences["notify_via_sms"] = r.Form.Get("notify_via_sms")
	preferences["notify_via_email"] = r.Form.Get("notify_via_email")
	preferences["sms_notify_number"] = r.Form.Get("sms_notify_number")

	if r.Form.Get("sms_enabled") == "0" {
		preferences["notify_via_sms"] = "0"
	}

	err := repo.DB.CreateOrUpdatePreferences(preferences)
	if err != nil {
		app.ErrorLog.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	// update app config
	for k, v := range preferences {
		app.Preferences[k] = v
	}

	app.Session.Put(r.Context(), "flash", "Changes saved")

	if r.Form.Get("action") == "1" {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
	}
}

func (repo *DBRepo) UserList(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)

	users, err := repo.DB.GetAllUsers()
	if err != nil {
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	vars.Set("users", users)

	err = helpers.RenderPage(w, r, "user_list", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) User(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var user models.User
	vars := make(jet.VarMap)

	if id > 0 {
		user, err = repo.DB.GetUserByID(id)
		if err != nil {
			ClientError(w, r, http.StatusBadRequest)
			return
		}
	}
	vars.Set("user", user)

	err = helpers.RenderPage(w, r, "user", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) CreateOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var user models.User

	if id > 0 {
		user, err = repo.DB.GetUserByID(id)
		user.FirstName = r.Form.Get("first_name")
		user.LastName = r.Form.Get("last_name")
		user.Email = r.Form.Get("email")
		user.IsActive, _ = strconv.Atoi(r.Form.Get("is_active"))
		err = repo.DB.UpdateUser(user)
		if err != nil {
			app.ErrorLog.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}

		if len(r.Form.Get("password")) > 0 {
			err = repo.DB.UpdatePassword(id, r.Form.Get("password"))
			if err != nil {
				app.ErrorLog.Println(err)
				ClientError(w, r, http.StatusBadRequest)
				return
			}
		}
	} else {
		user.FirstName = r.Form.Get("first_name")
		user.LastName = r.Form.Get("last_name")
		user.Email = r.Form.Get("email")
		user.Password = []byte(r.Form.Get("password"))
		user.IsActive, _ = strconv.Atoi(r.Form.Get("is_active"))
		user.AccessLevel = 3

		_, err = repo.DB.InsertUser(user)
		if err != nil {
			app.ErrorLog.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
	}

	repo.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (repo *DBRepo) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.ErrorLog.Println(err)
	}

	err = repo.DB.DeleteUser(id)
	if err != nil {
		app.ErrorLog.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	repo.App.Session.Put(r.Context(), "flash", "User deleted")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (repo *DBRepo) HostList(w http.ResponseWriter, r *http.Request) {
	hosts, err := repo.DB.GetAllHosts()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("hosts", hosts)

	err = helpers.RenderPage(w, r, "host_list", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) Host(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var host models.Host

	if id > 0 {
		host, err = repo.DB.GetHostByID(id)
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}
	}

	vars := make(jet.VarMap)
	vars.Set("host", host)

	err = helpers.RenderPage(w, r, "host", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) CreateOrUpdateHost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var host models.Host

	if id > 0 {
		host, err = repo.DB.GetHostByID(id)
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}
	}

	host.Name = r.Form.Get("name")
	host.FullName = r.Form.Get("full_name")
	host.URL = r.Form.Get("url")
	host.IP = r.Form.Get("ip")
	host.IPV6 = r.Form.Get("ipv6")
	host.Location = r.Form.Get("location")
	host.OS = r.Form.Get("os")
	host.IsActive, _ = strconv.Atoi(r.Form.Get("is_active"))

	if id > 0 {
		err = repo.DB.UpdateHost(host)
		if err != nil {
			app.ErrorLog.Println(err)
			ServerError(w, r, err)
			return
		}
	} else {
		newID, err := repo.DB.InsertHost(host)
		if err != nil {
			app.ErrorLog.Println(err)
			ServerError(w, r, err)
			return
		}
		host.ID = newID
	}

	repo.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/hosts/%d", host.ID), http.StatusSeeOther)
}

type toggleResponse struct {
	OK bool `json:"ok"`
}

func (repo *DBRepo) ToggleHostIsActive(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var res toggleResponse
	res.OK = true

	hostID, _ := strconv.Atoi(r.Form.Get("host_id"))
	isActive, _ := strconv.Atoi(r.Form.Get("is_active"))

	err = repo.DB.UpdateHostIsActive(hostID, isActive)
	if err != nil {
		app.ErrorLog.Println(err)
		res.OK = false
	}

	services, err := repo.DB.GetHostServices(hostID)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	for _, service := range services {
		if isActive == 1 {
			repo.pushStatusChanged(service, "pending")
			repo.pushScheduleChanged(service, "pending")
			repo.addToMonitorEntries(service)
		} else {
			repo.removeFromMonitorEntries(service)
		}
	}

	out, _ := json.MarshalIndent(res, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (repo *DBRepo) ToggleServiceIsActive(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	var res toggleResponse
	res.OK = true

	serviceID, _ := strconv.Atoi(r.Form.Get("service_id"))
	hostID, _ := strconv.Atoi(r.Form.Get("host_id"))
	isActive, _ := strconv.Atoi(r.Form.Get("is_active"))

	err = repo.DB.UpdateServiceIsActive(serviceID, hostID, isActive)
	if err != nil {
		app.ErrorLog.Println(err)
		res.OK = false
	}

	service, err := repo.DB.GetServiceByID(serviceID)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	if isActive == 1 {
		repo.pushStatusChanged(service, "pending")
		repo.pushScheduleChanged(service, "pending")
		repo.addToMonitorEntries(service)
	} else {
		repo.removeFromMonitorEntries(service)
	}

	out, _ := json.MarshalIndent(res, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (repo *DBRepo) ToggleMonitoring(w http.ResponseWriter, r *http.Request) {
	name := r.PostForm.Get("name")
	value := r.PostForm.Get("value")

	repo.App.Preferences["monitoring_live"] = value

	if value == "1" {
		StartMonitoring()
	} else {
		StopMonitoring()
	}

	var res jsonResponse
	res.OK = true

	err := repo.DB.CreateOrUpdatePreferences(map[string]string{name: value})
	if err != nil {
		app.ErrorLog.Println(err)
		res.OK = false
	}

	out, _ := json.MarshalIndent(res, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
