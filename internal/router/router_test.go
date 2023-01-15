package router

import (
	"github.com/Spear5030/YAGopherMart/internal/config"
	"github.com/Spear5030/YAGopherMart/internal/handler"
	"github.com/Spear5030/YAGopherMart/internal/storage"
	"github.com/Spear5030/YAGopherMart/internal/usecase"
	"github.com/Spear5030/YAGopherMart/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()
	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	lg, err := logger.New(true)
	require.NoError(t, err)
	repo, err := storage.New(lg, cfg.Database)
	require.NoError(t, err)
	useCase := usecase.New(lg, repo, cfg.Accrual)
	require.NoError(t, err)
	h := handler.New(lg, useCase)
	r := New(h)
	ts := httptest.NewServer(r)
	defer ts.Close()
	statusCode, _ := testRequest(t, ts, "POST", "/api/user/register", "")
	assert.Equal(t, http.StatusBadRequest, statusCode)
}
