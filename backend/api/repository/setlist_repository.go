package repository

import (
	"context"
	"database/sql"
	"errors"
	"setlist/api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
}

type SetlistRepository interface {
	CreateSetlist(ctx context.Context, db DBTX, name, color string, bandID int) (model.Setlist, error)
	UpdateSetlist(ctx context.Context, setlist model.Setlist) (model.Setlist, error)
	GetSetlistsByBandID(ctx context.Context, bandID int) ([]model.Setlist, error)
	GetSetlistByID(ctx context.Context, id int, bandID int) (model.Setlist, error)
	DeleteSetlist(ctx context.Context, setlistID int, bandID int) error
	GetSetlistItemsBySetlistID(ctx context.Context, setlistID int) ([]model.SetlistItem, error)
	AddItemToSetlist(ctx context.Context, item model.SetlistItem) (model.SetlistItem, error)
	UpdateItemOrder(ctx context.Context, setlistID int, itemIDs []int) error
	UpdateSetlistItem(ctx context.Context, itemID int, bandID int, notes *string) (model.SetlistItem, error)
	DeleteSetlistItem(ctx context.Context, itemID int, bandID int) (int, error)
	CopyItemsToNewSetlist(ctx context.Context, tx DBTX, newSetlistID int, items []model.SetlistItem) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
	GetDB() *pgxpool.Pool
}

type PgSetlistRepository struct {
	DB *pgxpool.Pool
}

func (r PgSetlistRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.DB.Begin(ctx)
}

func (r PgSetlistRepository) GetDB() *pgxpool.Pool {
	return r.DB
}

func (r PgSetlistRepository) CreateSetlist(ctx context.Context, db DBTX, name, color string, bandID int) (model.Setlist, error) {
	var setlist model.Setlist
	query := `
		INSERT INTO setlists (name, color, band_id)
		VALUES ($1, $2, $3)
		RETURNING id, band_id, name, color, is_archived, created_at
	`
	err := db.QueryRow(ctx, query, name, color, bandID).Scan(
		&setlist.ID, &setlist.BandID, &setlist.Name, &setlist.Color, &setlist.IsArchived, &setlist.CreatedAt,
	)
	return setlist, err
}

func (r PgSetlistRepository) UpdateSetlist(ctx context.Context, setlist model.Setlist) (model.Setlist, error) {
	query := `
		UPDATE setlists
		SET name = $1, color = $2, is_archived = $3
		WHERE id = $4 AND band_id = $5
		RETURNING id, band_id, name, color, is_archived, created_at
	`
	err := r.DB.QueryRow(ctx, query, setlist.Name, setlist.Color, setlist.IsArchived, setlist.ID, setlist.BandID).Scan(
		&setlist.ID, &setlist.BandID, &setlist.Name, &setlist.Color, &setlist.IsArchived, &setlist.CreatedAt,
	)
	return setlist, err
}

func (r PgSetlistRepository) GetSetlistsByBandID(ctx context.Context, bandID int) ([]model.Setlist, error) {
	setlists := make([]model.Setlist, 0)
	query := `
		SELECT id, band_id, name, color, is_archived, created_at
		FROM setlists
		WHERE band_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(ctx, query, bandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var setlist model.Setlist
		if err := rows.Scan(&setlist.ID, &setlist.BandID, &setlist.Name, &setlist.Color, &setlist.IsArchived, &setlist.CreatedAt); err != nil {
			return setlists, err
		}
		setlists = append(setlists, setlist)
	}

	return setlists, rows.Err()
}

func (r PgSetlistRepository) GetSetlistByID(ctx context.Context, id int, bandID int) (model.Setlist, error) {
	var setlist model.Setlist
	query := `SELECT id, band_id, name, color, is_archived, created_at FROM setlists WHERE id = $1 AND band_id = $2`
	err := r.DB.QueryRow(ctx, query, id, bandID).Scan(&setlist.ID, &setlist.BandID, &setlist.Name, &setlist.Color, &setlist.IsArchived, &setlist.CreatedAt)
	return setlist, err
}

func (r PgSetlistRepository) DeleteSetlist(ctx context.Context, setlistID int, bandID int) error {
	query := `DELETE FROM setlists WHERE id = $1 AND band_id = $2`
	cmdTag, err := r.DB.Exec(ctx, query, setlistID, bandID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r PgSetlistRepository) GetSetlistItemsBySetlistID(ctx context.Context, setlistID int) ([]model.SetlistItem, error) {
	items := make([]model.SetlistItem, 0)
	query := `
		SELECT
			si.id, si.setlist_id, si.position, si.item_type,
			si.song_id, si.interlude_id, si.notes, si.transition_duration_seconds,
			COALESCE(s.title, i.title) as title,
			COALESCE(s.duration_seconds, i.duration_seconds) as duration_seconds,
			s.tempo,
			i.speaker, 
			i.script,
			s.song_key,
			s.links
		FROM setlist_items si
		LEFT JOIN songs s ON si.song_id = s.id
		LEFT JOIN interludes i ON si.interlude_id = i.id
		WHERE si.setlist_id = $1
		ORDER BY si.position ASC
	`
	rows, err := r.DB.Query(ctx, query, setlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.SetlistItem
		err := rows.Scan(
			&item.ID, &item.SetlistID, &item.Position, &item.ItemType,
			&item.SongID, &item.InterludeID, &item.Notes, &item.TransitionDurationSeconds,
			&item.Title, &item.DurationSeconds, &item.Tempo,
			&item.Speaker, &item.Script,
			&item.SongKey, &item.Links,
		)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r PgSetlistRepository) AddItemToSetlist(ctx context.Context, item model.SetlistItem) (model.SetlistItem, error) {
	var maxPosition *int32
	posQuery := `SELECT MAX(position) FROM setlist_items WHERE setlist_id = $1`
	r.DB.QueryRow(ctx, posQuery, item.SetlistID).Scan(&maxPosition)

	nextPos := 0
	if maxPosition != nil {
		nextPos = int(*maxPosition) + 1
	}
	item.Position = nextPos

	insertQuery := `INSERT INTO setlist_items (setlist_id, position, item_type, song_id, interlude_id, notes, transition_duration_seconds)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					RETURNING id`

	err := r.DB.QueryRow(ctx, insertQuery,
		item.SetlistID,
		item.Position,
		item.ItemType,
		item.SongID,
		item.InterludeID,
		item.Notes,
		item.TransitionDurationSeconds,
	).Scan(&item.ID)

	return item, err
}

func (r PgSetlistRepository) UpdateItemOrder(ctx context.Context, setlistID int, itemIDs []int) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET CONSTRAINTS unique_position_in_setlist DEFERRED"); err != nil {
		return err
	}

	query := `UPDATE setlist_items SET position = $1 WHERE id = $2 AND setlist_id = $3`

	for i, id := range itemIDs {
		if _, err := tx.Exec(ctx, query, i, id, setlistID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r PgSetlistRepository) UpdateSetlistItem(ctx context.Context, itemID int, bandID int, notes *string) (model.SetlistItem, error) {
	var item model.SetlistItem
	query := `
		UPDATE setlist_items si SET notes = $1
		FROM setlists s
		WHERE si.id = $2 AND si.setlist_id = s.id AND s.band_id = $3
		RETURNING si.id, si.setlist_id, si.notes
	`
	err := r.DB.QueryRow(ctx, query, notes, itemID, bandID).Scan(&item.ID, &item.SetlistID, &item.Notes)
	return item, err
}

func (r PgSetlistRepository) DeleteSetlistItem(ctx context.Context, itemID int, bandID int) (int, error) {
	query := `
		DELETE FROM setlist_items si
		USING setlists s
		WHERE si.id = $1 AND si.setlist_id = s.id AND s.band_id = $2
		RETURNING si.setlist_id
	`
	var setlistID int
	if err := r.DB.QueryRow(ctx, query, itemID, bandID).Scan(&setlistID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, err
	}
	return setlistID, nil
}

func (r PgSetlistRepository) CopyItemsToNewSetlist(ctx context.Context, tx DBTX, newSetlistID int, items []model.SetlistItem) error {
	if len(items) == 0 {
		return nil
	}

	rows := make([][]interface{}, len(items))
	for i, item := range items {
		rows[i] = []interface{}{
			newSetlistID,
			item.Position,
			item.ItemType,
			item.SongID,
			item.InterludeID,
			item.Notes,
			item.TransitionDurationSeconds,
		}
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"setlist_items"},
		[]string{"setlist_id", "position", "item_type", "song_id", "interlude_id", "notes", "transition_duration_seconds"},
		pgx.CopyFromRows(rows),
	)

	return err
}
