package main

import (
	"bytes"
	"github.com/aymerick/douceur/inliner"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"jaytaylor.com/html2text"
	"server_monitor/internal/channels"
	"strconv"
	"time"
)

// todo simplify mail handling

type Dispatcher struct {
	jobQueue   chan channels.Email
	workerPool chan chan channels.Email
	maxWorkers int
}

func NewDispatcher(jobQueue chan channels.Email, maxWorkers int) *Dispatcher {
	return &Dispatcher{
		jobQueue:   jobQueue,
		workerPool: make(chan chan channels.Email, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

func (d *Dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i+1, d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case email := <-d.jobQueue:
			go func() {
				job := <-d.workerPool
				job <- email
			}()
		}
	}
}

type Worker struct {
	id         int
	job        chan channels.Email
	workerPool chan chan channels.Email
	quit       chan struct{}
}

func NewWorker(id int, workerPool chan chan channels.Email) Worker {
	return Worker{
		id:         id,
		job:        make(chan channels.Email),
		workerPool: workerPool,
		quit:       make(chan struct{}),
	}
}

func (w Worker) start() {
	go func() {
		for {
			w.workerPool <- w.job

			select {
			case email := <-w.job:
				w.sendEmail(email)
			case <-w.quit:
				app.InfoLog.Printf("worker%d stopping\n", w.id)
				return
			}
		}
	}()
}

func (w Worker) stop() {
	go func() {
		w.quit <- struct{}{}
	}()
}

func (w Worker) sendEmail(message channels.Email) {
	data := struct {
		FromName    string
		FromAddress string
		Content     template.HTML
		StringMap   map[string]string
		IntMap      map[string]int
		FloatMap    map[string]float32
		RowSets     map[string]any
		Preferences map[string]string
	}{
		FromName:    message.FromName,
		FromAddress: message.FromAddress,
		Content:     message.Content,
		StringMap:   message.StringMap,
		IntMap:      message.IntMap,
		FloatMap:    message.FloatMap,
		RowSets:     message.RowSets,
		Preferences: preferences,
	}

	paths := []string{"./views/mail.tmpl"}

	t := template.Must(template.New("mail.tmpl").ParseFiles(paths...))

	var tmpl bytes.Buffer
	if err := t.Execute(&tmpl, data); err != nil {
		app.ErrorLog.Println(err)
	}

	result := tmpl.String()

	plainText, err := html2text.FromString(result, html2text.Options{PrettyTables: true})

	html, err := inliner.Inline(result)
	if err != nil {
		app.ErrorLog.Println(err)
		html = result
	}

	port, _ := strconv.Atoi(preferences["smtp_port"])

	server := mail.NewSMTPClient()
	server.Host = preferences["smtp_server"]
	server.Port = port
	server.Username = preferences["smtp_user"]
	server.Password = preferences["smtp_password"]
	if preferences["smtp_server"] == "localhost" {
		server.Authentication = mail.AuthPlain
	} else {
		server.Authentication = mail.AuthLogin
	}
	server.Encryption = mail.EncryptionSTARTTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(message.FromAddress).AddTo(message.ToAddress).SetSubject(message.Subject)

	if len(message.AdditionalTo) > 0 {
		for _, item := range message.AdditionalTo {
			email.AddTo(item)
		}
	}

	if len(message.CC) > 0 {
		for _, item := range message.CC {
			email.AddCc(item)
		}
	}

	if len(message.Attachments) > 0 {
		for _, item := range message.Attachments {
			email.AddAttachment(item)
		}
	}

	email.SetBody(mail.TextHTML, html)
	email.AddAlternative(mail.TextPlain, plainText)

	err = email.Send(smtpClient)
	if err != nil {
		app.ErrorLog.Println(err)
	} else {
		app.InfoLog.Println("Email sent")
	}
}
