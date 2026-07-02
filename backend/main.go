package main

import (
	"fmt"
	"log"
	"net/http"
	"setlist/api/handler"
	"setlist/api/middleware"
	"setlist/api/repository"
	"setlist/api/service"
	"setlist/cache"
	"setlist/config"
	"setlist/db"
)

func main() {
	cfg := config.Load()
	dbPool := db.NewConnection(cfg.DatabaseURL)
	defer dbPool.Close()

	redisClient := cache.NewClient(cfg.RedisURL)

	userRepo := &repository.PgUserRepository{DB: dbPool}
	refreshTokenRepo := &repository.PgRefreshTokenRepository{DB: dbPool}
	userService := service.UserService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		JWTSecret:        cfg.JWTSecret,
	}
	authService := service.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		JWTSecret:        cfg.JWTSecret,
	}
	userHandler := handler.UserHandler{UserService: userService}
	authHandler := handler.AuthHandler{AuthService: authService}
	bandHandler := handler.BandHandler{UserService: userService}

	interludeRepo := &repository.PgInterludeRepository{DB: dbPool}
	interludeService := service.InterludeService{InterludeRepo: interludeRepo}
	interludeHandler := handler.InterludeHandler{InterludeService: interludeService}

	infoRepo := &repository.PgInfoRepository{DB: dbPool}
	infoHandler := handler.InfoHandler{InfoRepo: infoRepo, UserRepo: userRepo, Cache: redisClient}

	songRepo := &repository.PgSongRepository{DB: dbPool}
	songService := service.SongService{SongRepo: songRepo, Cache: redisClient}
	songHandler := handler.SongHandler{SongService: songService}

	setlistRepo := &repository.PgSetlistRepository{DB: dbPool}
	setlistService := service.SetlistService{SetlistRepo: setlistRepo, InterludeRepo: interludeRepo, SongRepo: songRepo, Cache: redisClient}
	setlistHandler := handler.SetlistHandler{SetlistService: setlistService}

	invitationRepo := &repository.PgInvitationRepository{DB: dbPool}
	invitationService := service.InvitationService{InvitationRepo: invitationRepo, UserRepo: userRepo}
	invitationHandler := handler.InvitationHandler{InvitationService: invitationService}

	authMiddleware := middleware.JWTAuth(cfg.JWTSecret, userRepo)
	authMiddlewareUserOnly := middleware.JWTAuthUserOnly(cfg.JWTSecret)
	adminMiddleware := middleware.AdminOnly(userRepo)
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimitEnabled,
		redisClient,
		middleware.ParseTrustedProxies(cfg.TrustedProxies),
	)

	mux := http.NewServeMux()

	mux.Handle("/api/auth/login", rateLimiter.LimitMiddleware(handler.Wrap(userHandler.Login)))
	mux.Handle("/api/auth/signup", rateLimiter.LimitMiddleware(handler.Wrap(userHandler.Signup)))
	mux.Handle("/api/auth/refresh", handler.Wrap(authHandler.RefreshToken))
	mux.Handle("/api/auth/logout", handler.Wrap(authHandler.Logout))
	mux.Handle("PUT /api/user/password", authMiddlewareUserOnly(handler.Wrap(userHandler.UpdatePassword)))
	mux.Handle("GET /api/user/info", authMiddlewareUserOnly(handler.Wrap(infoHandler.GetCurrentUserInfo)))
	mux.Handle("GET /api/user/bands", authMiddlewareUserOnly(handler.Wrap(bandHandler.GetUserBands)))
	mux.Handle("PUT /api/user/default-band", authMiddlewareUserOnly(handler.Wrap(bandHandler.SetDefaultBand)))

	mux.Handle("POST /api/bands", authMiddlewareUserOnly(handler.Wrap(bandHandler.CreateBand)))
	mux.Handle("GET /api/bands/{bandId}/members", authMiddleware(handler.Wrap(bandHandler.GetMembers)))
	mux.Handle("POST /api/bands/{bandId}/members", authMiddleware(adminMiddleware(handler.Wrap(bandHandler.InviteMember))))
	mux.Handle("PUT /api/bands/{bandId}/members/{userId}/role", authMiddleware(adminMiddleware(handler.Wrap(bandHandler.UpdateMemberRole))))
	mux.Handle("DELETE /api/bands/{bandId}/members/{userId}", authMiddleware(adminMiddleware(handler.Wrap(bandHandler.RemoveMember))))
	mux.Handle("DELETE /api/bands/{bandId}/members/me", authMiddlewareUserOnly(handler.Wrap(bandHandler.LeaveBand)))

	mux.Handle("POST /api/bands/{bandId}/invitations", authMiddleware(adminMiddleware(handler.Wrap(invitationHandler.CreateInvitation))))
	mux.Handle("GET /api/invitations/{token}", handler.Wrap(invitationHandler.GetInvitation))
	mux.Handle("POST /api/invitations/{token}/accept", authMiddlewareUserOnly(handler.Wrap(invitationHandler.AcceptInvitation)))

	mux.Handle("POST /api/setlist", authMiddleware(handler.Wrap(setlistHandler.CreateSetlist)))
	mux.Handle("GET /api/setlist", authMiddleware(handler.Wrap(setlistHandler.GetSetlists)))
	mux.Handle("GET /api/setlist/{id}", authMiddleware(handler.Wrap(setlistHandler.GetSetlistDetails)))
	mux.Handle("PUT /api/setlist/{id}", authMiddleware(handler.Wrap(setlistHandler.UpdateSetlist)))
	mux.Handle("DELETE /api/setlist/{id}", authMiddleware(adminMiddleware(handler.Wrap(setlistHandler.DeleteSetlist))))

	mux.Handle("POST /api/setlist/{id}/duplicate", authMiddleware(handler.Wrap(setlistHandler.DuplicateSetlist)))
	mux.Handle("POST /api/setlist/{id}/items", authMiddleware(handler.Wrap(setlistHandler.AddItem)))
	mux.Handle("PUT /api/setlist/{id}/items/order", authMiddleware(handler.Wrap(setlistHandler.UpdateItemOrder)))
	mux.Handle("PUT /api/setlist/item/{itemId}", authMiddleware(handler.Wrap(setlistHandler.UpdateItem)))
	mux.Handle("DELETE /api/setlist/item/{itemId}", authMiddleware(adminMiddleware(handler.Wrap(setlistHandler.DeleteItem))))

	mux.Handle("POST /api/song", authMiddleware(handler.Wrap(songHandler.CreateSong)))
	mux.Handle("GET /api/song", authMiddleware(handler.Wrap(songHandler.GetSongs)))
	mux.Handle("GET /api/song/{id}", authMiddleware(handler.Wrap(songHandler.GetSong)))
	mux.Handle("PUT /api/song/{id}", authMiddleware(handler.Wrap(songHandler.UpdateSong)))
	mux.Handle("DELETE /api/song/{id}", authMiddleware(adminMiddleware(handler.Wrap(songHandler.DeleteSong))))

	mux.Handle("POST /api/interlude", authMiddleware(handler.Wrap(interludeHandler.CreateInterlude)))
	mux.Handle("GET /api/interlude", authMiddleware(handler.Wrap(interludeHandler.GetInterludes)))
	mux.Handle("PUT /api/interlude/{id}", authMiddleware(handler.Wrap(interludeHandler.UpdateInterlude)))

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	port := "8089"
	address := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("Backend server starting on %s\n", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
