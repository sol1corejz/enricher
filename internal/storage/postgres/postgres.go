package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/sol1corejz/enricher/internal/domain/models"
	"github.com/sol1corejz/enricher/internal/storage"
	"os"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("pgx", GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, user models.EnrichedUser) (int64, error) {
	const op = "storage.postgres.SaveUser"

	stmt, err := s.db.Prepare(`INSERT INTO users (name, surname, patronymic, age, sex, country) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRowContext(ctx, user.Name, user.Surname, user.Patronymic, user.Age, user.Sex, user.Country).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) EditUser(ctx context.Context, user models.EnrichedUser) (models.EnrichedUser, error) {
	const op = "storage.postgres.EditUser"

	stmt, err := s.db.Prepare(`
		UPDATE users 
		SET name = $1, surname = $2, patronymic = $3, age = $4, sex = $5, country = $6
		WHERE id = $7
		RETURNING id, name, surname, patronymic, age, sex, country
	`)
	if err != nil {
		return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var countryData []byte
	var updatedUser models.EnrichedUser
	err = stmt.QueryRowContext(
		ctx,
		user.Name,
		user.Surname,
		user.Patronymic,
		user.Age,
		user.Sex,
		user.Country,
		user.ID,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Name,
		&updatedUser.Surname,
		&updatedUser.Patronymic,
		&updatedUser.Age,
		&updatedUser.Sex,
		&countryData,
	)
	if err != nil {
		return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(countryData) > 0 {
		if err := json.Unmarshal(countryData, &updatedUser.Country); err != nil {
			return models.EnrichedUser{}, fmt.Errorf("%s: failed to unmarshal country data: %w", op, err)
		}
	}

	return updatedUser, nil
}

func (s *Storage) DeleteUser(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteUser"

	stmt, err := s.db.Prepare(`DELETE FROM users WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, sql.ErrNoRows)
	}

	return nil
}
func (s *Storage) GetUser(ctx context.Context, id int64) (models.EnrichedUser, error) {
	const op = "storage.postgres.GetUser"

	stmt, err := s.db.Prepare(`
        SELECT id, name, surname, patronymic, age, sex, country 
        FROM users 
        WHERE id = $1
    `)
	if err != nil {
		return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var user models.EnrichedUser
	var countryData []byte

	err = stmt.QueryRowContext(ctx, id).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Patronymic,
		&user.Age,
		&user.Sex,
		&countryData,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(countryData) > 0 {
		if err := json.Unmarshal(countryData, &user.Country); err != nil {
			return models.EnrichedUser{}, fmt.Errorf("%s: failed to unmarshal country data: %w", op, err)
		}
	}

	return user, nil
}

func (s *Storage) GetUsers(ctx context.Context, filter models.UserFilter) ([]models.EnrichedUser, error) {
	const op = "storage.postgres.GetUsers"

	baseQuery := `SELECT id, name, surname, patronymic, age, sex, country FROM users WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filter.Name != "" {
		baseQuery += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}

	if filter.Surname != "" {
		baseQuery += fmt.Sprintf(" AND surname ILIKE $%d", argPos)
		args = append(args, "%"+filter.Surname+"%")
		argPos++
	}

	if filter.Patronymic != "" {
		baseQuery += fmt.Sprintf(" AND patronymic ILIKE $%d", argPos)
		args = append(args, "%"+filter.Patronymic+"%")
		argPos++
	}

	if filter.AgeFrom > 0 {
		baseQuery += fmt.Sprintf(" AND age >= $%d", argPos)
		args = append(args, filter.AgeFrom)
		argPos++
	}

	if filter.AgeTo > 0 {
		baseQuery += fmt.Sprintf(" AND age <= $%d", argPos)
		args = append(args, filter.AgeTo)
		argPos++
	}

	if filter.Sex != "" {
		baseQuery += fmt.Sprintf(" AND sex = $%d", argPos)
		args = append(args, filter.Sex)
		argPos++
	}

	if filter.Country != "" {
		baseQuery += fmt.Sprintf(" AND country::text ILIKE $%d", argPos)
		args = append(args, "%"+filter.Country+"%")
		argPos++
	}

	baseQuery += " ORDER BY id ASC"

	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []models.EnrichedUser

	for rows.Next() {
		var user models.EnrichedUser
		var countryData []byte

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Patronymic,
			&user.Age,
			&user.Sex,
			&countryData,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if len(countryData) > 0 {
			if err := json.Unmarshal(countryData, &user.Country); err != nil {
				return nil, fmt.Errorf("%s: failed to unmarshal country data: %w", op, err)
			}
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func GetDatabaseURL() string {
	if err := godotenv.Load(); err == nil {
		dbURL := os.Getenv("DB_URL")
		if dbURL != "" {
			return dbURL
		}
	}

	panic("DB_URL not found")
}
