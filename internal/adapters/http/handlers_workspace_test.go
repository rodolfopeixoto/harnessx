// SPDX-License-Identifier: MIT

package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func withRegistry(t *testing.T) (string, *Server) {
	t.Helper()
	root := bootstrap(t)
	t.Setenv(constants.EnvHarnessHome, t.TempDir())
	srv, _ := New(Options{Root: root})
	return root, srv
}

func TestWorkspace_AddListSwitchCurrent(t *testing.T) {
	root, srv := withRegistry(t)

	addBody, _ := json.Marshal(map[string]string{"path": root, "name": "Demo", "slug": "demo"})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/workspace/projects", bytes.NewReader(addBody)))
	require.Equal(t, http.StatusCreated, rec.Code)

	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/workspace/projects", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 1)
	require.Equal(t, "demo", list[0]["Slug"])

	swBody, _ := json.Marshal(map[string]string{"ref": "demo"})
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/workspace/switch", bytes.NewReader(swBody)))
	require.Equal(t, http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/workspace/current", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var cur map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &cur))
	require.Equal(t, "cwd", cur["source"]) // root has .harness => cwd resolution wins
}

func TestWorkspace_AddRejectsEmptyPath(t *testing.T) {
	_, srv := withRegistry(t)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/workspace/projects", bytes.NewReader([]byte(`{}`))))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWorkspace_MethodNotAllowed(t *testing.T) {
	_, srv := withRegistry(t)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/workspace/projects", nil))
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/workspace/switch", nil))
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestWorkspace_SwitchUnknownRefFails(t *testing.T) {
	_, srv := withRegistry(t)
	body, _ := json.Marshal(map[string]string{"ref": "no-such"})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/workspace/switch", bytes.NewReader(body)))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
