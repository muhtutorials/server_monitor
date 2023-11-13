package templates

import "server_monitor/internal/models"

type TemplateData struct {
	IsAuthenticated bool
	User            models.User
	CSRFToken       string
	Preferences     map[string]string
	Flash           string
	Warning         string
	Error           string
	GWVersion       string
}
