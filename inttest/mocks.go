package inttest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

type mockCreditScoreService struct {
	score string
	mu    sync.Mutex
}

func (m *mockCreditScoreService) ServeHTTP(resp http.ResponseWriter, _ *http.Request) {
	m.mu.Lock()
	score := m.score
	m.mu.Unlock()
	_, _ = resp.Write([]byte(fmt.Sprintf(`{"score": "%s"}`, score)))
}

func (m *mockCreditScoreService) SetNextScore(score string) {
	m.mu.Lock()
	m.score = score
	m.mu.Unlock()
}

func startMockCreditScoreService(t *testing.T, ctx context.Context, port int) (*mockCreditScoreService, func()) {
	mock := &mockCreditScoreService{
		score: "0",
	}
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mock,
	}
	go func() {
		_ = s.ListenAndServe()
	}()
	return mock, func() {
		_ = s.Shutdown(ctx)
	}
}
