package handler

import (
	"errors"
	"net/http"
	"setlist/api/apierror"
	"setlist/api/repository"
	"setlist/api/service"
)

type UserHandler struct {
	UserService service.UserService
}

// mapUserError translates the user service's sentinel errors into typed API
// errors; anything else is reported as an internal error on the operation.
func mapUserError(err error, operation string) error {
	var ve *service.ValidationError
	switch {
	case errors.Is(err, repository.ErrDuplicateUsername):
		return apierror.UsernameTaken()
	case errors.Is(err, service.ErrInvalidCredentials):
		return apierror.InvalidCredentials()
	case errors.Is(err, service.ErrWrongCurrentPassword):
		return apierror.WrongCurrentPassword()
	case errors.Is(err, service.ErrUserNotFound):
		return apierror.NotFound("Utilisateur")
	case errors.As(err, &ve):
		return apierror.ValidationFailed(ve.Msg)
	default:
		if appErr := asAppError(err); appErr != nil {
			return appErr
		}
		return apierror.InternalError(operation)
	}
}

func (h UserHandler) Signup(w http.ResponseWriter, r *http.Request) error {
	payload, err := DecodeJSON[service.AuthPayload](r)
	if err != nil {
		return err
	}

	response, err := h.UserService.Signup(r.Context(), payload)
	if err != nil {
		return mapUserError(err, "inscription")
	}

	RespondCreated(w, response)
	return nil
}

func (h UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	payload, err := DecodeJSON[service.LoginPayload](r)
	if err != nil {
		return err
	}

	response, err := h.UserService.Login(r.Context(), payload)
	if err != nil {
		return mapUserError(err, "connexion")
	}

	RespondOK(w, response)
	return nil
}

func (h UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) error {
	userID, err := GetUserID(r)
	if err != nil {
		return err
	}

	payload, err := DecodeJSON[service.UpdatePasswordPayload](r)
	if err != nil {
		return err
	}

	if payload.NewPassword == "" {
		return apierror.InvalidRequest("Le nouveau mot de passe ne peut pas être vide.")
	}

	if err := h.UserService.UpdatePassword(r.Context(), userID, payload); err != nil {
		return mapUserError(err, "mise à jour du mot de passe")
	}

	RespondOK(w, map[string]string{"message": "Mot de passe mis à jour avec succès."})
	return nil
}
