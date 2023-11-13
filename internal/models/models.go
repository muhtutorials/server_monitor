package models

import (
	"errors"
	"github.com/robfig/cron/v3"
	"time"
)

var (
	ErrNoRecordFound      = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrEmailAlreadyExists = errors.New("models: email already exists")
	ErrInactiveAccount    = errors.New("models: inactive account")
)

type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    []byte
	IsActive    int
	AccessLevel int
	Preferences map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

type Preference struct {
	ID        int
	Name      string
	Value     []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Host struct {
	ID        int
	Name      string
	FullName  string
	URL       string
	IP        string
	IPV6      string
	Location  string
	OS        string
	IsActive  int
	CreatedAt time.Time
	UpdatedAt time.Time
	Services  []Service
}

type Service struct {
	ID             int
	HostID         int
	HostName       string
	Name           string
	Icon           string
	ScheduleNumber int
	ScheduleUnit   string
	Status         string
	IsActive       int
	LastCheck      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Task struct {
	EntryID     cron.EntryID
	Entry       cron.Entry
	HostName    string
	ServiceName string
	ServiceID   int
	Schedule    string
	LastRun     time.Time
}

type Event struct {
	ID          int
	HostID      int
	ServiceID   int
	HostName    string
	ServiceName string
	EventType   string
	Message     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
