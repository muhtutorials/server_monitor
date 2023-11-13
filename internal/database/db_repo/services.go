package db_repo

import (
	"context"
	"server_monitor/internal/models"
	"time"
)

func (p *postgresDBRepo) GetHostServices(hostId int) ([]models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
			  	id, host_id, host_name, name, icon, schedule_number, schedule_unit,
				status, is_active, last_check, created_at, updated_at
			  FROM
			  	services
			  WHERE
			  	host_id=$1
			  ORDER BY
              	name`
	rows, err := p.DB.QueryContext(ctx, query, hostId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service
		err = rows.Scan(
			&service.ID,
			&service.HostID,
			&service.HostName,
			&service.Name,
			&service.Icon,
			&service.ScheduleNumber,
			&service.ScheduleUnit,
			&service.Status,
			&service.IsActive,
			&service.LastCheck,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return services, nil
}

func (p *postgresDBRepo) GetServiceByID(id int) (models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
    			  id, host_id, host_name, name, icon, schedule_number, schedule_unit,
    			  status, is_active, last_check, created_at, updated_at
			  FROM
				  services
			  WHERE
				  id=$1`
	row := p.DB.QueryRowContext(ctx, query, id)

	var service models.Service
	err := row.Scan(
		&service.ID,
		&service.HostID,
		&service.HostName,
		&service.Name,
		&service.Icon,
		&service.ScheduleNumber,
		&service.ScheduleUnit,
		&service.Status,
		&service.IsActive,
		&service.LastCheck,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	if err != nil {
		return service, err
	}

	return service, nil
}

func (p *postgresDBRepo) UpdateService(service models.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE
    		  	services
			  SET
			    host_id=$1, host_name=$2, name=$3, icon=$4, schedule_number=$5, schedule_unit=$6,
    				status=$7, is_active=$8, last_check=$9, updated_at=$10
			  WHERE
				id=$11`

	_, err := p.DB.ExecContext(ctx, query,
		service.HostID,
		service.HostName,
		service.Name,
		service.Icon,
		service.ScheduleNumber,
		service.ScheduleUnit,
		service.Status,
		service.IsActive,
		time.Now(),
		time.Now(),
		service.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) UpdateServiceIsActive(serviceID, hostID, isActive int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE services SET is_active=$1 WHERE id=$2 AND host_id=$3`
	_, err := p.DB.ExecContext(ctx, query, isActive, serviceID, hostID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) GetServiceStatusCounts() (int, int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
              	(SELECT COUNT(id) FROM services WHERE is_active=1 AND status='healthy') as healthy,
				(SELECT COUNT(id) FROM services WHERE is_active=1 AND status='warning') as warning,
				(SELECT COUNT(id) FROM services WHERE is_active=1 AND status='problem') as problem,
				(SELECT COUNT(id) FROM services WHERE is_active=1 AND status='pending') as pending
				`
	row := p.DB.QueryRowContext(ctx, query)
	var healthy, warning, problem, pending int

	err := row.Scan(&healthy, &warning, &problem, &pending)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return healthy, warning, problem, pending, nil
}

func (p *postgresDBRepo) GetServicesByStatus(status string) ([]models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
    			  id, host_id, host_name, name, icon, schedule_number, schedule_unit,
    			  status, is_active, last_check, created_at, updated_at
			  FROM
				  services
			  WHERE
				  status=$1`
	rows, err := p.DB.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service
		err = rows.Scan(
			&service.ID,
			&service.HostID,
			&service.HostName,
			&service.Name,
			&service.Icon,
			&service.ScheduleNumber,
			&service.ScheduleUnit,
			&service.Status,
			&service.IsActive,
			&service.LastCheck,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return services, nil
}

func (p *postgresDBRepo) GetServicesToMonitor() ([]models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT
    			  s.id, s.host_id, s.host_name, s.name, s.icon, s.schedule_number, s.schedule_unit,
    			  s.status, s.is_active, s.last_check, s.created_at, s.updated_at
			  FROM
				  services AS s
			  LEFT JOIN
				  hosts AS h
			  ON
			      h.id=s.host_id
			  WHERE
				  s.is_active=1 AND h.is_active=1`
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service
		err = rows.Scan(
			&service.ID,
			&service.HostID,
			&service.HostName,
			&service.Name,
			&service.Icon,
			&service.ScheduleNumber,
			&service.ScheduleUnit,
			&service.Status,
			&service.IsActive,
			&service.LastCheck,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return services, nil
}
