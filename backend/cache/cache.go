package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(redisURL string) *redis.Client {
	if redisURL == "" {
		log.Println("[cache] REDIS_URL not set — Redis cache disabled (fail-safe mode).")
		return nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("[cache] invalid Redis URL (%s) — cache disabled: %v\n", redisURL, err)
		return nil
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("[cache] cannot reach Redis — cache disabled: %v\n", err)
		client.Close()
		return nil
	}

	log.Println("[cache] Redis connection established ✓")
	return client
}

func SongKey(bandID int) string {
	return fmt.Sprintf("band:%d:songs", bandID)
}

func ProfileKey(userID int, bandID int) string {
	return fmt.Sprintf("user:%d:band:%d:profile", userID, bandID)
}

func SetlistKey(bandID int) string {
	return fmt.Sprintf("band:%d:setlists", bandID)
}

// SetlistDetailKey is the cache key for a single setlist's full details
// (setlist + its items). GetDetails is not cached yet; this key and the
// invalidations in AddItem/UpdateItem/DeleteItem/UpdateOrder are prepared so
// caching GetDetails later requires no extra wiring.
func SetlistDetailKey(setlistID int) string {
	return fmt.Sprintf("setlist:%d:details", setlistID)
}

func Get(ctx context.Context, client *redis.Client, key string) (string, bool) {
	if client == nil {
		return "", false
	}
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func Set(ctx context.Context, client *redis.Client, key string, value string, ttl time.Duration) {
	if client == nil {
		return
	}
	if err := client.Set(ctx, key, value, ttl).Err(); err != nil {
		log.Printf("[cache] error writing key %s: %v\n", key, err)
	}
}

func Delete(ctx context.Context, client *redis.Client, key string) {
	if client == nil {
		return
	}
	if err := client.Del(ctx, key).Err(); err != nil {
		log.Printf("[cache] error deleting key %s: %v\n", key, err)
	}
}
