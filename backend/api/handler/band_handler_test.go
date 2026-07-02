package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"setlist/api/middleware"
	"setlist/api/repository/mocks"
	"setlist/api/service"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func newBandHandler(mockUserRepo *mocks.MockUserRepository) BandHandler {
	return BandHandler{UserService: service.UserService{UserRepo: mockUserRepo}}
}

func updateRoleRequest(bandID int, userID string, role string) *http.Request {
	body, _ := json.Marshal(UpdateMemberRolePayload{Role: role})
	req := httptest.NewRequest(http.MethodPut, "/api/bands/1/members/"+userID+"/role", bytes.NewReader(body))
	req.SetPathValue("userId", userID)
	ctx := context.WithValue(req.Context(), middleware.BandIDKey, bandID)
	return req.WithContext(ctx)
}

func TestBandHandler_UpdateMemberRole(t *testing.T) {
	t.Run("PromoteMember_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		h := newBandHandler(mockUserRepo)

		mockUserRepo.EXPECT().GetUserRoleInBand(gomock.Any(), 42, 1).Return("member", nil)
		mockUserRepo.EXPECT().UpdateUserRoleInBand(gomock.Any(), 42, 1, "admin").Return(nil)

		w := httptest.NewRecorder()
		if err := h.UpdateMemberRole(w, updateRoleRequest(1, "42", "admin")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["role"] != "admin" {
			t.Errorf("expected role admin in response, got %q", resp["role"])
		}
	})

	t.Run("DemoteLastAdmin_Conflict", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		h := newBandHandler(mockUserRepo)

		mockUserRepo.EXPECT().GetUserRoleInBand(gomock.Any(), 42, 1).Return("admin", nil)
		mockUserRepo.EXPECT().GetAdminCountInBand(gomock.Any(), 1).Return(1, nil)

		w := httptest.NewRecorder()
		err := h.UpdateMemberRole(w, updateRoleRequest(1, "42", "member"))

		appErr := asAppError(err)
		if appErr == nil {
			t.Fatalf("expected an AppError, got %v", err)
		}
		if appErr.HTTPStatus != http.StatusConflict {
			t.Errorf("expected status 409, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("InvalidRole_ValidationError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		h := newBandHandler(mockUserRepo)

		// rôle rejeté avant tout accès au repo -> aucune EXPECT.
		w := httptest.NewRecorder()
		err := h.UpdateMemberRole(w, updateRoleRequest(1, "42", "superadmin"))

		appErr := asAppError(err)
		if appErr == nil {
			t.Fatalf("expected an AppError, got %v", err)
		}
		if appErr.HTTPStatus != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", appErr.HTTPStatus)
		}
	})

	t.Run("NotAMember_Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		h := newBandHandler(mockUserRepo)

		mockUserRepo.EXPECT().GetUserRoleInBand(gomock.Any(), 42, 1).Return("", pgx.ErrNoRows)

		w := httptest.NewRecorder()
		if err := h.UpdateMemberRole(w, updateRoleRequest(1, "42", "admin")); err == nil {
			t.Fatal("expected error for non-member, got nil")
		}
	})
}
