package helpers

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/justinas/nosurf"
	"net/http"
	"server_monitor/internal/models"
	"server_monitor/internal/templates"
	"time"
)

// views is the jet template set
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./views"),
	jet.InDevelopmentMode(),
)

func RenderPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data any) error {
	var vars jet.VarMap
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	var td templates.TemplateData
	if data != nil {
		td = data.(templates.TemplateData)
	}

	td = addDefaultTemplateData(td, r)

	addTemplateFunctions()

	t, err := views.GetTemplate(templateName + ".jet")
	if err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		app.ErrorLog.Println(err)
		return err
	}

	return nil
}

func addDefaultTemplateData(td templates.TemplateData, r *http.Request) templates.TemplateData {
	td.IsAuthenticated = IsAuthenticated(r)
	if td.IsAuthenticated {
		u := app.Session.Get(r.Context(), "user").(models.User)
		td.User = u
	}
	td.CSRFToken = nosurf.Token(r)
	td.Preferences = app.Preferences

	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")

	return td
}

func addTemplateFunctions() {
	views.AddGlobal("humanDate", func(t time.Time) string {
		return HumanDate(t)
	})
	views.AddGlobal("formatDate", func(t time.Time, l string) string {
		return FormatDate(t, l)
	})
	views.AddGlobal("dateAfterYearOne", func(t time.Time) bool {
		return DateAfterYearOne(t)
	})
}

// HumanDate formats a time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDate formats a time with provided (go compliant) format string, and returns it as a string
func FormatDate(t time.Time, l string) string {
	return t.Format(l)
}

// DateAfterYearOne is used to verify that a date is after the year 1 (since go hates nulls)
func DateAfterYearOne(t time.Time) bool {
	yearOne := time.Date(0001, 11, 17, 20, 34, 58, 651387237, time.UTC)
	return t.After(yearOne)
}
