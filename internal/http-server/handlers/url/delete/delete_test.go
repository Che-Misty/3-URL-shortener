package delete_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	del "url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func doRequest(t *testing.T, urlDeleterMock *mocks.MockURLDeleter, alias string) *httptest.ResponseRecorder {
	t.Helper()

	handler := del.New(slogdiscard.NewDiscardLogger(), urlDeleterMock)

	req := httptest.NewRequest(http.MethodDelete, "/url/"+alias, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", alias)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func TestDelete_Success(t *testing.T) {
	urlDeleterMock := mocks.NewMockURLDeleter(t)
	urlDeleterMock.On("DeleteURL", "test_alias").
		Return(int64(1), nil).
		Once()

	rr := doRequest(t, urlDeleterMock, "test_alias")

	require.Equal(t, http.StatusOK, rr.Code)

	var resp del.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "OK", resp.Status)
	require.Equal(t, int64(1), resp.CountDeleted)
}

func TestDelete_EmptyAlias(t *testing.T) {

	urlDeleterMock := mocks.NewMockURLDeleter(t)

	rr := doRequest(t, urlDeleterMock, "")

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp del.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "invalid request", resp.Error)
}

func TestDelete_InternalError(t *testing.T) {
	urlDeleterMock := mocks.NewMockURLDeleter(t)
	urlDeleterMock.On("DeleteURL", "test_alias").
		Return(int64(0), errors.New("db is down")).
		Once()

	rr := doRequest(t, urlDeleterMock, "test_alias")

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var resp del.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "internal server error", resp.Error)
}
