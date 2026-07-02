package repository

import (
	"context"
	"errors"
	"setlist/api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicateUsername = errors.New("username already exists")
)

type UserRepository interface {
	CreateBandAndUser(ctx context.Context, bandName, username, passwordHash string) (model.User, model.Band, error)
	CreateUser(ctx context.Context, username, passwordHash string) (model.User, error)
	CreateUserAndAddToBand(ctx context.Context, bandID int, username, passwordHash, role string) (model.User, error)
	CreateBand(ctx context.Context, name string, ownerUserID int) (model.Band, error)
	GetMembersByBandID(ctx context.Context, bandID int) ([]model.BandMember, error)
	RemoveUserFromBand(ctx context.Context, bandID int, userID int) error
	GetUserRoleInBand(ctx context.Context, userID int, bandID int) (string, error)
	FindUserByUsername(ctx context.Context, username string) (model.User, error)
	FindUserByID(ctx context.Context, id int) (model.User, error)
	UpdatePassword(ctx context.Context, userID int, newHash string) error
	FindBandsByUserID(ctx context.Context, userID int) ([]model.Band, error)
	FindBandsWithRoleByUserID(ctx context.Context, userID int) ([]model.BandWithRole, error)
	IsUserInBand(ctx context.Context, userID int, bandID int) (bool, error)
	GetAdminCountInBand(ctx context.Context, bandID int) (int, error)
	AddUserToBand(ctx context.Context, userID, bandID int, role string) error
	UpdateUserRoleInBand(ctx context.Context, userID, bandID int, role string) error
	SetDefaultBand(ctx context.Context, userID, bandID int) error
}

type PgUserRepository struct {
	DB *pgxpool.Pool
}

func (r *PgUserRepository) CreateBandAndUser(ctx context.Context, bandName, username, passwordHash string) (model.User, model.Band, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return model.User{}, model.Band{}, err
	}
	defer tx.Rollback(ctx)

	var band model.Band
	bandQuery := `INSERT INTO bands (name) VALUES ($1) RETURNING id, name`
	err = tx.QueryRow(ctx, bandQuery, bandName).Scan(&band.ID, &band.Name)
	if err != nil {
		return model.User{}, model.Band{}, err
	}

	var user model.User
	userQuery := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at`
	err = tx.QueryRow(ctx, userQuery, username, passwordHash).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, model.Band{}, ErrDuplicateUsername
		}
		return model.User{}, model.Band{}, err
	}

	linkQuery := `INSERT INTO band_users (user_id, band_id, role) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, linkQuery, user.ID, band.ID, "admin")
	if err != nil {
		return model.User{}, model.Band{}, err
	}

	return user, band, tx.Commit(ctx)
}

func (r *PgUserRepository) CreateUser(ctx context.Context, username, passwordHash string) (model.User, error) {
	var user model.User
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at`
	err := r.DB.QueryRow(ctx, query, username, passwordHash).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrDuplicateUsername
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *PgUserRepository) CreateUserAndAddToBand(ctx context.Context, bandID int, username, passwordHash, role string) (model.User, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return model.User{}, err
	}
	defer tx.Rollback(ctx)

	var user model.User
	userQuery := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at`
	err = tx.QueryRow(ctx, userQuery, username, passwordHash).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // duplicate key
			return model.User{}, ErrDuplicateUsername
		}
		return model.User{}, err
	}

	linkQuery := `INSERT INTO band_users (user_id, band_id, role) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, linkQuery, user.ID, bandID, role)
	if err != nil {
		return model.User{}, err
	}

	return user, tx.Commit(ctx)
}

func (r *PgUserRepository) GetMembersByBandID(ctx context.Context, bandID int) ([]model.BandMember, error) {
	members := make([]model.BandMember, 0)
	query := `
		SELECT u.id, u.username, bu.role
		FROM users u
		JOIN band_users bu ON u.id = bu.user_id
		WHERE bu.band_id = $1
		ORDER BY u.username
	`
	rows, err := r.DB.Query(ctx, query, bandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member model.BandMember
		if err := rows.Scan(&member.ID, &member.Username, &member.Role); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *PgUserRepository) RemoveUserFromBand(ctx context.Context, bandID int, userID int) error {
	var adminCount int
	countQuery := `SELECT COUNT(*) FROM band_users WHERE band_id = $1 AND role = 'admin'`
	err := r.DB.QueryRow(ctx, countQuery, bandID).Scan(&adminCount)
	if err != nil {
		return err
	}

	if adminCount <= 1 {
		var userRole string
		roleQuery := `SELECT role FROM band_users WHERE user_id = $1 AND band_id = $2`
		err := r.DB.QueryRow(ctx, roleQuery, userID, bandID).Scan(&userRole)
		if err != nil {
			return err
		}
		if userRole == "admin" {
			return errors.New("cannot remove the last admin of the band")
		}
	}

	query := `DELETE FROM band_users WHERE user_id = $1 AND band_id = $2`
	cmdTag, err := r.DB.Exec(ctx, query, userID, bandID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PgUserRepository) GetUserRoleInBand(ctx context.Context, userID int, bandID int) (string, error) {
	var role string
	query := `SELECT role FROM band_users WHERE user_id = $1 AND band_id = $2`
	err := r.DB.QueryRow(ctx, query, userID, bandID).Scan(&role)
	return role, err
}

func (r *PgUserRepository) FindUserByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	query := `SELECT id, password_hash, username FROM users WHERE username = $1`
	err := r.DB.QueryRow(ctx, query, username).Scan(&user.ID, &user.PasswordHash, &user.Username)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *PgUserRepository) FindUserByID(ctx context.Context, id int) (model.User, error) {
	var user model.User
	query := `SELECT id, password_hash, username FROM users WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).Scan(&user.ID, &user.PasswordHash, &user.Username)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *PgUserRepository) UpdatePassword(ctx context.Context, userID int, newHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	cmdTag, err := r.DB.Exec(ctx, query, newHash, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found or no update was needed")
	}
	return nil
}

func (r *PgUserRepository) FindBandsByUserID(ctx context.Context, userID int) ([]model.Band, error) {
	var bands []model.Band
	query := `
		SELECT b.id, b.name FROM bands b
		JOIN band_users bu ON b.id = bu.band_id
		WHERE bu.user_id = $1
		ORDER BY b.name
	`
	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var band model.Band
		if err := rows.Scan(&band.ID, &band.Name); err != nil {
			return nil, err
		}
		bands = append(bands, band)
	}

	return bands, rows.Err()
}

func (r *PgUserRepository) IsUserInBand(ctx context.Context, userID int, bandID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM band_users WHERE user_id = $1 AND band_id = $2)`
	err := r.DB.QueryRow(ctx, query, userID, bandID).Scan(&exists)
	return exists, err
}

func (r *PgUserRepository) AddUserToBand(ctx context.Context, userID, bandID int, role string) error {
	query := `INSERT INTO band_users (user_id, band_id, role) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(ctx, query, userID, bandID, role)
	return err
}

func (r *PgUserRepository) UpdateUserRoleInBand(ctx context.Context, userID, bandID int, role string) error {
	query := `UPDATE band_users SET role = $1 WHERE user_id = $2 AND band_id = $3`
	cmdTag, err := r.DB.Exec(ctx, query, role, userID, bandID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PgUserRepository) CreateBand(ctx context.Context, name string, ownerUserID int) (model.Band, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return model.Band{}, err
	}
	defer tx.Rollback(ctx)

	var band model.Band
	err = tx.QueryRow(ctx,
		`INSERT INTO bands (name) VALUES ($1) RETURNING id, name, created_at`,
		name,
	).Scan(&band.ID, &band.Name, &band.CreatedAt)
	if err != nil {
		return model.Band{}, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO band_users (user_id, band_id, role) VALUES ($1, $2, 'admin')`,
		ownerUserID, band.ID,
	)
	if err != nil {
		return model.Band{}, err
	}

	return band, tx.Commit(ctx)
}

func (r *PgUserRepository) FindBandsWithRoleByUserID(ctx context.Context, userID int) ([]model.BandWithRole, error) {
	bands := make([]model.BandWithRole, 0)
	query := `
		SELECT b.id, b.name, bu.role, bu.is_default, b.created_at
		FROM bands b
		JOIN band_users bu ON bu.band_id = b.id
		WHERE bu.user_id = $1
		ORDER BY bu.is_default DESC, b.name`

	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b model.BandWithRole
		if err := rows.Scan(&b.ID, &b.Name, &b.Role, &b.IsDefault, &b.CreatedAt); err != nil {
			return nil, err
		}
		bands = append(bands, b)
	}
	return bands, rows.Err()
}

func (r *PgUserRepository) GetAdminCountInBand(ctx context.Context, bandID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM band_users WHERE band_id = $1 AND role = 'admin'`
	err := r.DB.QueryRow(ctx, query, bandID).Scan(&count)
	return count, err
}

func (r *PgUserRepository) SetDefaultBand(ctx context.Context, userID, bandID int) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE band_users SET is_default = FALSE WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE band_users SET is_default = TRUE WHERE user_id = $1 AND band_id = $2`, userID, bandID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("band not found or user is not a member")
	}

	return tx.Commit(ctx)
}
