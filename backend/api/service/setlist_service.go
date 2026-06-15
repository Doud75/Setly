package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"setlist/api/model"
	"setlist/api/repository"
	"setlist/cache"
	"time"

	"github.com/redis/go-redis/v9"
)

const setlistCacheTTL = 30 * time.Minute

type SetlistService struct {
	SetlistRepo   repository.SetlistRepository
	InterludeRepo repository.InterludeRepository
	SongRepo      repository.SongRepository
	Cache         *redis.Client
}

var (
	ErrSetlistNotFound     = errors.New("setlist not found or does not belong to the user's band")
	ErrItemNotFound        = errors.New("song or interlude not found or does not belong to the user's band")
	ErrInvalidItemType     = errors.New("invalid item type")
	ErrSetlistNameRequired = errors.New("setlist name cannot be empty")
	ErrInvalidColor        = errors.New("invalid color format")
)

type CreateSetlistPayload struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateSetlistPayload struct {
	Name       *string `json:"name"`
	Color      *string `json:"color"`
	IsArchived *bool   `json:"is_archived"`
}

type SetlistDetails struct {
	model.Setlist
	Items []model.SetlistItem `json:"items"`
}

type AddItemPayload struct {
	ItemType string `json:"item_type"`
	ItemID   int    `json:"item_id"`
	Notes    string `json:"notes"`
}

type UpdateOrderPayload struct {
	ItemIDs []int `json:"item_ids"`
}

type UpdateItemPayload struct {
	Notes string `json:"notes"`
}

type DuplicateSetlistPayload struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

var hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)

func (s SetlistService) Create(ctx context.Context, payload CreateSetlistPayload, bandID int) (model.Setlist, error) {
	if payload.Name == "" {
		return model.Setlist{}, ErrSetlistNameRequired
	}
	if payload.Color == "" || !hexColorRegex.MatchString(payload.Color) {
		return model.Setlist{}, ErrInvalidColor
	}

	created, err := s.SetlistRepo.CreateSetlist(ctx, s.SetlistRepo.GetDB(), payload.Name, payload.Color, bandID)
	if err != nil {
		return model.Setlist{}, err
	}

	cache.Delete(ctx, s.Cache, cache.SetlistKey(bandID))
	return created, nil
}

func (s SetlistService) Update(ctx context.Context, id int, bandID int, payload UpdateSetlistPayload) (model.Setlist, error) {
	setlist, err := s.SetlistRepo.GetSetlistByID(ctx, id, bandID)
	if err != nil {
		return model.Setlist{}, mapNotFound(err, ErrSetlistNotFound)
	}

	if payload.Name != nil {
		if *payload.Name == "" {
			return model.Setlist{}, ErrSetlistNameRequired
		}
		setlist.Name = *payload.Name
	}
	if payload.Color != nil {
		if *payload.Color == "" || !hexColorRegex.MatchString(*payload.Color) {
			return model.Setlist{}, ErrInvalidColor
		}
		setlist.Color = *payload.Color
	}
	if payload.IsArchived != nil {
		setlist.IsArchived = *payload.IsArchived
	}

	updated, err := s.SetlistRepo.UpdateSetlist(ctx, setlist)
	if err != nil {
		return model.Setlist{}, err
	}

	cache.Delete(ctx, s.Cache, cache.SetlistKey(bandID))
	return updated, nil
}

func (s SetlistService) Delete(ctx context.Context, setlistID int, bandID int) error {
	_, err := s.SetlistRepo.GetSetlistByID(ctx, setlistID, bandID)
	if err != nil {
		return mapNotFound(err, ErrSetlistNotFound)
	}
	if err := s.SetlistRepo.DeleteSetlist(ctx, setlistID, bandID); err != nil {
		return err
	}

	cache.Delete(ctx, s.Cache, cache.SetlistKey(bandID))
	return nil
}

func (s SetlistService) GetAllForBand(ctx context.Context, bandID int) ([]model.Setlist, error) {
	key := cache.SetlistKey(bandID)

	if data, ok := cache.Get(ctx, s.Cache, key); ok {
		var setlists []model.Setlist
		if err := json.Unmarshal([]byte(data), &setlists); err == nil {
			return setlists, nil
		}
	}

	setlists, err := s.SetlistRepo.GetSetlistsByBandID(ctx, bandID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(setlists); err == nil {
		cache.Set(ctx, s.Cache, key, string(data), setlistCacheTTL)
	}

	return setlists, nil
}

// GetDetails returns a setlist with its items. It is NOT cached yet. When
// caching is added, store the result under cache.SetlistDetailKey(id); the
// item mutations (AddItem, UpdateItem, DeleteItem, UpdateOrder) already
// invalidate that key, so no extra wiring is required.
func (s SetlistService) GetDetails(ctx context.Context, id int, bandID int) (SetlistDetails, error) {
	setlist, err := s.SetlistRepo.GetSetlistByID(ctx, id, bandID)
	if err != nil {
		return SetlistDetails{}, mapNotFound(err, ErrSetlistNotFound)
	}
	items, err := s.SetlistRepo.GetSetlistItemsBySetlistID(ctx, id)
	if err != nil {
		return SetlistDetails{}, err
	}
	return SetlistDetails{Setlist: setlist, Items: items}, nil
}

func (s SetlistService) AddItem(ctx context.Context, setlistID int, bandID int, payload AddItemPayload) (model.SetlistItem, error) {
	if _, err := s.SetlistRepo.GetSetlistByID(ctx, setlistID, bandID); err != nil {
		return model.SetlistItem{}, mapNotFound(err, ErrSetlistNotFound)
	}

	var notes *string
	if payload.Notes != "" {
		notes = &payload.Notes
	}
	item := model.SetlistItem{
		SetlistID: setlistID,
		ItemType:  payload.ItemType,
		Notes:     notes,
	}

	if payload.ItemType == "song" {
		itemID := int32(payload.ItemID)
		item.SongID = &itemID
		if _, err := s.SongRepo.GetSongByID(ctx, payload.ItemID, bandID); err != nil {
			return model.SetlistItem{}, mapNotFound(err, ErrItemNotFound)
		}
	} else if payload.ItemType == "interlude" {
		itemID := int32(payload.ItemID)
		item.InterludeID = &itemID
		interlude, err := s.InterludeRepo.GetInterludeByID(ctx, payload.ItemID, bandID)
		if err != nil {
			return model.SetlistItem{}, mapNotFound(err, ErrItemNotFound)
		}
		item.Notes = interlude.Script
	} else {
		return model.SetlistItem{}, ErrInvalidItemType
	}
	created, err := s.SetlistRepo.AddItemToSetlist(ctx, item)
	if err != nil {
		return model.SetlistItem{}, err
	}
	cache.Delete(ctx, s.Cache, cache.SetlistDetailKey(setlistID))
	return created, nil
}

func (s SetlistService) UpdateOrder(ctx context.Context, setlistID int, bandID int, payload UpdateOrderPayload) error {
	if _, err := s.SetlistRepo.GetSetlistByID(ctx, setlistID, bandID); err != nil {
		return mapNotFound(err, ErrSetlistNotFound)
	}
	if len(payload.ItemIDs) == 0 {
		return nil
	}
	if err := s.SetlistRepo.UpdateItemOrder(ctx, setlistID, payload.ItemIDs); err != nil {
		return err
	}
	cache.Delete(ctx, s.Cache, cache.SetlistDetailKey(setlistID))
	return nil
}

func (s SetlistService) UpdateItem(ctx context.Context, itemID int, bandID int, payload UpdateItemPayload) (model.SetlistItem, error) {
	var notes *string
	if payload.Notes != "" {
		notes = &payload.Notes
	}
	item, err := s.SetlistRepo.UpdateSetlistItem(ctx, itemID, bandID, notes)
	if err != nil {
		return model.SetlistItem{}, mapNotFound(err, ErrItemNotFound)
	}
	cache.Delete(ctx, s.Cache, cache.SetlistDetailKey(item.SetlistID))
	return item, nil
}

func (s SetlistService) DeleteItem(ctx context.Context, itemID int, bandID int) error {
	setlistID, err := s.SetlistRepo.DeleteSetlistItem(ctx, itemID, bandID)
	if err != nil {
		return mapNotFound(err, ErrItemNotFound)
	}
	cache.Delete(ctx, s.Cache, cache.SetlistDetailKey(setlistID))
	return nil
}

func (s SetlistService) Duplicate(ctx context.Context, originalSetlistID int, bandID int, newName, newColor string) (model.Setlist, error) {
	if newName == "" {
		return model.Setlist{}, ErrSetlistNameRequired
	}
	if newColor == "" || !hexColorRegex.MatchString(newColor) {
		return model.Setlist{}, ErrInvalidColor
	}

	tx, err := s.SetlistRepo.BeginTx(ctx)
	if err != nil {
		return model.Setlist{}, err
	}
	defer tx.Rollback(ctx)

	_, err = s.SetlistRepo.GetSetlistByID(ctx, originalSetlistID, bandID)
	if err != nil {
		return model.Setlist{}, mapNotFound(err, ErrSetlistNotFound)
	}

	originalItems, err := s.SetlistRepo.GetSetlistItemsBySetlistID(ctx, originalSetlistID)
	if err != nil {
		return model.Setlist{}, err
	}

	newSetlist, err := s.SetlistRepo.CreateSetlist(ctx, tx, newName, newColor, bandID)
	if err != nil {
		return model.Setlist{}, err
	}

	if err := s.SetlistRepo.CopyItemsToNewSetlist(ctx, tx, newSetlist.ID, originalItems); err != nil {
		return model.Setlist{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Setlist{}, err
	}

	cache.Delete(ctx, s.Cache, cache.SetlistKey(bandID))
	return newSetlist, nil
}
