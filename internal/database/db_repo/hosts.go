package db_repo

import (
	"context"
	"server_monitor/internal/models"
	"time"
)

func (p *postgresDBRepo) GetAllHosts() ([]*models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, full_name, url, ip, ipv6, location, os, is_active, created_at, updated_at
				FROM hosts`
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*models.Host
	for rows.Next() {
		host := &models.Host{}
		err = rows.Scan(
			&host.ID,
			&host.Name,
			&host.FullName,
			&host.URL,
			&host.IP,
			&host.IPV6,
			&host.Location,
			&host.OS,
			&host.IsActive,
			&host.CreatedAt,
			&host.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		services, err := p.GetHostServices(host.ID)
		if err != nil {
			return nil, err
		}

		host.Services = services

		hosts = append(hosts, host)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return hosts, nil
}

func (p *postgresDBRepo) GetHostByID(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, full_name, url, ip, ipv6, location, os, is_active, created_at, updated_at
				FROM hosts WHERE id=$1`
	row := p.DB.QueryRowContext(ctx, query, id)

	var host models.Host
	err := row.Scan(
		&host.ID,
		&host.Name,
		&host.FullName,
		&host.URL,
		&host.IP,
		&host.IPV6,
		&host.Location,
		&host.OS,
		&host.IsActive,
		&host.CreatedAt,
		&host.UpdatedAt,
	)
	if err != nil {
		return host, err
	}

	services, err := p.GetHostServices(id)
	if err != nil {
		return host, err
	}

	host.Services = services

	return host, nil
}

func (p *postgresDBRepo) InsertHost(host models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO hosts (
                   name, full_name, url, ip, ipv6, location, os, is_active, created_at, updated_at
			   ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id`

	var newID int
	err := p.DB.QueryRowContext(ctx, query,
		host.Name,
		host.FullName,
		host.URL,
		host.IP,
		host.IPV6,
		host.Location,
		host.OS,
		host.IsActive,
		time.Now(),
		time.Now()).Scan(&newID)
	if err != nil {
		return newID, err
	}

	query = `INSERT INTO
    		 	services (host_id, host_name, name, icon, created_at, updated_at)  
    		 VALUES
			 	($1, $2, 'HTTP', 'fas fa-server', $3, $4),
			 	($1, $2, 'HTTPS', 'fas fa-server', $3, $4),
			 	($1, $2, 'SSL', 'fas fa-lock', $3, $4)`

	_, err = p.DB.ExecContext(ctx, query, newID, host.Name, time.Now(), time.Now())
	if err != nil {
		return newID, err
	}

	return newID, nil
}

func (p *postgresDBRepo) UpdateHost(host models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE hosts
				SET name=$1, full_name=$2, url=$3, ip=$4, ipv6=$5, location=$6, os=$7, is_active=$8, updated_at=$9
				WHERE id=$10`
	_, err := p.DB.ExecContext(ctx, query,
		host.Name,
		host.FullName,
		host.URL,
		host.IP,
		host.IPV6,
		host.Location,
		host.OS,
		host.IsActive,
		time.Now(),
		host.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) UpdateHostIsActive(hostID, isActive int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE hosts SET is_active=$1 WHERE id=$2`
	_, err := p.DB.ExecContext(ctx, query, isActive, hostID)
	if err != nil {
		return err
	}

	return nil
}
