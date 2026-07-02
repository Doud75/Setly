package handler

import (
	"errors"
	"net/http"
	"setlist/api/apierror"
	"setlist/api/model"
	"setlist/api/repository"
	"setlist/api/service"
)

type BandHandler struct {
	UserService service.UserService
}

// mapBandError translates the user service's band-related sentinel errors
// into typed API errors; anything else is reported as an internal error.
func mapBandError(err error, operation string) error {
	var ve *service.ValidationError
	switch {
	case errors.Is(err, repository.ErrDuplicateUsername):
		return apierror.UsernameTaken()
	case errors.Is(err, service.ErrAlreadyBandMember):
		return apierror.NewUserError(apierror.ErrInvalidRequest, "Cet utilisateur est déjà membre du groupe.", http.StatusConflict)
	case errors.Is(err, service.ErrLastAdmin):
		return apierror.NewUserError(
			apierror.ErrInvalidRequest,
			"Impossible de quitter : vous êtes le dernier administrateur. Supprimez le groupe ou promouvez un autre membre.",
			http.StatusConflict,
		)
	case errors.Is(err, service.ErrNotBandMember):
		return apierror.InvalidRequest("Vous n'êtes pas membre de ce groupe.")
	case errors.Is(err, service.ErrInvalidRole):
		return apierror.ValidationFailed("Rôle invalide.")
	case errors.Is(err, service.ErrCannotDemoteLastAdmin):
		return apierror.NewUserError(
			apierror.ErrInvalidRequest,
			"Impossible de rétrograder : c'est le dernier administrateur du groupe.",
			http.StatusConflict,
		)
	case errors.Is(err, service.ErrBandNameRequired):
		return apierror.ValidationFailed("Le nom du groupe est requis.")
	case errors.Is(err, service.ErrUserPasswordRequired):
		return apierror.ValidationFailed("Utilisateur introuvable : un mot de passe est requis pour le créer.")
	case errors.Is(err, service.ErrBandNotFoundOrNotMember):
		return apierror.NotFound("Groupe")
	case errors.As(err, &ve):
		return apierror.ValidationFailed(ve.Msg)
	default:
		if appErr := asAppError(err); appErr != nil {
			return appErr
		}
		return apierror.InternalError(operation)
	}
}

func (h BandHandler) GetMembers(w http.ResponseWriter, r *http.Request) error {
	bandID, err := GetBandID(r)
	if err != nil {
		return err
	}

	members, err := h.UserService.GetBandMembers(r.Context(), bandID)
	if err != nil {
		return apierror.InternalError("récupération des membres")
	}
	if members == nil {
		members = make([]model.BandMember, 0)
	}

	RespondOK(w, members)
	return nil
}

func (h BandHandler) InviteMember(w http.ResponseWriter, r *http.Request) error {
	bandID, err := GetBandID(r)
	if err != nil {
		return err
	}

	payload, err := DecodeJSON[service.InviteMemberPayload](r)
	if err != nil {
		return err
	}

	user, err := h.UserService.InviteMember(r.Context(), bandID, payload)
	if err != nil {
		return mapBandError(err, "invitation d'un membre")
	}

	RespondCreated(w, user)
	return nil
}

func (h BandHandler) RemoveMember(w http.ResponseWriter, r *http.Request) error {
	bandID, err := GetBandID(r)
	if err != nil {
		return err
	}

	userID, err := GetIntParam(r, "userId")
	if err != nil {
		return apierror.InvalidRequest("Identifiant utilisateur invalide.")
	}

	if err := h.UserService.RemoveMember(r.Context(), bandID, userID); err != nil {
		return mapBandError(err, "retrait d'un membre")
	}

	RespondNoContent(w)
	return nil
}

type UpdateMemberRolePayload struct {
	Role string `json:"role"`
}

func (h BandHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) error {
	bandID, err := GetBandID(r)
	if err != nil {
		return err
	}

	userID, err := GetIntParam(r, "userId")
	if err != nil {
		return apierror.InvalidRequest("Identifiant utilisateur invalide.")
	}

	payload, err := DecodeJSON[UpdateMemberRolePayload](r)
	if err != nil {
		return err
	}

	if err := h.UserService.ChangeMemberRole(r.Context(), bandID, userID, payload.Role); err != nil {
		return mapBandError(err, "changement de rôle d'un membre")
	}

	RespondOK(w, map[string]string{"role": payload.Role})
	return nil
}

func (h BandHandler) GetUserBands(w http.ResponseWriter, r *http.Request) error {
	userID, err := GetUserID(r)
	if err != nil {
		return err
	}

	bands, err := h.UserService.GetUserBands(r.Context(), userID)
	if err != nil {
		return apierror.InternalError("récupération des groupes")
	}

	RespondOK(w, bands)
	return nil
}

type CreateBandPayload struct {
	Name string `json:"name"`
}

func (h BandHandler) CreateBand(w http.ResponseWriter, r *http.Request) error {
	userID, err := GetUserID(r)
	if err != nil {
		return err
	}

	payload, err := DecodeJSON[CreateBandPayload](r)
	if err != nil {
		return err
	}

	band, err := h.UserService.CreateBand(r.Context(), payload.Name, userID)
	if err != nil {
		return mapBandError(err, "création du groupe")
	}

	RespondCreated(w, band)
	return nil
}

func (h BandHandler) LeaveBand(w http.ResponseWriter, r *http.Request) error {
	userID, err := GetUserID(r)
	if err != nil {
		return err
	}

	bandID, err := GetBandID(r)
	if err != nil {
		return err
	}

	if err := h.UserService.LeaveBand(r.Context(), userID, bandID); err != nil {
		return mapBandError(err, "départ du groupe")
	}

	RespondNoContent(w)
	return nil
}

type SetDefaultBandPayload struct {
	BandID int `json:"band_id"`
}

func (h BandHandler) SetDefaultBand(w http.ResponseWriter, r *http.Request) error {
	userID, err := GetUserID(r)
	if err != nil {
		return err
	}

	payload, err := DecodeJSON[SetDefaultBandPayload](r)
	if err != nil {
		return err
	}

	if err := h.UserService.SetDefaultBand(r.Context(), userID, payload.BandID); err != nil {
		return mapBandError(err, "définition du groupe par défaut")
	}

	RespondNoContent(w)
	return nil
}
