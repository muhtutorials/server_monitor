package handlers

import (
	"fmt"
)

type job struct {
	ServiceID int
}

func (j job) Run() {
	Repo.StartScheduledCheck(j.ServiceID)
}

func StartMonitoring() {
	if app.Preferences["monitoring_live"] == "1" {
		data := make(map[string]string)
		data["message"] = "Monitoring is starting..."
		err := app.WS.Trigger("public-channel", "app-starting", data)
		if err != nil {
			app.ErrorLog.Println(err)
		}

		servicesToMonitor, err := Repo.DB.GetServicesToMonitor()
		if err != nil {
			app.ErrorLog.Println(err)
		}

		for _, service := range servicesToMonitor {
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

			Repo.pushScheduleChanged(service, "pending")
		}

		app.Scheduler.Start()
	}
}

func StopMonitoring() {
	for _, v := range app.MonitorEntries {
		app.Scheduler.Remove(v)
	}

	for k := range app.MonitorEntries {
		delete(app.MonitorEntries, k)
	}

	for _, e := range app.Scheduler.Entries() {
		app.Scheduler.Remove(e.ID)
	}

	data := make(map[string]string)
	data["message"] = "Monitoring has been stopped"
	err := app.WS.Trigger("public-channel", "app-stopping", data)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	app.Scheduler.Stop()
}
