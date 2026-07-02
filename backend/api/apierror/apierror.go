package apierror

import (
	"encoding/json"
	"net/http"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	IsUserError bool
}

func (e *AppError) Error() string {
	return e.Message
}

// Write sérialise une AppError en réponse JSON `{"error", "code"}`.
// Partagé par les handlers et les middlewares pour garantir que TOUTE
// réponse d'erreur du backend soit du JSON.
func Write(w http.ResponseWriter, e *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]string{
		"error": e.Message,
		"code":  e.Code,
	})
}

func NewUserError(code, message string, status int) *AppError {
	return &AppError{
		Code:        code,
		Message:     message,
		HTTPStatus:  status,
		IsUserError: true,
	}
}

func NewServerError(code, message string) *AppError {
	return &AppError{
		Code:        code,
		Message:     message,
		HTTPStatus:  http.StatusInternalServerError,
		IsUserError: false,
	}
}


func InvalidRequest(msg string) *AppError {
	return NewUserError(ErrInvalidRequest, msg, http.StatusBadRequest)
}

func InvalidCredentials() *AppError {
	return NewUserError(ErrInvalidCredentials, "Identifiant ou mot de passe incorrect.", http.StatusUnauthorized)
}

func UsernameTaken() *AppError {
	return NewUserError(ErrUsernameTaken, "Ce nom d'utilisateur est déjà pris.", http.StatusConflict)
}

func BandNameTaken() *AppError {
	return NewUserError(ErrBandNameTaken, "Ce nom de groupe existe déjà.", http.StatusConflict)
}

func ValidationFailed(msg string) *AppError {
	return NewUserError(ErrValidationFailed, msg, http.StatusBadRequest)
}

func NotFound(entity string) *AppError {
	return NewUserError(ErrNotFound, entity+" introuvable.", http.StatusNotFound)
}

func InvalidRefreshToken() *AppError {
	return NewUserError(ErrInvalidRefreshToken, "Token de rafraîchissement invalide ou expiré.", http.StatusUnauthorized)
}

func WrongCurrentPassword() *AppError {
	return NewUserError(ErrWrongCurrentPassword, "Le mot de passe actuel est incorrect.", http.StatusUnauthorized)
}

func InternalError(operation string) *AppError {
	return NewServerError(ErrInternal, "Une erreur interne s'est produite lors de: "+operation)
}

func Unauthorized(msg string) *AppError {
	return NewUserError(ErrUnauthorized, msg, http.StatusUnauthorized)
}

func Forbidden(msg string) *AppError {
	return NewUserError(ErrForbidden, msg, http.StatusForbidden)
}


const (
	ErrInvalidRequest      = "INVALID_REQUEST"
	ErrInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrUsernameTaken       = "USERNAME_TAKEN"
	ErrBandNameTaken       = "BAND_NAME_TAKEN"
	ErrValidationFailed    = "VALIDATION_FAILED"
	ErrNotFound            = "NOT_FOUND"
	ErrInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
	ErrWrongCurrentPassword = "WRONG_CURRENT_PASSWORD"
	ErrInternal            = "INTERNAL_ERROR"
	ErrUnauthorized        = "UNAUTHORIZED"
	ErrForbidden           = "FORBIDDEN"
)
