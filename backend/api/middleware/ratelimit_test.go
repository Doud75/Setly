package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// --- getIP (5.2) -----------------------------------------------------------

func TestGetIP(t *testing.T) {
	trusted := ParseTrustedProxies("127.0.0.1/32,::1/128,172.16.0.0/12")

	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{
			name:       "proxy de confiance + XFF → vraie IP client",
			remoteAddr: "172.19.0.7:5000",
			xff:        "88.1.2.3",
			want:       "88.1.2.3",
		},
		{
			name:       "remote non trusté → XFF ignoré (anti-spoofing)",
			remoteAddr: "88.9.9.9:5000",
			xff:        "1.2.3.4",
			want:       "88.9.9.9",
		},
		{
			name:       "XFF multi-valeurs → premier non-proxy en partant de la droite",
			remoteAddr: "172.19.0.7:5000",
			xff:        "88.1.2.3, 172.19.0.8",
			want:       "88.1.2.3",
		},
		{
			name:       "proxy de confiance sans XFF → RemoteAddr",
			remoteAddr: "172.19.0.7:5000",
			xff:        "",
			want:       "172.19.0.7",
		},
		{
			name:       "XFF entièrement composé de proxies → RemoteAddr",
			remoteAddr: "172.19.0.7:5000",
			xff:        "172.19.0.9",
			want:       "172.19.0.7",
		},
		{
			name:       "IPv6 sans XFF → host correctement extrait",
			remoteAddr: "[::1]:1234",
			xff:        "",
			want:       "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/auth/login", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if got := getIP(req, trusted); got != tt.want {
				t.Errorf("getIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Rate limiter sur Redis (5.1) ------------------------------------------

// newTestLimiter démarre un miniredis et retourne un RateLimiter branché dessus.
func newTestLimiter(t *testing.T) (*RateLimiter, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	// trustedProxies nil → getIP renvoie toujours le RemoteAddr.
	return NewRateLimiter(true, client, nil), mr
}

func TestRateLimiter_RateWindow(t *testing.T) {
	rl, _ := newTestLimiter(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.LimitMiddleware(next)

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.2:1234"

	for i := 0; i < rateMaxInWindow; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("requête %d : attendu 200, obtenu %d", i+1, rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("requête %d : attendu 429, obtenu %d", rateMaxInWindow+1, rr.Code)
	}
}

func TestRateLimiter_BruteForceBlock(t *testing.T) {
	rl, mr := newTestLimiter(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	handler := rl.LimitMiddleware(next)

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	// 5 échecs successifs. On avance le temps entre chaque pour que la fenêtre
	// de débit (12s) ne déclenche pas avant d'atteindre le seuil brute-force.
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("échec %d : attendu 401, obtenu %d", i+1, rr.Code)
		}
		mr.FastForward(rateWindow)
	}

	// L'IP doit désormais être bloquée : la requête suivante est rejetée en 429
	// avant même d'atteindre le handler.
	rrCheck := httptest.NewRecorder()
	handler.ServeHTTP(rrCheck, req)
	if rrCheck.Code != http.StatusTooManyRequests {
		t.Fatalf("après 5 échecs : attendu 429, obtenu %d", rrCheck.Code)
	}

	var resp map[string]string
	json.NewDecoder(rrCheck.Body).Decode(&resp)
	if _, ok := resp["error"]; !ok {
		t.Error("message d'erreur JSON manquant dans la réponse 429")
	}
}

func TestRateLimiter_SuccessResetsFailures(t *testing.T) {
	rl, mr := newTestLimiter(t)

	status := http.StatusUnauthorized
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
	handler := rl.LimitMiddleware(next)

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.3:1234"

	// 3 échecs (sous le seuil de 5).
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		mr.FastForward(rateWindow)
	}

	// Un succès doit réinitialiser le compteur d'échecs.
	status = http.StatusOK
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("attendu 200, obtenu %d", rr.Code)
	}

	if mr.Exists("ratelimit:fail:192.168.1.3") {
		t.Error("le compteur d'échecs devrait être supprimé après un succès")
	}
}

func TestRateLimiter_FailOpenWhenNoRedis(t *testing.T) {
	// client nil → fail-open : la requête passe toujours.
	rl := NewRateLimiter(true, nil, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.LimitMiddleware(next)

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.4:1234"

	for i := 0; i < 20; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("fail-open attendu, obtenu %d à la requête %d", rr.Code, i+1)
		}
	}
}
