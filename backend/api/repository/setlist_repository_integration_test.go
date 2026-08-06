package repository_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"setlist/api/model"
	"setlist/api/repository"
)

const concurrentAdds = 10

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL non défini, test d'intégration ignoré (voir make test-backend-integration)")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL invalide: %v", err)
	}
	cfg.MaxConns = concurrentAdds + 2

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connexion à la base de test: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("base de test injoignable: %v", err)
	}

	return pool
}

func seedSetlistWithSongs(t *testing.T, pool *pgxpool.Pool, songCount int) (int, []int32) {
	t.Helper()
	ctx := context.Background()

	var bandID int
	bandName := fmt.Sprintf("concurrency-%d", time.Now().UnixNano())
	if err := pool.QueryRow(ctx, `INSERT INTO bands (name) VALUES ($1) RETURNING id`, bandName).Scan(&bandID); err != nil {
		t.Fatalf("création du groupe: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM bands WHERE id = $1`, bandID); err != nil {
			t.Errorf("nettoyage du groupe %d: %v", bandID, err)
		}
	})

	var setlistID int
	if err := pool.QueryRow(ctx, `INSERT INTO setlists (band_id, name) VALUES ($1, $2) RETURNING id`, bandID, bandName).Scan(&setlistID); err != nil {
		t.Fatalf("création de la setlist: %v", err)
	}

	songIDs := make([]int32, songCount)
	for i := range songIDs {
		title := fmt.Sprintf("%s-song-%d", bandName, i)
		if err := pool.QueryRow(ctx, `INSERT INTO songs (band_id, title) VALUES ($1, $2) RETURNING id`, bandID, title).Scan(&songIDs[i]); err != nil {
			t.Fatalf("création du morceau %d: %v", i, err)
		}
	}

	return setlistID, songIDs
}

func TestAddItemToSetlistConcurrent(t *testing.T) {
	pool := newTestPool(t)
	repo := repository.PgSetlistRepository{DB: pool}

	setlistID, songIDs := seedSetlistWithSongs(t, pool, concurrentAdds)

	errs := make([]error, concurrentAdds)
	positions := make([]int, concurrentAdds)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range songIDs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			songID := songIDs[i]
			<-start
			item, err := repo.AddItemToSetlist(context.Background(), model.SetlistItem{
				SetlistID: setlistID,
				ItemType:  "song",
				SongID:    &songID,
			})
			errs[i] = err
			positions[i] = item.Position
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("ajout concurrent %d en échec: %v", i, err)
		}
	}
	if t.Failed() {
		t.FailNow()
	}

	seen := make(map[int]int, concurrentAdds)
	for i, pos := range positions {
		if previous, duplicate := seen[pos]; duplicate {
			t.Fatalf("position %d attribuée aux ajouts %d et %d", pos, previous, i)
		}
		seen[pos] = i
	}
	for expected := 0; expected < concurrentAdds; expected++ {
		if _, ok := seen[expected]; !ok {
			t.Fatalf("position %d manquante: les positions ne sont pas contiguës", expected)
		}
	}

	items, err := repo.GetSetlistItemsBySetlistID(context.Background(), setlistID)
	if err != nil {
		t.Fatalf("relecture des éléments: %v", err)
	}
	if len(items) != concurrentAdds {
		t.Fatalf("attendu %d éléments en base, obtenu %d", concurrentAdds, len(items))
	}
	for i, item := range items {
		if item.Position != i {
			t.Fatalf("élément %d: position attendue %d, obtenue %d", item.ID, i, item.Position)
		}
	}
}
