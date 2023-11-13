package handlers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
	"server_monitor/internal/channels"
	"server_monitor/internal/helpers"
	"server_monitor/internal/models"
	"server_monitor/internal/sms"
	"strconv"
	"strings"
	"time"
)

const (
	HTTP           = "HTTP"
	HTTPS          = "HTTPS"
	SSLCertificate = "SSL"
)

type jsonResponse struct {
	OK        bool      `json:"ok"`
	Message   string    `json:"message"`
	HostID    int       `json:"host_id"`
	ServiceID int       `json:"service_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	LastCheck time.Time `json:"last_check"`
}

func (repo *DBRepo) StartScheduledCheck(serviceID int) {
	app.InfoLog.Println(">>>>> Running check for service:", serviceID)

	service, err := repo.DB.GetServiceByID(serviceID)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	host, err := repo.DB.GetHostByID(service.HostID)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	newStatus, _ := repo.checkService(host, service)

	if newStatus != service.Status {
		repo.updateServiceStatusCount(service, newStatus)
	}
}

func (repo *DBRepo) updateServiceStatusCount(service models.Service, newStatus string) {
	service.Status = newStatus
	service.LastCheck = time.Now()

	err := repo.DB.UpdateService(service)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	healthy, warning, problem, pending, err := repo.DB.GetServiceStatusCounts()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	data := make(map[string]string)
	data["healthyCount"] = strconv.Itoa(healthy)
	data["warningCount"] = strconv.Itoa(warning)
	data["problemCount"] = strconv.Itoa(problem)
	data["pendingCount"] = strconv.Itoa(pending)
	helpers.BroadcastMessage("public-channel", "service-count-changed", data)
}

func (repo *DBRepo) CheckStatus(w http.ResponseWriter, r *http.Request) {
	ok := true
	serviceID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.ErrorLog.Println(err)
		ok = false
	}

	oldStatus := chi.URLParam(r, "oldStatus")

	service, err := repo.DB.GetServiceByID(serviceID)
	if err != nil {
		app.ErrorLog.Println(err)
		ok = false
	}

	host, err := repo.DB.GetHostByID(service.HostID)
	if err != nil {
		app.ErrorLog.Println(err)
		ok = false
	}

	newStatus, message := repo.checkService(host, service)

	event := models.Event{
		HostID:      host.ID,
		ServiceID:   service.ID,
		HostName:    host.Name,
		ServiceName: service.Name,
		EventType:   newStatus,
		Message:     message,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = repo.DB.InsertEvent(event)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	if newStatus != service.Status {
		repo.pushStatusChanged(service, newStatus)
	}

	service.Status = newStatus
	service.LastCheck = time.Now()

	err = repo.DB.UpdateService(service)
	if err != nil {
		app.ErrorLog.Println(err)
		ok = false
	}

	res := jsonResponse{
		OK:        ok,
		Message:   message,
		HostID:    host.ID,
		ServiceID: serviceID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		LastCheck: service.LastCheck,
	}

	out, _ := json.MarshalIndent(res, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (repo *DBRepo) checkService(host models.Host, service models.Service) (string, string) {
	var newStatus, message string

	switch service.Name {
	case HTTP:
		newStatus, message = checkHTTP(host.URL)
	case HTTPS:
		newStatus, message = checkHTTPS(host.URL)
	case SSLCertificate:
		newStatus, message = checkSSLCertificate(host.URL)
	}

	if service.Status != newStatus {
		repo.pushStatusChanged(service, newStatus)

		event := models.Event{
			HostID:      host.ID,
			ServiceID:   service.ID,
			HostName:    host.Name,
			ServiceName: service.Name,
			EventType:   newStatus,
			Message:     message,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err := repo.DB.InsertEvent(event)
		if err != nil {
			app.ErrorLog.Println(err)
		}

		if repo.App.Preferences["notify_via_email"] == "1" {
			if service.Status != "pending" {
				email := channels.Email{
					ToName:    repo.App.Preferences["notify_name"],
					ToAddress: repo.App.Preferences["notify_email"],
				}

				if newStatus == "healthy" {
					email.Subject = fmt.Sprintf("HEALTHY: service %s on %s", service.Name, service.HostName)
					email.Content = template.HTML(fmt.Sprintf(
						`<p>service %s on %s reported healthy status</p>
							 <p><strong>Message received: %s</strong></p>`,
						service.Name, service.HostName, message,
					))
				} else if newStatus == "warning" {
					// not implemented for now
				} else if newStatus == "problem" {
					email.Subject = fmt.Sprintf("PROBLEM: service %s on %s", service.Name, service.HostName)
					email.Content = template.HTML(fmt.Sprintf(
						`<p>service %s on %s reported problem</p>
							 <p><strong>Message received: %s</strong></p>`,
						service.Name, service.HostName, message,
					))
				}

				helpers.SendEmail(email)
			}
		}

		if repo.App.Preferences["notify_via_sms"] == "1" {
			to := repo.App.Preferences["sms_notify_number"]
			msg := ""

			if newStatus == "healthy" {
				msg = fmt.Sprintf("HEALTHY: service %s on %s", service.Name, service.HostName)
			} else if newStatus == "warning" {
				// not implemented for now
			} else if newStatus == "problem" {
				msg = fmt.Sprintf("PROBLEM: service %s on %s", service.Name, service.HostName)
			}

			err = sms.SendSMSViaTwilio(repo.App, to, msg)
			if err != nil {
				app.ErrorLog.Println(err)
			}
		}

	}

	repo.pushScheduleChanged(service, newStatus)

	return newStatus, message
}

func (repo *DBRepo) pushStatusChanged(service models.Service, newStatus string) {
	data := make(map[string]string)
	data["hostID"] = strconv.Itoa(service.HostID)
	data["hostName"] = service.HostName
	data["serviceID"] = strconv.Itoa(service.ID)
	data["serviceName"] = service.Name
	data["icon"] = service.Icon
	data["status"] = newStatus
	data["message"] = fmt.Sprintf("%s on %s reports %s",
		service.Name, service.HostName, newStatus,
	)
	data["lastCheck"] = time.Now().Format("2006-01-02 3:04:12")

	helpers.BroadcastMessage("public-channel", "service-status-changed", data)
}

func (repo *DBRepo) pushScheduleChanged(service models.Service, newStatus string) {
	yearOne := time.Date(0001, 1, 1, 0, 0, 0, 1, time.UTC)
	data := make(map[string]string)
	data["hostID"] = strconv.Itoa(service.HostID)
	data["hostName"] = service.HostName
	data["serviceID"] = strconv.Itoa(service.ID)
	data["serviceName"] = service.Name
	data["icon"] = service.Icon
	data["status"] = newStatus

	if repo.App.Scheduler.Entry(repo.App.MonitorEntries[service.ID]).Next.After(yearOne) {
		data["nextRun"] = repo.App.Scheduler.Entry(repo.App.MonitorEntries[service.ID]).Next.Format("2006-01-02 15:04:05")
	} else {
		data["nextRun"] = "Pending..."
	}

	data["lastRun"] = time.Now().Format("2006-01-02 15:04:05")
	data["schedule"] = fmt.Sprintf("Every %d%s", service.ScheduleNumber, service.ScheduleUnit)

	helpers.BroadcastMessage("public-channel", "schedule-changed", data)
}

func checkHTTP(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "https://", "http://", -1)

	res, err := http.Get(url)
	if err != nil {
		return "problem", fmt.Sprintf("%s - %s", url, "error connecting")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "problem", fmt.Sprintf("%s - %s", url, res.Status)
	}

	return "healthy", fmt.Sprintf("%s - %s", url, res.Status)
}

func checkHTTPS(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "http://", "https://", -1)

	res, err := http.Get(url)
	if err != nil {
		return "problem", fmt.Sprintf("%s - %s", url, "error connecting")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "problem", fmt.Sprintf("%s - %s", url, res.Status)
	}

	return "healthy", fmt.Sprintf("%s - %s", url, res.Status)
}

func checkSSLCertificate(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	if strings.HasPrefix(url, "https://") {
		url = strings.Replace(url, "https://", "", -1)
	}

	if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "", -1)
	}

	conn, err := tls.Dial("tcp", url+":443", nil)
	if err != nil {
		fmt.Println(err)
		return "problem", fmt.Sprintf("%s - Server doesn't support SSL certificate: %s", url, err)
	}

	err = conn.VerifyHostname(url)
	if err != nil {
		fmt.Println(err)
		return "problem", fmt.Sprintf("%s - Hostname doesn't match with certificate: %s", url, err)
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter

	return "healthy", fmt.Sprintf("%s - Issuer: %s\nExpiry: %v", url, conn.ConnectionState().PeerCertificates[0].Issuer, expiry.Format(time.RFC850))
}

func (repo *DBRepo) addToMonitorEntries(service models.Service) {
	if app.Preferences["monitoring_live"] == "1" {
		var scheduleStr string
		if service.ScheduleUnit == "d" {
			scheduleStr = fmt.Sprintf("@every %d%s", service.ScheduleNumber*24, "h")
		} else {
			scheduleStr = fmt.Sprintf("@every %d%s", service.ScheduleNumber, service.ScheduleUnit)
		}

		var j job
		j.ServiceID = service.ID
		scheduleID, err := app.Scheduler.AddJob(scheduleStr, j)
		if err != nil {
			app.ErrorLog.Println(err)
		}

		app.MonitorEntries[service.ID] = scheduleID
	}
}

func (repo *DBRepo) removeFromMonitorEntries(service models.Service) {
	if app.Preferences["monitoring_live"] == "1" {
		repo.App.Scheduler.Remove(app.MonitorEntries[service.ID])

		data := make(map[string]string)
		data["serviceID"] = strconv.Itoa(service.ID)

		helpers.BroadcastMessage("public-channel", "schedule-task-removed", data)
	}
}
