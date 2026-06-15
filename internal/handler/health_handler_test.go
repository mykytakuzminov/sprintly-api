package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

type MockHealthService struct {
	checkFn func(ctx context.Context) *domain.HealthStats
}

func (m *MockHealthService) Check(ctx context.Context) *domain.HealthStats {
	if m.checkFn != nil {
		return m.checkFn(ctx)
	}
	return &domain.HealthStats{}
}

func newTestHealthHandler(svc domain.HealthService) *HealthHandler {
	logger, _ := zap.NewDevelopment()
	return NewHealthHandler(svc, logger.Sugar())
}

func TestHealthHandler_Health_Healthy(t *testing.T) {
	svc := &MockHealthService{
		checkFn: func(_ context.Context) *domain.HealthStats {
			return &domain.HealthStats{
				Status: "ok",
				DB:     "ok",
				Redis:  "ok",
				Uptime: 3600,
			}
		},
	}

	h := newTestHealthHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp domain.HealthStats
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %v", resp.Status)
	}
	if resp.DB != "ok" {
		t.Errorf("expected db 'ok', got %v", resp.DB)
	}
	if resp.Redis != "ok" {
		t.Errorf("expected redis 'ok', got %v", resp.Redis)
	}
	if resp.Uptime != 3600 {
		t.Errorf("expected uptime 3600, got %v", resp.Uptime)
	}
}

func TestHealthHandler_Health_DBDegraded(t *testing.T) {
	svc := &MockHealthService{
		checkFn: func(_ context.Context) *domain.HealthStats {
			return &domain.HealthStats{
				Status: "degraded",
				DB:     "unavailable",
				Redis:  "ok",
			}
		},
	}

	h := newTestHealthHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}

	var resp domain.HealthStats
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got %v", resp.Status)
	}
	if resp.DB != "unavailable" {
		t.Errorf("expected db 'unavailable', got %v", resp.DB)
	}
}

func TestHealthHandler_Health_RedisDegraded(t *testing.T) {
	svc := &MockHealthService{
		checkFn: func(_ context.Context) *domain.HealthStats {
			return &domain.HealthStats{
				Status: "degraded",
				DB:     "ok",
				Redis:  "unavailable",
			}
		},
	}

	h := newTestHealthHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}

	var resp domain.HealthStats
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Redis != "unavailable" {
		t.Errorf("expected redis 'unavailable', got %v", resp.Redis)
	}
}

func TestHealthHandler_Health_BothDegraded(t *testing.T) {
	svc := &MockHealthService{
		checkFn: func(_ context.Context) *domain.HealthStats {
			return &domain.HealthStats{
				Status: "degraded",
				DB:     "unavailable",
				Redis:  "unavailable",
			}
		},
	}

	h := newTestHealthHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}

	var resp domain.HealthStats
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.DB != "unavailable" {
		t.Errorf("expected db 'unavailable', got %v", resp.DB)
	}
	if resp.Redis != "unavailable" {
		t.Errorf("expected redis 'unavailable', got %v", resp.Redis)
	}
}
