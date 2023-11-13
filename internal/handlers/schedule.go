package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"net/http"
	"server_monitor/internal/helpers"
	"server_monitor/internal/models"
	"sort"
)

type ByHost []models.Task

func (s ByHost) Len() int {
	return len(s)
}

func (s ByHost) Less(i, j int) bool {
	return s[i].HostName < s[j].HostName
}

func (s ByHost) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (repo *DBRepo) TaskList(w http.ResponseWriter, r *http.Request) {
	var tasks []models.Task

	for serviceID, entryID := range repo.App.MonitorEntries {
		var task models.Task

		task.ServiceID = serviceID
		task.EntryID = entryID
		task.Entry = repo.App.Scheduler.Entry(entryID)

		service, err := repo.DB.GetServiceByID(serviceID)
		if err != nil {
			repo.App.ErrorLog.Println(err)
			return
		}
		task.HostName = service.HostName
		task.ServiceName = service.Name
		task.Schedule = fmt.Sprintf("Every %d%s", service.ScheduleNumber, service.ScheduleUnit)
		task.LastRun = service.LastCheck

		tasks = append(tasks, task)
	}

	sort.Sort(ByHost(tasks))

	vars := make(jet.VarMap)

	vars.Set("tasks", tasks)

	err := helpers.RenderPage(w, r, "schedule", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
