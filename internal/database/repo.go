package database

import "server_monitor/internal/models"

// Repo is database repository
type Repo interface {
	GetAllPreferences() ([]models.Preference, error)
	CreateOrUpdatePreferences(preferences map[string]string) error

	GetAllUsers() ([]*models.User, error)
	GetUserByID(id int) (models.User, error)
	InsertUser(u models.User) (int, error)
	UpdateUser(u models.User) error
	DeleteUser(id int) error
	UpdatePassword(id int, newPassword string) error
	Authenticate(email, password string) (int, string, error)
	InsertRememberMeToken(userID int, token string) error
	DeleteRememberMeToken(token string) error
	CheckForRememberMeToken(id int, token string) bool

	GetAllHosts() ([]*models.Host, error)
	GetHostByID(id int) (models.Host, error)
	InsertHost(h models.Host) (int, error)
	UpdateHost(h models.Host) error
	UpdateHostIsActive(hostID, isActive int) error

	GetHostServices(hostId int) ([]models.Service, error)
	GetServiceByID(id int) (models.Service, error)
	UpdateService(service models.Service) error
	UpdateServiceIsActive(hostID, serviceID, isActive int) error
	GetServiceStatusCounts() (int, int, int, int, error)
	GetServicesByStatus(status string) ([]models.Service, error)
	GetServicesToMonitor() ([]models.Service, error)

	InsertEvent(event models.Event) (int, error)
	GetAllEvents() ([]models.Event, error)
}
