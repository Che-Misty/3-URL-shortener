package update_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/update"
	"url-shortener/internal/http-server/handlers/url/update/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func doRequest(t *testing.T, urlUpdaterMock *mocks.MockURLUpdater, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	req := httptest.NewRequest(http.MethodPut, "/url", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	return rr
}

func TestUpdate_Success(t *testing.T) {
	urlUpdaterMock := mocks.NewMockURLUpdater(t)
	urlUpdaterMock.On("UpdateURL", "test_alias", "https://example.com").
		Return(int64(1), nil).
		Once()

	body, err := json.Marshal(update.Request{
		Alias:  "test_alias",
		NewURL: "https://example.com",
	})
	require.NoError(t, err)

	rr := doRequest(t, urlUpdaterMock, body)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp update.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "OK", resp.Status)
	require.Equal(t, int64(1), resp.CountUodated)
}

func TestUpdate_BadRequest(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		newURL    string
		respError string
	}{
		{
			name:      "empty alias",
			alias:     "",
			newURL:    "https://example.com",
			respError: "field Alias is a required field",
		},
		{
			name:      "invalid new_url",
			alias:     "test_alias",
			newURL:    "not a url",
			respError: "field NewURL is not a valid URL",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlUpdaterMock := mocks.NewMockURLUpdater(t)

			body, err := json.Marshal(update.Request{
				Alias:  tc.alias,
				NewURL: tc.newURL,
			})
			require.NoError(t, err)

			rr := doRequest(t, urlUpdaterMock, body)

			require.Equal(t, http.StatusBadRequest, rr.Code)

			var resp update.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

func TestUpdate_BadJSON(t *testing.T) {
	urlUpdaterMock := mocks.NewMockURLUpdater(t)

	rr := doRequest(t, urlUpdaterMock, []byte(`{"alias": "test_alias"`)) // намеренно оборвано

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var resp update.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "failed to decode request", resp.Error)
}

func TestUpdate_InternalError(t *testing.T) {
	urlUpdaterMock := mocks.NewMockURLUpdater(t)
	urlUpdaterMock.On("UpdateURL", "test_alias", "https://example.com").
		Return(int64(0), errors.New("db is down")).
		Once()

	body, err := json.Marshal(update.Request{
		Alias:  "test_alias",
		NewURL: "https://example.com",
	})
	require.NoError(t, err)

	rr := doRequest(t, urlUpdaterMock, body)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var resp update.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "internal server error", resp.Error)
}
