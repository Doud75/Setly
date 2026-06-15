package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"setlist/api/model"
	"setlist/api/repository/mocks"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func TestSetlistService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSetlistRepository(ctrl)
	// We don't need InterludeRepo for Create
	svc := SetlistService{SetlistRepo: mockRepo}
	ctx := context.Background()
	bandID := 1

	payload := CreateSetlistPayload{
		Name:  "My Setlist",
		Color: "#FF0000",
	}

	expectedSetlist := model.Setlist{
		BandID: bandID,
		Name:   payload.Name,
		Color:  payload.Color,
	}

	// Expect GetDB to be called. We can return nil because the mock implementation of CreateSetlist
	// doesn't actually use the DB connection, it just verifies arguments.
	mockRepo.EXPECT().GetDB().Return(nil)

	// Expect CreateSetlist with the nil DB
	mockRepo.EXPECT().
		CreateSetlist(ctx, nil, payload.Name, payload.Color, bandID).
		Return(expectedSetlist, nil)

	created, err := svc.Create(ctx, payload, bandID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Name != payload.Name {
		t.Errorf("expected name %s, got %s", payload.Name, created.Name)
	}
}

func TestSetlistService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSetlistRepository(ctrl)
	svc := SetlistService{SetlistRepo: mockRepo}
	ctx := context.Background()

	setlistID := 10
	bandID := 1
	newName := "Updated Name"
	payload := UpdateSetlistPayload{
		Name: &newName,
	}

	existingSetlist := model.Setlist{ID: setlistID, BandID: bandID, Name: "Old Name"}

	mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(existingSetlist, nil)

	updatedExpected := existingSetlist
	updatedExpected.Name = newName

	mockRepo.EXPECT().UpdateSetlist(ctx, updatedExpected).Return(updatedExpected, nil)

	updated, err := svc.Update(ctx, setlistID, bandID, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}
}

func TestSetlistService_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSetlistRepository(ctrl)
	// Mock Tx is generated in pgx_tx.go (package mocks)
	mockTx := mocks.NewMockTx(ctrl)

	svc := SetlistService{SetlistRepo: mockRepo}
	ctx := context.Background()

	originalID := 10
	bandID := 1
	newName := "Copy of Setlist"
	newColor := "#00FF00"

	// Expect transaction start
	mockRepo.EXPECT().BeginTx(ctx).Return(mockTx, nil)
	// Expect Rollback (deferred)
	mockTx.EXPECT().Rollback(ctx).Return(nil)

	// Expect Get original setlist
	mockRepo.EXPECT().GetSetlistByID(ctx, originalID, bandID).Return(model.Setlist{ID: originalID, BandID: bandID}, nil)

	// Expect Get items
	songID := int32(5)
	items := []model.SetlistItem{{ID: 1, SongID: &songID}}
	mockRepo.EXPECT().GetSetlistItemsBySetlistID(ctx, originalID).Return(items, nil)

	// Expect Create new setlist within TX
	newSetlist := model.Setlist{ID: 20, Name: newName, Color: newColor}
	mockRepo.EXPECT().CreateSetlist(ctx, mockTx, newName, newColor, bandID).Return(newSetlist, nil)

	// Expect Copy items within TX
	mockRepo.EXPECT().CopyItemsToNewSetlist(ctx, mockTx, newSetlist.ID, items).Return(nil)

	// Expect Commit
	mockTx.EXPECT().Commit(ctx).Return(nil)

	result, err := svc.Duplicate(ctx, originalID, bandID, newName, newColor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != newSetlist.ID {
		t.Errorf("expected new ID %d, got %d", newSetlist.ID, result.ID)
	}
}

func TestSetlistService_AddItem(t *testing.T) {
	ctx := context.Background()
	setlistID := 10
	bandID := 1
	ownedSetlist := model.Setlist{ID: setlistID, BandID: bandID}

	t.Run("rejects setlist from another band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		// Setlist does not belong to the band: repo returns no row.
		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(model.Setlist{}, pgx.ErrNoRows)
		// No AddItemToSetlist call expected.

		_, err := svc.AddItem(ctx, setlistID, bandID, AddItemPayload{ItemType: "song", ItemID: 5})
		if !errors.Is(err, ErrSetlistNotFound) {
			t.Fatalf("expected ErrSetlistNotFound, got %v", err)
		}
	})

	t.Run("rejects song from another band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		mockSongRepo := mocks.NewMockSongRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo, SongRepo: mockSongRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(ownedSetlist, nil)
		mockSongRepo.EXPECT().GetSongByID(ctx, 5, bandID).Return(model.Song{}, pgx.ErrNoRows)

		_, err := svc.AddItem(ctx, setlistID, bandID, AddItemPayload{ItemType: "song", ItemID: 5})
		if !errors.Is(err, ErrItemNotFound) {
			t.Fatalf("expected ErrItemNotFound, got %v", err)
		}
	})

	t.Run("rejects interlude from another band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		mockInterludeRepo := mocks.NewMockInterludeRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo, InterludeRepo: mockInterludeRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(ownedSetlist, nil)
		mockInterludeRepo.EXPECT().GetInterludeByID(ctx, 7, bandID).Return(model.Interlude{}, pgx.ErrNoRows)

		_, err := svc.AddItem(ctx, setlistID, bandID, AddItemPayload{ItemType: "interlude", ItemID: 7})
		if !errors.Is(err, ErrItemNotFound) {
			t.Fatalf("expected ErrItemNotFound, got %v", err)
		}
	})

	t.Run("rejects invalid item type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(ownedSetlist, nil)

		_, err := svc.AddItem(ctx, setlistID, bandID, AddItemPayload{ItemType: "video", ItemID: 5})
		if !errors.Is(err, ErrInvalidItemType) {
			t.Fatalf("expected ErrInvalidItemType, got %v", err)
		}
	})

	t.Run("adds a song owned by the band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		mockSongRepo := mocks.NewMockSongRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo, SongRepo: mockSongRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(ownedSetlist, nil)
		mockSongRepo.EXPECT().GetSongByID(ctx, 5, bandID).Return(model.Song{ID: 5, BandID: bandID}, nil)
		mockRepo.EXPECT().AddItemToSetlist(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, item model.SetlistItem) (model.SetlistItem, error) {
				if item.SetlistID != setlistID || item.ItemType != "song" || item.SongID == nil || *item.SongID != 5 {
					t.Errorf("unexpected item passed to repo: %+v", item)
				}
				item.ID = 1
				return item, nil
			})

		created, err := svc.AddItem(ctx, setlistID, bandID, AddItemPayload{ItemType: "song", ItemID: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ID != 1 {
			t.Errorf("expected created item ID 1, got %d", created.ID)
		}
	})

	t.Run("adds an interlude owned by the band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		mockInterludeRepo := mocks.NewMockInterludeRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo, InterludeRepo: mockInterludeRepo}

		script := "Talk to the crowd"
		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(ownedSetlist, nil)
		mockInterludeRepo.EXPECT().GetInterludeByID(ctx, 7, bandID).Return(model.Interlude{ID: 7, BandID: bandID, Script: &script}, nil)
		mockRepo.EXPECT().AddItemToSetlist(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, item model.SetlistItem) (model.SetlistItem, error) {
				if item.InterludeID == nil || *item.InterludeID != 7 || item.Notes == nil || *item.Notes != script {
					t.Errorf("unexpected item passed to repo: %+v", item)
				}
				return item, nil
			})

		_, err := svc.AddItem(ctx, setlistID, bandID, AddItemPayload{ItemType: "interlude", ItemID: 7})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSetlistService_UpdateOrder(t *testing.T) {
	ctx := context.Background()
	setlistID := 10
	bandID := 1

	t.Run("rejects setlist from another band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(model.Setlist{}, pgx.ErrNoRows)
		// No UpdateItemOrder call expected.

		err := svc.UpdateOrder(ctx, setlistID, bandID, UpdateOrderPayload{ItemIDs: []int{3, 1, 2}})
		if !errors.Is(err, ErrSetlistNotFound) {
			t.Fatalf("expected ErrSetlistNotFound, got %v", err)
		}
	})

	t.Run("reorders items of an owned setlist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		itemIDs := []int{3, 1, 2}
		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(model.Setlist{ID: setlistID, BandID: bandID}, nil)
		mockRepo.EXPECT().UpdateItemOrder(ctx, setlistID, itemIDs).Return(nil)

		if err := svc.UpdateOrder(ctx, setlistID, bandID, UpdateOrderPayload{ItemIDs: itemIDs}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSetlistService_Create_Validation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSetlistRepository(ctrl)
	svc := SetlistService{SetlistRepo: mockRepo}
	ctx := context.Background()

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateSetlistPayload{Name: "", Color: "#FF0000"}, 1)
		if !errors.Is(err, ErrSetlistNameRequired) {
			t.Fatalf("expected ErrSetlistNameRequired, got %v", err)
		}
	})

	t.Run("rejects invalid color", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateSetlistPayload{Name: "My Setlist", Color: "red"}, 1)
		if !errors.Is(err, ErrInvalidColor) {
			t.Fatalf("expected ErrInvalidColor, got %v", err)
		}
	})
}

func TestSetlistService_Update_Errors(t *testing.T) {
	ctx := context.Background()
	setlistID := 10
	bandID := 1

	t.Run("returns ErrSetlistNotFound when setlist is not in band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(model.Setlist{}, pgx.ErrNoRows)

		newName := "New Name"
		_, err := svc.Update(ctx, setlistID, bandID, UpdateSetlistPayload{Name: &newName})
		if !errors.Is(err, ErrSetlistNotFound) {
			t.Fatalf("expected ErrSetlistNotFound, got %v", err)
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(model.Setlist{ID: setlistID, BandID: bandID}, nil)

		emptyName := ""
		_, err := svc.Update(ctx, setlistID, bandID, UpdateSetlistPayload{Name: &emptyName})
		if !errors.Is(err, ErrSetlistNameRequired) {
			t.Fatalf("expected ErrSetlistNameRequired, got %v", err)
		}
	})

	t.Run("propagates unexpected repository errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		dbErr := errors.New("connection lost")
		mockRepo.EXPECT().GetSetlistByID(ctx, setlistID, bandID).Return(model.Setlist{}, dbErr)

		newName := "New Name"
		_, err := svc.Update(ctx, setlistID, bandID, UpdateSetlistPayload{Name: &newName})
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected raw db error, got %v", err)
		}
	})
}

func TestSetlistService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSetlistRepository(ctrl)
	svc := SetlistService{SetlistRepo: mockRepo}
	ctx := context.Background()

	mockRepo.EXPECT().GetSetlistByID(ctx, 10, 1).Return(model.Setlist{}, pgx.ErrNoRows)

	if err := svc.Delete(ctx, 10, 1); !errors.Is(err, ErrSetlistNotFound) {
		t.Fatalf("expected ErrSetlistNotFound, got %v", err)
	}
}

func TestSetlistService_GetDetails_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSetlistRepository(ctrl)
	svc := SetlistService{SetlistRepo: mockRepo}
	ctx := context.Background()

	mockRepo.EXPECT().GetSetlistByID(ctx, 10, 1).Return(model.Setlist{}, pgx.ErrNoRows)

	if _, err := svc.GetDetails(ctx, 10, 1); !errors.Is(err, ErrSetlistNotFound) {
		t.Fatalf("expected ErrSetlistNotFound, got %v", err)
	}
}

func TestSetlistService_Duplicate_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects empty name", func(t *testing.T) {
		svc := SetlistService{}
		_, err := svc.Duplicate(ctx, 10, 1, "", "#00FF00")
		if !errors.Is(err, ErrSetlistNameRequired) {
			t.Fatalf("expected ErrSetlistNameRequired, got %v", err)
		}
	})

	t.Run("returns ErrSetlistNotFound when original is not in band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		mockTx := mocks.NewMockTx(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().BeginTx(ctx).Return(mockTx, nil)
		mockTx.EXPECT().Rollback(ctx).Return(nil)
		mockRepo.EXPECT().GetSetlistByID(ctx, 10, 1).Return(model.Setlist{}, pgx.ErrNoRows)

		_, err := svc.Duplicate(ctx, 10, 1, "Copy", "#00FF00")
		if !errors.Is(err, ErrSetlistNotFound) {
			t.Fatalf("expected ErrSetlistNotFound, got %v", err)
		}
	})
}

func TestSetlistService_ItemErrors(t *testing.T) {
	ctx := context.Background()
	itemID := 3
	bandID := 1

	t.Run("UpdateItem returns ErrItemNotFound when item is not in band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().UpdateSetlistItem(ctx, itemID, bandID, gomock.Any()).Return(model.SetlistItem{}, pgx.ErrNoRows)

		_, err := svc.UpdateItem(ctx, itemID, bandID, UpdateItemPayload{Notes: "notes"})
		if !errors.Is(err, ErrItemNotFound) {
			t.Fatalf("expected ErrItemNotFound, got %v", err)
		}
	})

	t.Run("DeleteItem returns ErrItemNotFound when item is not in band", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSetlistRepository(ctrl)
		svc := SetlistService{SetlistRepo: mockRepo}

		mockRepo.EXPECT().DeleteSetlistItem(ctx, itemID, bandID).Return(0, sql.ErrNoRows)

		if err := svc.DeleteItem(ctx, itemID, bandID); !errors.Is(err, ErrItemNotFound) {
			t.Fatalf("expected ErrItemNotFound, got %v", err)
		}
	})
}
