package user

import (
	"context"
	"errors"
	"testing"

	"github.com/MrAndreID/goapi/v2/internal/entity"
)

type stubRepository struct {
	createFunc  func(CreateData) (User, error)
	readFunc    func(context.Context, ReadData) (entity.PaginatorResponse, error)
	updateFunc  func(UpdateData) error
	deleteFunc  func(DeleteData) error
	createCalls int
	updateCalls int
	deleteCalls int
}

func (s *stubRepository) Create(req CreateData) (User, error) {
	s.createCalls++
	if s.createFunc != nil {
		return s.createFunc(req)
	}
	return User{}, nil
}

func (s *stubRepository) Read(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
	if s.readFunc != nil {
		return s.readFunc(ctx, req)
	}
	return entity.PaginatorResponse{}, nil
}

func (s *stubRepository) Update(req UpdateData) error {
	s.updateCalls++
	if s.updateFunc != nil {
		return s.updateFunc(req)
	}
	return nil
}

func (s *stubRepository) Delete(req DeleteData) error {
	s.deleteCalls++
	if s.deleteFunc != nil {
		return s.deleteFunc(req)
	}
	return nil
}

func TestServiceCreateRejectsDuplicateEmails(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo)

	_, err := svc.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com", "andre@gmail.com"}})
	if err == nil {
		t.Fatal("expected duplicate email error")
	}

	if err.Error() != "DUPLICATE_EMAIL" {
		t.Fatalf("expected DUPLICATE_EMAIL, got %v", err)
	}

	if repo.createCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.createCalls)
	}
}

func TestServiceCreateDelegatesToRepository(t *testing.T) {
	repo := &stubRepository{createFunc: func(req CreateData) (User, error) {
		return User{ID: "user-1", Name: req.Name, Emails: []Email{{Email: req.Emails[0]}}}, nil
	}}
	svc := NewService(repo)

	user, err := svc.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-1" || user.Name != "Andre" {
		t.Fatalf("unexpected user returned: %+v", user)
	}

	if repo.createCalls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.createCalls)
	}
}

func TestServiceUpdateRejectsDuplicateEmails(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo)

	err := svc.Update(UpdateData{ID: "user-1", Emails: []string{"andre@gmail.com", "andre@gmail.com"}})
	if err == nil {
		t.Fatal("expected duplicate email error")
	}

	if err.Error() != "DUPLICATE_EMAIL" {
		t.Fatalf("expected DUPLICATE_EMAIL, got %v", err)
	}

	if repo.updateCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.updateCalls)
	}
}

func TestServiceDeleteDelegatesToRepository(t *testing.T) {
	repo := &stubRepository{deleteFunc: func(req DeleteData) error {
		if req.ID != "user-1" {
			t.Fatalf("unexpected delete id: %s", req.ID)
		}
		return nil
	}}
	svc := NewService(repo)

	if err := svc.Delete(DeleteData{ID: "user-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.deleteCalls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.deleteCalls)
	}
}

func TestServiceReadDelegatesToRepository(t *testing.T) {
	repo := &stubRepository{readFunc: func(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
		return entity.PaginatorResponse{Total: 1}, nil
	}}
	svc := NewService(repo)

	data, err := svc.Read(context.Background(), ReadData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Total != 1 {
		t.Fatalf("expected total 1, got %d", data.Total)
	}
}

func TestServiceReadReturnsRepositoryError(t *testing.T) {
	repo := &stubRepository{readFunc: func(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
		return entity.PaginatorResponse{}, errors.New("read failed")
	}}
	svc := NewService(repo)

	_, err := svc.Read(context.Background(), ReadData{})
	if err == nil || err.Error() != "read failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceUpdateReturnsRepositoryError(t *testing.T) {
	repo := &stubRepository{updateFunc: func(req UpdateData) error {
		return errors.New("update failed")
	}}
	svc := NewService(repo)

	err := svc.Update(UpdateData{ID: "user-1", Emails: []string{"andre@gmail.com"}})
	if err == nil || err.Error() != "update failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceDeleteReturnsRepositoryError(t *testing.T) {
	repo := &stubRepository{deleteFunc: func(req DeleteData) error {
		return errors.New("delete failed")
	}}
	svc := NewService(repo)

	err := svc.Delete(DeleteData{ID: "user-1"})
	if err == nil || err.Error() != "delete failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceCreateReturnsRepositoryError(t *testing.T) {
	repo := &stubRepository{createFunc: func(req CreateData) (User, error) {
		return User{}, errors.New("create failed")
	}}
	svc := NewService(repo)

	if _, err := svc.Create(CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}}); err == nil || err.Error() != "create failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceUpdateDelegatesToRepository(t *testing.T) {
	repo := &stubRepository{updateFunc: func(req UpdateData) error {
		if req.ID != "user-1" {
			t.Fatalf("unexpected update id: %s", req.ID)
		}
		return nil
	}}
	svc := NewService(repo)

	if err := svc.Update(UpdateData{ID: "user-1", Name: "Andre", Emails: []string{"andre@gmail.com", "bndre@gmail.com"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.updateCalls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.updateCalls)
	}
}
