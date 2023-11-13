package db_repo

import (
	"context"
	"server_monitor/internal/models"
	"time"
)

func (p *postgresDBRepo) InsertEvent(event models.Event) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO
    		  	events (host_id, service_id, host_name, service_name, event_type, message, created_at, updated_at)
			  VALUES
			  	($1, $2, $3, $4, $5, $6, $7, $8) returning id`

	var newID int
	err := p.DB.QueryRowContext(ctx, query,
		event.HostID,
		event.ServiceID,
		event.HostName,
		event.ServiceName,
		event.EventType,
		event.Message,
		time.Now(),
		time.Now()).Scan(&newID)
	if err != nil {
		return newID, err
	}

	return newID, nil
}

func (p *postgresDBRepo) GetAllEvents() ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
    			  id, host_id, service_id, host_name, service_name, event_type, message, created_at, updated_at
			  FROM
				  events
			  ORDER BY
			      created_at`

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err = rows.Scan(
			&event.ID,
			&event.HostID,
			&event.ServiceID,
			&event.HostName,
			&event.ServiceName,
			&event.EventType,
			&event.Message,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return events, nil
}
