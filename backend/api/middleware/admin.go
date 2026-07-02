package middleware

import (
	"net/http"
	"setlist/api/apierror"
	"setlist/api/repository"
)

func AdminOnly(userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(int)
			if !ok {
				apierror.Write(w, apierror.NewServerError(apierror.ErrInternal, "Utilisateur non identifié."))
				return
			}
			bandID, ok := r.Context().Value(BandIDKey).(int)
			if !ok {
				apierror.Write(w, apierror.NewServerError(apierror.ErrInternal, "Groupe non identifié."))
				return
			}

			role, err := userRepo.GetUserRoleInBand(r.Context(), userID, bandID)
			if err != nil {
				apierror.Write(w, apierror.NewServerError(apierror.ErrInternal, "Impossible de vérifier le rôle de l'utilisateur."))
				return
			}

			if role != "admin" {
				apierror.Write(w, apierror.Forbidden("Accès réservé aux administrateurs."))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
