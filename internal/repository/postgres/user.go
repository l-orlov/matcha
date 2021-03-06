package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/l-orlov/matcha/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	usersTable = "users"
)

type UserPostgres struct {
	db        *sqlx.DB
	dbTimeout time.Duration
}

func NewUserPostgres(db *sqlx.DB, dbTimeout time.Duration) *UserPostgres {
	return &UserPostgres{
		db:        db,
		dbTimeout: dbTimeout,
	}
}

func (r *UserPostgres) CreateUser(ctx context.Context, user models.UserToCreate) (uint64, error) {
	query := fmt.Sprintf(`
INSERT INTO %s (email, username, first_name, last_name, password)
VALUES ($1, $2, $3, $4, $5) RETURNING id`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	row := r.db.QueryRowContext(dbCtx, query,
		&user.Email, &user.Username, &user.FirstName, &user.LastName, &user.Password)
	if err := row.Err(); err != nil {
		return 0, getDBError(err)
	}

	var id uint64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UserPostgres) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := fmt.Sprintf(`
SELECT id, email, username, first_name, last_name, password FROM %s WHERE username=$1`, usersTable)
	var user models.User

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	if err := r.db.GetContext(dbCtx, &user, query, &username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserPostgres) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := fmt.Sprintf(`
SELECT id, email, username, first_name, last_name, password FROM %s WHERE email=$1`, usersTable)
	var user models.User

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	if err := r.db.GetContext(dbCtx, &user, query, &email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserPostgres) GetUserByID(ctx context.Context, id uint64) (*models.User, error) {
	query := fmt.Sprintf(`
SELECT id, email, username, first_name, last_name, password, is_email_confirmed FROM %s WHERE id=$1`, usersTable)
	var user models.User

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	if err := r.db.GetContext(dbCtx, &user, query, &id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserPostgres) UpdateUser(ctx context.Context, user models.User) error {
	query := fmt.Sprintf(`
UPDATE %s SET username = $1, first_name = $2, last_name = $3 WHERE id = $4`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	_, err := r.db.ExecContext(dbCtx, query, &user.Username, &user.FirstName, &user.LastName, &user.ID)
	if err != nil {
		return getDBError(err)
	}

	return nil
}

func (r *UserPostgres) UpdateUserPassword(ctx context.Context, userID uint64, password string) error {
	query := fmt.Sprintf(`UPDATE %s SET password = $1 WHERE id = $2`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	if _, err := r.db.ExecContext(dbCtx, query, &password, &userID); err != nil {
		return getDBError(err)
	}

	return nil
}

func (r *UserPostgres) GetAllUsers(ctx context.Context) ([]models.User, error) {
	query := fmt.Sprintf(`
SELECT id, email, username, first_name, last_name, is_email_confirmed FROM %s`, usersTable)
	var users []models.User

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	err := r.db.SelectContext(dbCtx, &users, query)

	return users, err
}

func (r *UserPostgres) DeleteUser(ctx context.Context, id uint64) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	if _, err := r.db.ExecContext(dbCtx, query, &id); err != nil {
		return err
	}

	return nil
}

func (r *UserPostgres) ConfirmEmail(ctx context.Context, id uint64) error {
	query := fmt.Sprintf(`UPDATE %s SET is_email_confirmed = true WHERE id = $1`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	if _, err := r.db.ExecContext(dbCtx, query, &id); err != nil {
		return getDBError(err)
	}

	return nil
}

func (r *UserPostgres) GetUserProfileByID(ctx context.Context, id uint64) (*models.UserProfile, error) {
	query := fmt.Sprintf(`
SELECT id, email, username, first_name, last_name, is_email_confirmed,
gender, sexual_preferences, biography, tags, avatar_path, likes_num, views_num, gps_position
FROM %s WHERE id=$1`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	row := r.db.QueryRowContext(dbCtx, query, &id)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var user models.UserProfile
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.FirstName, &user.LastName,
		&user.IsEmailConfirmed, &user.Gender, &user.SexualPreferences, &user.Biography,
		pq.Array(&user.Tags), &user.AvatarPath, &user.LikesNum, &user.ViewsNum, &user.GPSPosition,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserPostgres) UpdateUserProfile(ctx context.Context, user models.UserProfile) error {
	query := fmt.Sprintf(`
UPDATE %s SET username = $1, first_name = $2, last_name = $3,
gender = $4, sexual_preferences = $5, biography = $6, tags = $7,
likes_num = $8, views_num = $9, gps_position = $10
WHERE id = $11`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	_, err := r.db.ExecContext(dbCtx, query, &user.Username, &user.FirstName, &user.LastName,
		&user.Gender, &user.SexualPreferences, &user.Biography, pq.Array(&user.Tags),
		&user.LikesNum, &user.ViewsNum, &user.GPSPosition, &user.ID)
	if err != nil {
		return getDBError(err)
	}

	return nil
}

func (r *UserPostgres) UpdateUserAvatarPath(ctx context.Context, userID uint64, avatarPath string) error {
	query := fmt.Sprintf(`UPDATE %s SET avatar_path = $1 WHERE id = $2`, usersTable)

	dbCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	_, err := r.db.ExecContext(dbCtx, query, &avatarPath, &userID)
	if err != nil {
		return getDBError(err)
	}

	return nil
}
