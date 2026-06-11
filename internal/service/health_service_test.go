package service

import (
	"context"
	"errors"
	"testing"
)

type MockDBPinger struct{ err error }

func (m *MockDBPinger) Ping(ctx context.Context) error { return m.err }

type MockRedisPinger struct{ err error }

func (m *MockRedisPinger) Ping(ctx context.Context) error { return m.err }

func TestHealthService_Check_Healthy(t *testing.T) {
	dbPinger := &MockDBPinger{err: nil}
	redisPinger := &MockRedisPinger{err: nil}

	svc := NewHealthService(dbPinger, redisPinger)

	healthStats := svc.Check(context.Background())
	if healthStats.Status != "ok" {
		t.Errorf("expected status ok, got %v", healthStats.Status)
	}
	if healthStats.DB != "ok" {
		t.Errorf("expected db status ok, got %v", healthStats.DB)
	}
	if healthStats.Redis != "ok" {
		t.Errorf("expected redis status ok, got %v", healthStats.DB)
	}
}

func TestHealthService_Check_DBDegraded(t *testing.T) {
	dbPinger := &MockDBPinger{err: errors.New("unexpected error")}
	redisPinger := &MockRedisPinger{err: nil}

	svc := NewHealthService(dbPinger, redisPinger)

	healthStats := svc.Check(context.Background())
	if healthStats.Status != "degraded" {
		t.Errorf("expected status degraded, got %v", healthStats.Status)
	}
	if healthStats.DB != "unavailable" {
		t.Errorf("expected db status unavailable, got %v", healthStats.DB)
	}
	if healthStats.Redis != "ok" {
		t.Errorf("expected redis status ok, got %v", healthStats.DB)
	}
}

func TestHealthService_Check_RedisDegraded(t *testing.T) {
	dbPinger := &MockDBPinger{err: nil}
	redisPinger := &MockRedisPinger{err: errors.New("unexpected error")}

	svc := NewHealthService(dbPinger, redisPinger)

	healthStats := svc.Check(context.Background())
	if healthStats.Status != "degraded" {
		t.Errorf("expected status degraded, got %v", healthStats.Status)
	}
	if healthStats.DB != "ok" {
		t.Errorf("expected db status ok, got %v", healthStats.DB)
	}
	if healthStats.Redis != "unavailable" {
		t.Errorf("expected redis status unavailable, got %v", healthStats.DB)
	}
}

func TestHealthService_Check_DBAndRedisDegraded(t *testing.T) {
	dbPinger := &MockDBPinger{err: errors.New("unexpected error")}
	redisPinger := &MockRedisPinger{err: errors.New("unexpected error")}

	svc := NewHealthService(dbPinger, redisPinger)

	healthStats := svc.Check(context.Background())
	if healthStats.Status != "degraded" {
		t.Errorf("expected status degraded, got %v", healthStats.Status)
	}
	if healthStats.DB != "unavailable" {
		t.Errorf("expected db status unavailable, got %v", healthStats.DB)
	}
	if healthStats.Redis != "unavailable" {
		t.Errorf("expected redis status unavailable, got %v", healthStats.DB)
	}
}
