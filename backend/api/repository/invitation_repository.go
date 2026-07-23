package repository

import (
	"context"
	"errors"
	"setlist/api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvitationExpired  = errors.New("invitation is expired")
	ErrInvitationNotFound = errors.New("invitation not found")
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *model.Invitation) error
	GetByToken(ctx context.Context, token string) (*model.InvitationDetail, error)
	Delete(ctx context.Context, token string) error
}

type PgInvitationRepository struct {
	DB *pgxpool.Pool
}

func (r *PgInvitationRepository) Create(ctx context.Context, inv *model.Invitation) error {
	query := `
		INSERT INTO band_invitations (token, band_id, role, expires_at, max_uses)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.DB.QueryRow(ctx, query, inv.Token, inv.BandID, inv.Role, inv.ExpiresAt, inv.MaxUses).
		Scan(&inv.ID, &inv.CreatedAt)
	return err
}

func (r *PgInvitationRepository) GetByToken(ctx context.Context, token string) (*model.InvitationDetail, error) {
	query := `
		SELECT i.token, i.band_id, i.role, i.expires_at, b.name as band_name
		FROM band_invitations i
		JOIN bands b ON i.band_id = b.id
		WHERE i.token = $1
	`
	var detail model.InvitationDetail
	err := r.DB.QueryRow(ctx, query, token).
		Scan(&detail.Token, &detail.BandID, &detail.Role, &detail.ExpiresAt, &detail.BandName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	return &detail, nil
}

func (r *PgInvitationRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM band_invitations WHERE token = $1`
	cmdTag, err := r.DB.Exec(ctx, query, token)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *PgInvitationRepository) CleanExpiredInvitations(ctx context.Context) error {
	query := `DELETE FROM band_invitations WHERE expires_at < NOW()`
	_, err := r.DB.Exec(ctx, query)
	return err
}
