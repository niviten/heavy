package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/niviten/heavy/internal/model"
)

type stubMuscleGroupRepository struct {
	mu           sync.Mutex
	muscleGroups []model.MuscleGroup
	err          error
	calls        int
}

func (r *stubMuscleGroupRepository) GetAll(context.Context) ([]model.MuscleGroup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	return r.muscleGroups, r.err
}

func (r *stubMuscleGroupRepository) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func TestMuscleGroupServiceLoadsOnce(t *testing.T) {
	repository := &stubMuscleGroupRepository{muscleGroups: []model.MuscleGroup{
		{ID: 1, Name: "chest", DisplayName: "Chest"},
		{ID: 2, Name: "back", DisplayName: "Back"},
	}}
	muscleGroups := NewMuscleGroupService(repository)

	all, err := muscleGroups.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	byID, err := muscleGroups.GetByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	byName, err := muscleGroups.GetByName(context.Background(), " chest ")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}

	if len(all) != 2 || byID.DisplayName != "Back" || byName.ID != 1 {
		t.Fatalf("unexpected results: all=%+v byID=%+v byName=%+v", all, byID, byName)
	}
	if repository.callCount() != 1 {
		t.Errorf("repository calls = %d, want 1", repository.callCount())
	}

	all[0].Name = "changed by caller"
	cached, err := muscleGroups.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID() after result mutation error = %v", err)
	}
	if cached.Name != "chest" || cached.DisplayName != "Chest" {
		t.Errorf("cached muscle group = %+v, want original values", cached)
	}
}

func TestMuscleGroupServiceNotFound(t *testing.T) {
	muscleGroups := NewMuscleGroupService(&stubMuscleGroupRepository{})

	if _, err := muscleGroups.GetByID(context.Background(), 99); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("GetByID() error = %v, want ErrMuscleGroupNotFound", err)
	}
	if _, err := muscleGroups.GetByName(context.Background(), "unknown"); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("GetByName() error = %v, want ErrMuscleGroupNotFound", err)
	}
}

func TestMuscleGroupServiceRetriesFailedLoad(t *testing.T) {
	repository := &stubMuscleGroupRepository{err: errors.New("temporary database error")}
	muscleGroups := NewMuscleGroupService(repository)

	if _, err := muscleGroups.GetAll(context.Background()); err == nil {
		t.Fatal("GetAll() error = nil, want database error")
	}

	repository.mu.Lock()
	repository.err = nil
	repository.muscleGroups = []model.MuscleGroup{{ID: 1, Name: "Chest"}}
	repository.mu.Unlock()

	if _, err := muscleGroups.GetAll(context.Background()); err != nil {
		t.Fatalf("GetAll() retry error = %v", err)
	}
	if repository.callCount() != 2 {
		t.Errorf("repository calls = %d, want 2", repository.callCount())
	}
}

func TestMuscleGroupServiceConcurrentLoad(t *testing.T) {
	repository := &stubMuscleGroupRepository{muscleGroups: []model.MuscleGroup{{ID: 1, Name: "Chest"}}}
	muscleGroups := NewMuscleGroupService(repository)

	var waitGroup sync.WaitGroup
	for range 20 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if _, err := muscleGroups.GetAll(context.Background()); err != nil {
				t.Errorf("GetAll() error = %v", err)
			}
		}()
	}
	waitGroup.Wait()

	if repository.callCount() != 1 {
		t.Errorf("repository calls = %d, want 1", repository.callCount())
	}
}
