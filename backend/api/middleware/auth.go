package middleware

import (
	"context"
	"net/http"
	"setlist/api/apierror"
	"setlist/api/repository"
	"setlist/auth"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
	UserIDKey ContextKey = "userID"
	BandIDKey ContextKey = "bandID"
)

func JWTAuth(jwtSecret string, userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claimsVal, ok := validateJWT(w, r, jwtSecret)
			if !ok {
				return
			}

			bandIDStr := r.Header.Get("X-Band-ID")
			if bandIDStr == "" {
				apierror.Write(w, apierror.InvalidRequest("En-tête X-Band-ID manquant."))
				return
			}
			bandID, err := strconv.Atoi(bandIDStr)
			if err != nil {
				apierror.Write(w, apierror.InvalidRequest("En-tête X-Band-ID invalide."))
				return
			}

			isMember, err := userRepo.IsUserInBand(r.Context(), claimsVal.UserID, bandID)
			if err != nil {
				apierror.Write(w, apierror.NewServerError(apierror.ErrInternal, "Erreur lors de la vérification de l'appartenance au groupe."))
				return
			}
			if !isMember {
				apierror.Write(w, apierror.Forbidden("Vous n'êtes pas membre de ce groupe."))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claimsVal.UserID)
			ctx = context.WithValue(ctx, BandIDKey, bandID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JWTAuthUserOnly(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claimsVal, ok := validateJWT(w, r, jwtSecret)
			if !ok {
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claimsVal.UserID)

			if bandIDStr := r.Header.Get("X-Band-ID"); bandIDStr != "" {
				if bandID, err := strconv.Atoi(bandIDStr); err == nil {
					ctx = context.WithValue(ctx, BandIDKey, bandID)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateJWT(w http.ResponseWriter, r *http.Request, jwtSecret string) (*auth.JWTClaims, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		apierror.Write(w, apierror.Unauthorized("En-tête d'autorisation manquant."))
		return nil, false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		apierror.Write(w, apierror.Unauthorized("Format de token invalide."))
		return nil, false
	}

	claims := &auth.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		apierror.Write(w, apierror.Unauthorized("Token invalide."))
		return nil, false
	}

	return claims, true
}
