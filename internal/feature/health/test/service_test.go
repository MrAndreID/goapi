package health_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/MrAndreID/goapi/v2/internal/feature/health"
)

type stubRepository struct {
	checkFunc         func(context.Context) (bool, error)
	cacheFunc         func(context.Context) (bool, error)
	messageBrokerFunc func(context.Context) (bool, error)
	objectStorageFunc func(context.Context) (bool, error)
	checkCalls        int
}

func (s *stubRepository) CheckDatabase(ctx context.Context) (bool, error) {
	s.checkCalls++
	if s.checkFunc != nil {
		return s.checkFunc(ctx)
	}
	return true, nil
}

func (s *stubRepository) CheckCache(ctx context.Context) (bool, error) {
	if s.cacheFunc != nil {
		return s.cacheFunc(ctx)
	}
	return true, nil
}

func (s *stubRepository) CheckMessageBroker(ctx context.Context) (bool, error) {
	if s.messageBrokerFunc != nil {
		return s.messageBrokerFunc(ctx)
	}
	return true, nil
}

func (s *stubRepository) CheckObjectStorage(ctx context.Context) (bool, error) {
	if s.objectStorageFunc != nil {
		return s.objectStorageFunc(ctx)
	}
	return true, nil
}

func TestServiceCheckReturnsHealthyWhenRepositoryReportsUp(t *testing.T) {
	repo := &stubRepository{checkFunc: func(ctx context.Context) (bool, error) {
		return true, nil
	}}
	svc := NewService(repo)

	if err := svc.Check(context.Background()); err != nil {
		t.Fatalf("expected healthy, got %v", err)
	}

	if repo.checkCalls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.checkCalls)
	}
}

func TestServiceCheckReturnsSentinelWhenStatusFalse(t *testing.T) {
	repo := &stubRepository{checkFunc: func(ctx context.Context) (bool, error) {
		return false, nil
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if !errors.Is(err, ErrDatabaseUnavailable) {
		t.Fatalf("expected ErrDatabaseUnavailable, got %v", err)
	}
}

func TestServiceCheckPropagatesRepositoryError(t *testing.T) {
	repo := &stubRepository{checkFunc: func(ctx context.Context) (bool, error) {
		return false, errors.New("connection refused")
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if err == nil || err.Error() != "connection refused" {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceCheckReturnsCacheSentinelWhenCacheStatusFalse(t *testing.T) {
	repo := &stubRepository{cacheFunc: func(ctx context.Context) (bool, error) {
		return false, nil
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if !errors.Is(err, ErrCacheUnavailable) {
		t.Fatalf("expected ErrCacheUnavailable, got %v", err)
	}
}

func TestServiceCheckPropagatesCacheError(t *testing.T) {
	repo := &stubRepository{cacheFunc: func(ctx context.Context) (bool, error) {
		return false, errors.New("cache refused")
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if err == nil || err.Error() != "cache refused" {
		t.Fatalf("expected cache error, got %v", err)
	}
}

func TestServiceCheckReturnsMessageBrokerSentinelWhenStatusFalse(t *testing.T) {
	repo := &stubRepository{messageBrokerFunc: func(ctx context.Context) (bool, error) {
		return false, nil
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if !errors.Is(err, ErrMessageBrokerUnavailable) {
		t.Fatalf("expected ErrMessageBrokerUnavailable, got %v", err)
	}
}

func TestServiceCheckPropagatesMessageBrokerError(t *testing.T) {
	repo := &stubRepository{messageBrokerFunc: func(ctx context.Context) (bool, error) {
		return false, errors.New("broker refused")
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if err == nil || err.Error() != "broker refused" {
		t.Fatalf("expected message broker error, got %v", err)
	}
}

func TestServiceCheckReturnsObjectStorageSentinelWhenStatusFalse(t *testing.T) {
	repo := &stubRepository{objectStorageFunc: func(ctx context.Context) (bool, error) {
		return false, nil
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if !errors.Is(err, ErrObjectStorageUnavailable) {
		t.Fatalf("expected ErrObjectStorageUnavailable, got %v", err)
	}
}

func TestServiceCheckPropagatesObjectStorageError(t *testing.T) {
	repo := &stubRepository{objectStorageFunc: func(ctx context.Context) (bool, error) {
		return false, errors.New("object storage refused")
	}}
	svc := NewService(repo)

	err := svc.Check(context.Background())
	if err == nil || err.Error() != "object storage refused" {
		t.Fatalf("expected object storage error, got %v", err)
	}
}
