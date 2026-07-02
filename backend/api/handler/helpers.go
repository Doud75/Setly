package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"setlist/api/apierror"
	"setlist/api/middleware"
	"strconv"
)

func writeAppError(w http.ResponseWriter, appErr *apierror.AppError) {
	if !appErr.IsUserError {
		log.Printf("[ERROR][%s] %s", appErr.Code, appErr.Message)
	}
	apierror.Write(w, appErr)
}

func DecodeJSON[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, apierror.InvalidRequest("Corps de la requête invalide.")
	}
	return v, nil
}

func GetBandID(r *http.Request) (int, error) {
	bandID, ok := r.Context().Value(middleware.BandIDKey).(int)
	if !ok {
		return 0, apierror.NewServerError(apierror.ErrInternal, "Impossible d'identifier le groupe depuis le token.")
	}
	return bandID, nil
}

func GetOptionalBandID(r *http.Request) (int, bool) {
	bandID, ok := r.Context().Value(middleware.BandIDKey).(int)
	return bandID, ok
}

func GetUserID(r *http.Request) (int, error) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		return 0, apierror.NewServerError(apierror.ErrInternal, "Impossible d'identifier l'utilisateur depuis le token.")
	}
	return userID, nil
}

func GetIntParam(r *http.Request, key string) (int, error) {
	val := r.PathValue(key)
	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, apierror.InvalidRequest("Paramètre invalide : " + key + ".")
	}
	return id, nil
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondOK(w http.ResponseWriter, data any) {
	RespondJSON(w, http.StatusOK, data)
}

func RespondCreated(w http.ResponseWriter, data any) {
	RespondJSON(w, http.StatusCreated, data)
}

func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func asAppError(err error) *apierror.AppError {
	var appErr *apierror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
