package service

import (
	"context"
	"errors"
	"testing"

	"setlist/api/model"
	"setlist/api/repository/mocks"
	"setlist/auth"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func TestUserService_Signup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)

	jwtSecret := "testsecret"
	svc := UserService{
		UserRepo:         mockUserRepo,
		RefreshTokenRepo: mockRefreshTokenRepo,
		JWTSecret:        jwtSecret,
	}

	ctx := context.Background()
	payload := AuthPayload{
		Username: "testuser",
		Password: "Password123!",
	}

	t.Run("Success", func(t *testing.T) {
		expectedUser := model.User{ID: 1, Username: "testuser"}

		mockUserRepo.EXPECT().
			CreateUser(ctx, payload.Username, gomock.Any()).
			Return(expectedUser, nil)
		mockRefreshTokenRepo.EXPECT().
			ReplaceUserRefreshToken(ctx, expectedUser.ID, gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := svc.Signup(ctx, payload)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("response should not be nil")
		}
		if resp.Token == "" {
			t.Error("token should not be empty")
		}
		if resp.RefreshToken == "" {
			t.Error("refresh token should not be empty")
		}
		if len(resp.Bands) != 0 {
			t.Errorf("expected 0 bands, got %d", len(resp.Bands))
		}
	})

	t.Run("RepoError", func(t *testing.T) {
		repoErr := errors.New("database error")

		mockUserRepo.EXPECT().
			CreateUser(ctx, payload.Username, gomock.Any()).
			Return(model.User{}, repoErr)

		resp, err := svc.Signup(ctx, payload)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if resp != nil {
			t.Fatal("response should be nil on error")
		}
		if !errors.Is(err, repoErr) {
			t.Errorf("expected error %v, got %v", repoErr, err)
		}
	})
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)

	svc := UserService{
		UserRepo:         mockUserRepo,
		RefreshTokenRepo: mockRefreshTokenRepo,
		JWTSecret:        "testsecret",
	}

	ctx := context.Background()
	payload := LoginPayload{Username: "testuser", Password: "Password123!"}

	t.Run("LoginWithBands", func(t *testing.T) {
		hashedPw, _ := hashPasswordForTest("Password123!")
		expectedUser := model.User{ID: 1, Username: "testuser", PasswordHash: hashedPw}
		expectedBands := []model.BandWithRole{{ID: 1, Name: "Test Band", Role: "admin", IsDefault: true}}

		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, payload.Username).
			Return(expectedUser, nil)
		mockUserRepo.EXPECT().
			FindBandsWithRoleByUserID(ctx, expectedUser.ID).
			Return(expectedBands, nil)
		mockRefreshTokenRepo.EXPECT().
			ReplaceUserRefreshToken(ctx, expectedUser.ID, gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := svc.Login(ctx, payload)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Token == "" {
			t.Error("token should not be empty")
		}
		if len(resp.Bands) != 1 {
			t.Errorf("expected 1 band, got %d", len(resp.Bands))
		}
		if resp.DefaultBandID == nil || *resp.DefaultBandID != 1 {
			t.Errorf("expected DefaultBandID=1, got %v", resp.DefaultBandID)
		}
	})

	t.Run("LoginWithNoBands_OrphanUser", func(t *testing.T) {
		hashedPw, _ := hashPasswordForTest("Password123!")
		expectedUser := model.User{ID: 2, Username: "testuser", PasswordHash: hashedPw}

		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, payload.Username).
			Return(expectedUser, nil)
		mockUserRepo.EXPECT().
			FindBandsWithRoleByUserID(ctx, expectedUser.ID).
			Return([]model.BandWithRole{}, nil)

		mockRefreshTokenRepo.EXPECT().
			ReplaceUserRefreshToken(ctx, expectedUser.ID, gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := svc.Login(ctx, payload)

		if err != nil {
			t.Fatalf("orphan user login should succeed, got error: %v", err)
		}
		if resp.Token == "" {
			t.Error("token should not be empty")
		}
		if len(resp.Bands) != 0 {
			t.Errorf("expected 0 bands for orphan user, got %d", len(resp.Bands))
		}
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		hashedPw, _ := hashPasswordForTest("Password123!")
		expectedUser := model.User{ID: 3, Username: "testuser", PasswordHash: hashedPw}

		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, payload.Username).
			Return(expectedUser, nil)

		resp, err := svc.Login(ctx, LoginPayload{Username: "testuser", Password: "wrongpassword"})

		if err == nil {
			t.Fatal("expected error for invalid password, got nil")
		}
		if resp != nil {
			t.Fatal("response should be nil on error")
		}
	})
}

func hashPasswordForTest(password string) (string, error) {
	return auth.HashPassword(password)
}

func TestUserService_CreateBand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)

	svc := UserService{
		UserRepo:         mockUserRepo,
		RefreshTokenRepo: mockRefreshTokenRepo,
		JWTSecret:        "testsecret",
	}

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expectedBand := model.Band{ID: 10, Name: "New Band"}

		mockUserRepo.EXPECT().
			CreateBand(ctx, "New Band", 1).
			Return(expectedBand, nil)
		mockUserRepo.EXPECT().
			FindBandsWithRoleByUserID(ctx, 1).
			Return([]model.BandWithRole{{ID: 10, Name: "New Band", IsDefault: false}}, nil)
		mockUserRepo.EXPECT().
			SetDefaultBand(ctx, 1, 10).
			Return(nil)

		band, err := svc.CreateBand(ctx, "New Band", 1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if band.ID != expectedBand.ID {
			t.Errorf("expected band ID %d, got %d", expectedBand.ID, band.ID)
		}
		if band.Name != expectedBand.Name {
			t.Errorf("expected band name %s, got %s", expectedBand.Name, band.Name)
		}
	})

	t.Run("EmptyName", func(t *testing.T) {
		band, err := svc.CreateBand(ctx, "", 1)

		if !errors.Is(err, ErrBandNameRequired) {
			t.Fatalf("expected ErrBandNameRequired, got %v", err)
		}
		if band.ID != 0 {
			t.Errorf("expected zero Band on error, got ID=%d", band.ID)
		}
	})
}

func TestUserService_LeaveBand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := UserService{UserRepo: mockUserRepo, JWTSecret: "testsecret"}
	ctx := context.Background()

	const userID, bandID = 10, 5

	t.Run("MemberLeaves_Success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, userID, bandID).Return("member", nil)
		mockUserRepo.EXPECT().FindBandsWithRoleByUserID(ctx, userID).Return([]model.BandWithRole{
			{ID: bandID, Name: "Band", IsDefault: false},
		}, nil)
		mockUserRepo.EXPECT().RemoveUserFromBand(ctx, bandID, userID).Return(nil)

		if err := svc.LeaveBand(ctx, userID, bandID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("LastAdmin_Blocked", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, userID, bandID).Return("admin", nil)
		mockUserRepo.EXPECT().GetAdminCountInBand(ctx, bandID).Return(1, nil)

		err := svc.LeaveBand(ctx, userID, bandID)
		if !errors.Is(err, ErrLastAdmin) {
			t.Errorf("expected ErrLastAdmin, got: %v", err)
		}
	})

	t.Run("AdminWithCoAdmin_Success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, userID, bandID).Return("admin", nil)
		mockUserRepo.EXPECT().GetAdminCountInBand(ctx, bandID).Return(2, nil)
		mockUserRepo.EXPECT().FindBandsWithRoleByUserID(ctx, userID).Return([]model.BandWithRole{
			{ID: bandID, Name: "Band", IsDefault: false},
		}, nil)
		mockUserRepo.EXPECT().RemoveUserFromBand(ctx, bandID, userID).Return(nil)

		if err := svc.LeaveBand(ctx, userID, bandID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("NotMember_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, userID, bandID).Return("", pgx.ErrNoRows)

		err := svc.LeaveBand(ctx, userID, bandID)
		if !errors.Is(err, ErrNotBandMember) {
			t.Errorf("expected ErrNotBandMember, got %v", err)
		}
	})
}

func TestUserService_Signup_Validation(t *testing.T) {
	svc := UserService{}
	ctx := context.Background()

	_, err := svc.Signup(ctx, AuthPayload{Username: "ab", Password: "Password123!"})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %v", err)
	}
}

func TestUserService_Login_Errors(t *testing.T) {
	ctx := context.Background()
	payload := LoginPayload{Username: "testuser", Password: "Password123!"}

	t.Run("unknown user -> ErrInvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo, JWTSecret: "testsecret"}

		mockUserRepo.EXPECT().FindUserByUsername(ctx, payload.Username).Return(model.User{}, pgx.ErrNoRows)

		_, err := svc.Login(ctx, payload)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong password -> ErrInvalidCredentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo, JWTSecret: "testsecret"}

		hashedPw, _ := hashPasswordForTest("Password123!")
		mockUserRepo.EXPECT().FindUserByUsername(ctx, payload.Username).
			Return(model.User{ID: 1, Username: "testuser", PasswordHash: hashedPw}, nil)

		_, err := svc.Login(ctx, LoginPayload{Username: "testuser", Password: "wrongpassword"})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("propagates unexpected repository errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo, JWTSecret: "testsecret"}

		dbErr := errors.New("connection lost")
		mockUserRepo.EXPECT().FindUserByUsername(ctx, payload.Username).Return(model.User{}, dbErr)

		_, err := svc.Login(ctx, payload)
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected raw db error, got %v", err)
		}
	})
}

func TestUserService_UpdatePassword_Errors(t *testing.T) {
	ctx := context.Background()
	const userID = 1

	t.Run("user not found -> ErrUserNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		mockUserRepo.EXPECT().FindUserByID(ctx, userID).Return(model.User{}, pgx.ErrNoRows)

		err := svc.UpdatePassword(ctx, userID, UpdatePasswordPayload{CurrentPassword: "x", NewPassword: "Password123!"})
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("wrong current password -> ErrWrongCurrentPassword", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		hashedPw, _ := hashPasswordForTest("Password123!")
		mockUserRepo.EXPECT().FindUserByID(ctx, userID).
			Return(model.User{ID: userID, PasswordHash: hashedPw}, nil)

		err := svc.UpdatePassword(ctx, userID, UpdatePasswordPayload{CurrentPassword: "wrong", NewPassword: "NewPassword123!"})
		if !errors.Is(err, ErrWrongCurrentPassword) {
			t.Fatalf("expected ErrWrongCurrentPassword, got %v", err)
		}
	})
}

func TestUserService_InviteMember_Errors(t *testing.T) {
	ctx := context.Background()
	const bandID = 5

	t.Run("already member -> ErrAlreadyBandMember", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		existing := model.User{ID: 42, Username: "member"}
		mockUserRepo.EXPECT().FindUserByUsername(ctx, "member").Return(existing, nil)
		mockUserRepo.EXPECT().IsUserInBand(ctx, existing.ID, bandID).Return(true, nil)

		_, err := svc.InviteMember(ctx, bandID, InviteMemberPayload{Username: "member"})
		if !errors.Is(err, ErrAlreadyBandMember) {
			t.Fatalf("expected ErrAlreadyBandMember, got %v", err)
		}
	})

	t.Run("unknown user without password -> ErrUserPasswordRequired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		mockUserRepo.EXPECT().FindUserByUsername(ctx, "newuser").Return(model.User{}, pgx.ErrNoRows)

		_, err := svc.InviteMember(ctx, bandID, InviteMemberPayload{Username: "newuser"})
		if !errors.Is(err, ErrUserPasswordRequired) {
			t.Fatalf("expected ErrUserPasswordRequired, got %v", err)
		}
	})

	t.Run("empty username -> ValidationError", func(t *testing.T) {
		svc := UserService{}

		_, err := svc.InviteMember(ctx, bandID, InviteMemberPayload{Username: ""})

		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("expected *ValidationError, got %v", err)
		}
	})
}

func TestUserService_ChangeMemberRole(t *testing.T) {
	ctx := context.Background()
	const targetUserID, bandID = 10, 5

	t.Run("PromoteMemberToAdmin_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, targetUserID, bandID).Return("member", nil)
		mockUserRepo.EXPECT().UpdateUserRoleInBand(ctx, targetUserID, bandID, "admin").Return(nil)

		if err := svc.ChangeMemberRole(ctx, bandID, targetUserID, "admin"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("DemoteAdminWithCoAdmin_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, targetUserID, bandID).Return("admin", nil)
		mockUserRepo.EXPECT().GetAdminCountInBand(ctx, bandID).Return(2, nil)
		mockUserRepo.EXPECT().UpdateUserRoleInBand(ctx, targetUserID, bandID, "member").Return(nil)

		if err := svc.ChangeMemberRole(ctx, bandID, targetUserID, "member"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("DemoteLastAdmin_Blocked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, targetUserID, bandID).Return("admin", nil)
		mockUserRepo.EXPECT().GetAdminCountInBand(ctx, bandID).Return(1, nil)

		err := svc.ChangeMemberRole(ctx, bandID, targetUserID, "member")
		if !errors.Is(err, ErrCannotDemoteLastAdmin) {
			t.Errorf("expected ErrCannotDemoteLastAdmin, got %v", err)
		}
	})

	t.Run("InvalidRole_Rejected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		// aucun appel repo attendu : le rôle est rejeté avant.
		err := svc.ChangeMemberRole(ctx, bandID, targetUserID, "superadmin")
		if !errors.Is(err, ErrInvalidRole) {
			t.Errorf("expected ErrInvalidRole, got %v", err)
		}
	})

	t.Run("NotMember_Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, targetUserID, bandID).Return("", pgx.ErrNoRows)

		err := svc.ChangeMemberRole(ctx, bandID, targetUserID, "admin")
		if !errors.Is(err, ErrNotBandMember) {
			t.Errorf("expected ErrNotBandMember, got %v", err)
		}
	})

	t.Run("SameRole_NoOp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		svc := UserService{UserRepo: mockUserRepo}

		// déjà admin : pas d'UPDATE attendu.
		mockUserRepo.EXPECT().GetUserRoleInBand(ctx, targetUserID, bandID).Return("admin", nil)

		if err := svc.ChangeMemberRole(ctx, bandID, targetUserID, "admin"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUserService_SetDefaultBand_NotMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := UserService{UserRepo: mockUserRepo}
	ctx := context.Background()

	mockUserRepo.EXPECT().IsUserInBand(ctx, 1, 5).Return(false, nil)

	if err := svc.SetDefaultBand(ctx, 1, 5); !errors.Is(err, ErrBandNotFoundOrNotMember) {
		t.Fatalf("expected ErrBandNotFoundOrNotMember, got %v", err)
	}
}
