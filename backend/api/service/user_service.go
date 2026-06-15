package service

import (
	"context"
	"errors"
	"setlist/api/model"
	"setlist/api/repository"
	"setlist/auth"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserNotFound            = errors.New("user not found")
	ErrWrongCurrentPassword    = errors.New("invalid current password")
	ErrAlreadyBandMember       = errors.New("user is already a member of this band")
	ErrUserPasswordRequired    = errors.New("user not found, and password is required to create a new one")
	ErrNotBandMember           = errors.New("you are not a member of this band")
	ErrLastAdmin               = errors.New("cannot leave: user is the last admin of the band")
	ErrBandNameRequired        = errors.New("band name cannot be empty")
	ErrBandNotFoundOrNotMember = errors.New("band not found or user is not a member")
)

type UserService struct {
	UserRepo         repository.UserRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	JWTSecret        string
}

type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdatePasswordPayload struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type InviteMemberPayload struct {
	Username string  `json:"username"`
	Password *string `json:"password"`
}

type AuthResponse struct {
	Token         string              `json:"token"`
	RefreshToken  string              `json:"refresh_token"`
	Bands         []model.BandWithRole `json:"bands"`
	DefaultBandID *int                `json:"default_band_id"`
}

func (s UserService) Signup(ctx context.Context, payload AuthPayload) (*AuthResponse, error) {
	if err := ValidateUsername(payload.Username); err != nil {
		return nil, err
	}
	if err := ValidatePassword(payload.Password); err != nil {
		return nil, err
	}

	payload.Username = SanitizeString(payload.Username)

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.UserRepo.CreateUser(ctx, payload.Username, hashedPassword)
	if err != nil {
		return nil, err
	}

	token, err := auth.GenerateJWT(s.JWTSecret, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash, err := auth.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(auth.RefreshTokenDuration)
	err = s.RefreshTokenRepo.ReplaceUserRefreshToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Bands:        []model.BandWithRole{},
	}, nil
}

func (s UserService) Login(ctx context.Context, payload LoginPayload) (*AuthResponse, error) {
	user, err := s.UserRepo.FindUserByUsername(ctx, payload.Username)
	if err != nil {
		return nil, mapNotFound(err, ErrInvalidCredentials)
	}

	if !auth.CheckPasswordHash(payload.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	bands, err := s.UserRepo.FindBandsWithRoleByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var defaultBandID *int
	for _, b := range bands {
		if b.IsDefault {
			id := b.ID
			defaultBandID = &id
			break
		}
	}

	token, err := auth.GenerateJWT(s.JWTSecret, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash, err := auth.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(auth.RefreshTokenDuration)
	err = s.RefreshTokenRepo.ReplaceUserRefreshToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:         token,
		RefreshToken:  refreshToken,
		Bands:         bands,
		DefaultBandID: defaultBandID,
	}, nil
}

func (s UserService) UpdatePassword(ctx context.Context, userID int, payload UpdatePasswordPayload) error {
	user, err := s.UserRepo.FindUserByID(ctx, userID)
	if err != nil {
		return mapNotFound(err, ErrUserNotFound)
	}

	if !auth.CheckPasswordHash(payload.CurrentPassword, user.PasswordHash) {
		return ErrWrongCurrentPassword
	}

	if err := ValidatePassword(payload.NewPassword); err != nil {
		return err
	}

	newHashedPassword, err := auth.HashPassword(payload.NewPassword)
	if err != nil {
		return err
	}

	return s.UserRepo.UpdatePassword(ctx, userID, newHashedPassword)
}

func (s UserService) GetBandMembers(ctx context.Context, bandID int) ([]model.BandMember, error) {
	return s.UserRepo.GetMembersByBandID(ctx, bandID)
}

func (s UserService) InviteMember(ctx context.Context, bandID int, payload InviteMemberPayload) (model.User, error) {
	if payload.Username == "" {
		return model.User{}, &ValidationError{Msg: "le nom d'utilisateur est requis"}
	}
	if err := ValidateUsername(payload.Username); err != nil {
		return model.User{}, err
	}
	payload.Username = SanitizeString(payload.Username)

	existingUser, err := s.UserRepo.FindUserByUsername(ctx, payload.Username)
	if err == nil {
		isMember, err := s.UserRepo.IsUserInBand(ctx, existingUser.ID, bandID)
		if err != nil {
			return model.User{}, err
		}
		if isMember {
			return model.User{}, ErrAlreadyBandMember
		}

		err = s.UserRepo.AddUserToBand(ctx, existingUser.ID, bandID, "member")
		if err != nil {
			return model.User{}, err
		}
		return existingUser, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, err
	}

	if payload.Password == nil || *payload.Password == "" {
		return model.User{}, ErrUserPasswordRequired
	}
	if err := ValidatePassword(*payload.Password); err != nil {
		return model.User{}, err
	}

	hashedPassword, err := auth.HashPassword(*payload.Password)
	if err != nil {
		return model.User{}, err
	}

	newUser, err := s.UserRepo.CreateUserAndAddToBand(ctx, bandID, payload.Username, hashedPassword, "member")
	if err != nil {
		return model.User{}, err
	}

	return newUser, nil
}

func (s UserService) RemoveMember(ctx context.Context, bandID int, userID int) error {
	return s.UserRepo.RemoveUserFromBand(ctx, bandID, userID)
}

func (s UserService) LeaveBand(ctx context.Context, userID int, bandID int) error {
	role, err := s.UserRepo.GetUserRoleInBand(ctx, userID, bandID)
	if err != nil {
		return mapNotFound(err, ErrNotBandMember)
	}

	if role == "admin" {
		count, err := s.UserRepo.GetAdminCountInBand(ctx, bandID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastAdmin
		}
	}

	allBands, err := s.UserRepo.FindBandsWithRoleByUserID(ctx, userID)
	if err != nil {
		return err
	}
	var wasDefault bool
	for _, b := range allBands {
		if b.ID == bandID && b.IsDefault {
			wasDefault = true
			break
		}
	}

	if err := s.UserRepo.RemoveUserFromBand(ctx, bandID, userID); err != nil {
		return err
	}

	if wasDefault {
		for _, b := range allBands {
			if b.ID != bandID {
				_ = s.UserRepo.SetDefaultBand(ctx, userID, b.ID)
				break
			}
		}
	}

	return nil
}

func (s UserService) CreateBand(ctx context.Context, name string, ownerUserID int) (model.Band, error) {
	if name == "" {
		return model.Band{}, ErrBandNameRequired
	}
	band, err := s.UserRepo.CreateBand(ctx, name, ownerUserID)
	if err != nil {
		return model.Band{}, err
	}

	bands, err := s.UserRepo.FindBandsWithRoleByUserID(ctx, ownerUserID)
	if err == nil && len(bands) == 1 {
		_ = s.UserRepo.SetDefaultBand(ctx, ownerUserID, band.ID)
	}

	return band, nil
}

func (s UserService) GetUserBands(ctx context.Context, userID int) ([]model.BandWithRole, error) {
	return s.UserRepo.FindBandsWithRoleByUserID(ctx, userID)
}

func (s UserService) SetDefaultBand(ctx context.Context, userID int, bandID int) error {
	isMember, err := s.UserRepo.IsUserInBand(ctx, userID, bandID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrBandNotFoundOrNotMember
	}
	return s.UserRepo.SetDefaultBand(ctx, userID, bandID)
}
