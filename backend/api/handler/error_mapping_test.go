package handler

import (
	"errors"
	"net/http"
	"testing"

	"setlist/api/apierror"
	"setlist/api/repository"
	"setlist/api/service"
)

func assertAppError(t *testing.T, err error, wantStatus int, wantCode string) {
	t.Helper()
	var appErr *apierror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apierror.AppError, got %T (%v)", err, err)
	}
	if appErr.HTTPStatus != wantStatus {
		t.Errorf("expected HTTP status %d, got %d", wantStatus, appErr.HTTPStatus)
	}
	if appErr.Code != wantCode {
		t.Errorf("expected code %q, got %q", wantCode, appErr.Code)
	}
}

func TestMapSetlistError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"setlist not found -> 404", service.ErrSetlistNotFound, http.StatusNotFound, apierror.ErrNotFound},
		{"item not found -> 404", service.ErrItemNotFound, http.StatusNotFound, apierror.ErrNotFound},
		{"invalid item type -> 400", service.ErrInvalidItemType, http.StatusBadRequest, apierror.ErrInvalidRequest},
		{"name required -> 400", service.ErrSetlistNameRequired, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"invalid color -> 400", service.ErrInvalidColor, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"unexpected error -> 500", errors.New("db down"), http.StatusInternalServerError, apierror.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertAppError(t, mapSetlistError(tc.err, "test"), tc.wantStatus, tc.wantCode)
		})
	}
}

func TestMapInterludeError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"interlude not found -> 404", service.ErrInterludeNotFound, http.StatusNotFound, apierror.ErrNotFound},
		{"title required -> 400", service.ErrInterludeTitleRequired, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"unexpected error -> 500", errors.New("db down"), http.StatusInternalServerError, apierror.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertAppError(t, mapInterludeError(tc.err, "test"), tc.wantStatus, tc.wantCode)
		})
	}
}

func TestMapUserError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"duplicate username -> 409", repository.ErrDuplicateUsername, http.StatusConflict, apierror.ErrUsernameTaken},
		{"invalid credentials -> 401", service.ErrInvalidCredentials, http.StatusUnauthorized, apierror.ErrInvalidCredentials},
		{"wrong current password -> 401", service.ErrWrongCurrentPassword, http.StatusUnauthorized, apierror.ErrWrongCurrentPassword},
		{"user not found -> 404", service.ErrUserNotFound, http.StatusNotFound, apierror.ErrNotFound},
		{"validation error -> 400", &service.ValidationError{Msg: "le mot de passe est trop court"}, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"unexpected error -> 500", errors.New("db down"), http.StatusInternalServerError, apierror.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertAppError(t, mapUserError(tc.err, "test"), tc.wantStatus, tc.wantCode)
		})
	}
}

func TestMapBandError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"duplicate username -> 409", repository.ErrDuplicateUsername, http.StatusConflict, apierror.ErrUsernameTaken},
		{"already member -> 409", service.ErrAlreadyBandMember, http.StatusConflict, apierror.ErrInvalidRequest},
		{"last admin -> 409", service.ErrLastAdmin, http.StatusConflict, apierror.ErrInvalidRequest},
		{"cannot demote last admin -> 409", service.ErrCannotDemoteLastAdmin, http.StatusConflict, apierror.ErrInvalidRequest},
		{"invalid role -> 400", service.ErrInvalidRole, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"not a member -> 400", service.ErrNotBandMember, http.StatusBadRequest, apierror.ErrInvalidRequest},
		{"band name required -> 400", service.ErrBandNameRequired, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"password required -> 400", service.ErrUserPasswordRequired, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"band not found -> 404", service.ErrBandNotFoundOrNotMember, http.StatusNotFound, apierror.ErrNotFound},
		{"validation error -> 400", &service.ValidationError{Msg: "le nom d'utilisateur est requis"}, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"unexpected error -> 500", errors.New("db down"), http.StatusInternalServerError, apierror.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertAppError(t, mapBandError(tc.err, "test"), tc.wantStatus, tc.wantCode)
		})
	}
}

func TestMapInvitationError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"invitation not found -> 404", repository.ErrInvitationNotFound, http.StatusNotFound, apierror.ErrNotFound},
		{"invitation expired -> 410", repository.ErrInvitationExpired, http.StatusGone, apierror.ErrInvalidRequest},
		{"already member -> 409", service.ErrAlreadyBandMember, http.StatusConflict, apierror.ErrInvalidRequest},
		{"unexpected error -> 500", errors.New("db down"), http.StatusInternalServerError, apierror.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertAppError(t, mapInvitationError(tc.err, "test"), tc.wantStatus, tc.wantCode)
		})
	}
}

func TestMapSongError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"song not found -> 404", service.ErrSongNotFound, http.StatusNotFound, apierror.ErrNotFound},
		{"title required -> 400", service.ErrSongTitleRequired, http.StatusBadRequest, apierror.ErrValidationFailed},
		{"unexpected error -> 500", errors.New("db down"), http.StatusInternalServerError, apierror.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertAppError(t, mapSongError(tc.err, "test"), tc.wantStatus, tc.wantCode)
		})
	}
}
