package db_repo

import (
	"context"
	"server_monitor/internal/models"
	"time"
)

func (p *postgresDBRepo) GetAllPreferences() ([]models.Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, name, value FROM preferences"
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []models.Preference
	for rows.Next() {
		preference := &models.Preference{}
		err = rows.Scan(&preference.ID, &preference.Name, &preference.Value)
		if err != nil {
			return nil, err
		}
		preferences = append(preferences, *preference)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return preferences, nil
}

func (p *postgresDBRepo) CreateOrUpdatePreferences(preferences map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for name, value := range preferences {
		query := `UPDATE
    		  	  	preferences
			      SET
			      	value=$1, updated_at=$2
			      WHERE
				  	name=$3`
		result, err := p.DB.ExecContext(ctx, query, value, time.Now(), name)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected < 1 {
			query = `INSERT INTO
						preferences (name, value, created_at, updated_at)
					 VALUES
					 	($1, $2, $3, $4)`
			_, err = p.DB.ExecContext(ctx, query, name, value, time.Now(), time.Now())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
