package redirect_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/redirect"
	"url-shortener/internal/http-server/handlers/url/redirect/mocks"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func doRequest(t *testing.T, urlGetterMock *mocks.MockURLGetter, alias string) *httptest.ResponseRecorder {
	t.Helper()

	handler := redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock)

	req := httptest.NewRequest(http.MethodGet, "/url/"+alias, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("alias", alias)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func TestRedirect_Success(t *testing.T) {
	urlGetterMock := mocks.NewMockURLGetter(t)
	urlGetterMock.On("GetURL", "test_alias").
		Return("https://google.com", nil).
		Once()

	rr := doRequest(t, urlGetterMock, "test_alias")

	require.Equal(t, http.StatusFound, rr.Code)
	require.Equal(t, "https://google.com", rr.Header().Get("Location"))
}

func TestRedirect_EmptyAlias(t *testing.T) {
	urlGetterMock := mocks.NewMockURLGetter(t)

	rr := doRequest(t, urlGetterMock, "")

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var body resp.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Equal(t, "invalid request", body.Error)
}

func TestRedirect_NotFound(t *testing.T) {
	urlGetterMock := mocks.NewMockURLGetter(t)
	urlGetterMock.On("GetURL", "missing_alias").
		Return("", storage.ErrURLNotFound).
		Once()

	rr := doRequest(t, urlGetterMock, "missing_alias")

	require.Equal(t, http.StatusNotFound, rr.Code)

	var body resp.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Equal(t, "url not found", body.Error)
}

func TestRedirect_InternalError(t *testing.T) {
	urlGetterMock := mocks.NewMockURLGetter(t)
	urlGetterMock.On("GetURL", "test_alias").
		Return("", errors.New("db is down")).
		Once()

	rr := doRequest(t, urlGetterMock, "test_alias")

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var body resp.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.Equal(t, "internal error", body.Error)
}
