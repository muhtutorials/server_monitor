package handlers

import (
	"github.com/CloudyKit/jet/v6"
	"net/http"
	"server_monitor/internal/helpers"
)

func (repo *DBRepo) HealthyServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("healthy")
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "healthy_services", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) WarningServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("warning")
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "warning_services", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) ProblemServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("problem")
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "problem_services", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

func (repo *DBRepo) PendingServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("pending")
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "pending_services", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
