package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Fenêtre de débit : max rateMaxInWindow requêtes par rateWindow.
	rateWindow      = 12 * time.Second
	rateMaxInWindow = 5
	// Durée de vie du compteur d'échecs (reset glissant en cas d'inactivité).
	failTTL = 15 * time.Minute
)

// RateLimiter applique la limitation de débit et la protection brute-force.
// L'état est stocké dans Redis pour être partagé entre instances et survivre
// aux redémarrages de conteneurs.
type RateLimiter struct {
	client         *redis.Client
	enabled        bool
	trustedProxies []*net.IPNet
}

// NewRateLimiter crée un RateLimiter. Si client est nil (Redis indisponible),
// le middleware fonctionne en fail-open (laisse passer les requêtes).
func NewRateLimiter(enabled bool, client *redis.Client, trustedProxies []*net.IPNet) *RateLimiter {
	return &RateLimiter{
		client:         client,
		enabled:        enabled,
		trustedProxies: trustedProxies,
	}
}

// ParseTrustedProxies transforme une liste CSV d'IP/CIDR en réseaux. Les IP
// nues sont converties en /32 (IPv4) ou /128 (IPv6).
func ParseTrustedProxies(raw string) []*net.IPNet {
	var nets []*net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.Contains(part, "/") {
			if ip := net.ParseIP(part); ip != nil {
				if ip.To4() != nil {
					part += "/32"
				} else {
					part += "/128"
				}
			}
		}
		if _, n, err := net.ParseCIDR(part); err == nil {
			nets = append(nets, n)
		} else {
			log.Printf("[ratelimit] entrée TRUSTED_PROXIES ignorée (invalide) : %q", part)
		}
	}
	return nets
}

// isTrustedProxy indique si ip appartient à un des réseaux de confiance.
func isTrustedProxy(ip net.IP, trusted []*net.IPNet) bool {
	for _, n := range trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// getIP retourne l'IP client. Le X-Forwarded-For n'est lu que si la connexion
// provient d'un proxy de confiance, sinon on se base sur l'IP distante réelle
// (protection contre l'IP spoofing).
func getIP(r *http.Request, trusted []*net.IPNet) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	remoteIP := net.ParseIP(host)

	if remoteIP != nil && isTrustedProxy(remoteIP, trusted) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// Parcours de droite à gauche : on retourne la première IP qui
			// n'est pas elle-même un proxy de confiance (= vrai client).
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				candidate := strings.TrimSpace(parts[i])
				if ip := net.ParseIP(candidate); ip != nil && !isTrustedProxy(ip, trusted) {
					return candidate
				}
			}
		}
	}
	return host
}

// statusRecorder est un wrapper autour de http.ResponseWriter pour capturer le
// code de statut renvoyé par le handler.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// writeTooMany écrit une réponse 429 JSON avec un header Retry-After.
func writeTooMany(w http.ResponseWriter, retryAfterSeconds int, msg string) {
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// LimitMiddleware applique la limitation de débit et la protection brute-force.
func (rl *RateLimiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fail-open : désactivé ou Redis indisponible → on laisse passer.
		if !rl.enabled || rl.client == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ip := getIP(r, rl.trustedProxies)
		blockKey := "ratelimit:block:" + ip
		failKey := "ratelimit:fail:" + ip
		rateKey := "ratelimit:rate:" + ip

		// 1. Blocage brute-force actif ?
		if ttl, err := rl.client.PTTL(ctx, blockKey).Result(); err == nil && ttl > 0 {
			retry := int(math.Ceil(ttl.Seconds()))
			writeTooMany(w, retry, fmt.Sprintf("Too many failed attempts. Please try again in %ds.", retry))
			return
		}

		// 2. Fenêtre de débit (fail-open en cas d'erreur Redis).
		if count, err := rl.client.Incr(ctx, rateKey).Result(); err == nil {
			if count == 1 {
				rl.client.Expire(ctx, rateKey, rateWindow)
			}
			if count > rateMaxInWindow {
				writeTooMany(w, int(rateWindow.Seconds()), "Rate limit exceeded. Please wait a moment.")
				return
			}
		}

		// 3. Exécution du handler + suivi des échecs.
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		switch rec.statusCode {
		case http.StatusUnauthorized:
			failures, _ := rl.client.Incr(ctx, failKey).Result()
			rl.client.Expire(ctx, failKey, failTTL)
			log.Printf("[Warning] Failed login attempt %d from IP: %s", failures, ip)

			var blockDuration time.Duration
			switch {
			case failures >= 15:
				blockDuration = 15 * time.Minute
			case failures >= 10:
				blockDuration = 5 * time.Minute
			case failures >= 5:
				blockDuration = 1 * time.Minute
			}

			if blockDuration > 0 {
				rl.client.Set(ctx, blockKey, "1", blockDuration)
				log.Printf("[Alert] Blocking IP %s for %v due to repeated failures", ip, blockDuration)
			}

		case http.StatusOK:
			rl.client.Del(ctx, failKey, blockKey)
		}
	})
}
