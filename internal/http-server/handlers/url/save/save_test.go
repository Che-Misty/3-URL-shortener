package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)


func doRequest(t *testing.T, urlSaverMock *mocks.MockURLSaver, url, alias string) (*httptest.ResponseRecorder, save.Response) {
	t.Helper()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, url, alias)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var resp save.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

	return rr, resp
}

func TestSave_Success(t *testing.T) {
	urlSaverMock := mocks.NewMockURLSaver(t)
	urlSaverMock.On("SaveURL", "https://google.com", "test_alias").
		Return("test_alias", nil).
		Once()

	rr, resp := doRequest(t, urlSaverMock, "https://google.com", "test_alias")

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Equal(t, "OK", resp.Status)
	require.Equal(t, "test_alias", resp.Alias)
	require.Empty(t, resp.Error)
}

func TestSave_BadRequest(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		alias     string
		respError string
	}{
		{
			name:      "empty URL",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field",
		},
		{
			name:      "invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewMockURLSaver(t)

			rr, resp := doRequest(t, urlSaverMock, tc.url, tc.alias)

			require.Equal(t, http.StatusBadRequest, rr.Code)
			require.Equal(t, "Error", resp.Status)
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

func TestSave_BadJSON(t *testing.T) {
	urlSaverMock := mocks.NewMockURLSaver(t)
	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(`{"url": "https://google.com"`)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp save.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "failed to decode request", resp.Error)
}

func TestSave_Conflict(t *testing.T) {
	urlSaverMock := mocks.NewMockURLSaver(t)
	urlSaverMock.On("SaveURL", "https://google.com", "test_alias").
		Return("", storage.ErrURLExist).
		Once()

	rr, resp := doRequest(t, urlSaverMock, "https://google.com", "test_alias")

	require.Equal(t, http.StatusConflict, rr.Code)
	require.Equal(t, "Error", resp.Status)
	require.Equal(t, "url already exist", resp.Error)
}

func TestSave_InternalError(t *testing.T) {
	urlSaverMock := mocks.NewMockURLSaver(t)
	urlSaverMock.On("SaveURL", "https://google.com", "test_alias").
		Return("", errors.New("unexpected db error")).
		Once()

	rr, resp := doRequest(t, urlSaverMock, "https://google.com", "test_alias")

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Equal(t, "Error", resp.Status)
	require.Equal(t, "failed to save url", resp.Error)
}
