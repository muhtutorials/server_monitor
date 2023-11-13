package db_repo

import (
	"context"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"server_monitor/internal/models"
	"time"
)

func (p *postgresDBRepo) GetAllUsers() ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, is_active, created_at, updated_at FROM users
				WHERE deleted_at is null`
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		err = rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		app.ErrorLog.Println(err)
		return nil, err
	}

	return users, nil
}

func (p *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, is_active, access_level, created_at, updated_at
				FROM users WHERE id=$1`
	row := p.DB.QueryRowContext(ctx, query, id)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.IsActive,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (p *postgresDBRepo) InsertUser(u models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword(u.Password, 12)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO users (
                   first_name, last_name, email, password, is_active, access_level, created_at, updated_at
			   ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	var newID int
	err = p.DB.QueryRowContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		hashedPassword,
		u.IsActive,
		u.AccessLevel,
		time.Now(),
		time.Now()).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (p *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE users
				SET first_name=$1, last_name=$2, email=$3, is_active=$4, access_level=$5, updated_at=$6
				WHERE id=$7`
	_, err := p.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.IsActive,
		u.AccessLevel,
		time.Now(),
		u.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "UPDATE users SET is_active=0, deleted_at=$1 WHERE id=$2"
	_, err := p.DB.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) UpdatePassword(id int, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	query := "UPDATE users SET password=$1, updated_at=$2 WHERE id=$3"
	_, err = p.DB.ExecContext(ctx, query, hashedPassword, time.Now(), id)
	if err != nil {
		return err
	}

	query = "DELETE FROM remember_me_tokens WHERE user_id=$1"
	_, err = p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) Authenticate(email, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string
	var isActive int

	query := "SELECT id, password, is_active FROM users WHERE email=$1 and deleted_at is null"
	row := p.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword, &isActive)
	if err == sql.ErrNoRows {
		return id, hashedPassword, models.ErrInvalidCredentials
	} else if err != nil {
		return id, hashedPassword, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return id, hashedPassword, models.ErrInvalidCredentials
	} else if err != nil {
		return id, hashedPassword, err
	}

	if isActive == 0 {
		return id, hashedPassword, models.ErrInactiveAccount
	}

	return id, hashedPassword, nil
}

func (p *postgresDBRepo) InsertRememberMeToken(userID int, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO remember_me_tokens (user_id, token) VALUES ($1, $2)"
	_, err := p.DB.ExecContext(ctx, query, userID, token)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) DeleteRememberMeToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "DELETE FROM remember_me_tokens WHERE token=$1"
	_, err := p.DB.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresDBRepo) CheckForRememberMeToken(id int, token string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id FROM remember_me_tokens WHERE user_id=$1 AND token=$2"
	row := p.DB.QueryRowContext(ctx, query, id, token)
	err := row.Scan(&id)

	return err == nil
}
