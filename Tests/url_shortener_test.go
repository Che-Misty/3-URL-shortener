package tests

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/update"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const host = "0.0.0.0:8082"

var (
	authUser     string
	authPassword string
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env")

	authUser = os.Getenv("AUTH_USER")
	authPassword = os.Getenv("AUTH_PASSWORD")

	os.Exit(m.Run())
}

func baseExpect(t *testing.T) *httpexpect.Expect {
	u := url.URL{Scheme: "http", Host: host}
	return httpexpect.Default(t, u.String())
}

func TestURLShortener_HappyPath(t *testing.T) {
	e := baseExpect(t)

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		ContainsKey("alias")
}

//nolint:funlen
func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := baseExpect(t)

			expectedStatus := http.StatusCreated
			if tc.error != "" {
				expectedStatus = http.StatusBadRequest
			}

			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth(authUser, authPassword).
				Expect().Status(expectedStatus).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.error)
				return
			}

			alias := tc.alias
			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}
			testRedirect(t, alias, tc.url)
		})
	}
}

func TestURLShortener_Conflict(t *testing.T) {
	e := baseExpect(t)

	alias := gofakeit.Word() + gofakeit.Word()
	urlToSave := gofakeit.URL()

	e.POST("/url").
		WithJSON(save.Request{URL: urlToSave, Alias: alias}).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusCreated)

	e.POST("/url").
		WithJSON(save.Request{URL: gofakeit.URL(), Alias: alias}).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().
		Value("error").String().IsEqual("url already exist")
}

func TestURLShortener_Unauthorized(t *testing.T) {
	e := baseExpect(t)

	e.POST("/url").
		WithJSON(save.Request{URL: gofakeit.URL(), Alias: gofakeit.Word()}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestURLShortener_Update(t *testing.T) {
	e := baseExpect(t)

	alias := gofakeit.Word() + gofakeit.Word()
	originalURL := gofakeit.URL()
	updatedURL := gofakeit.URL()

	e.POST("/url").
		WithJSON(save.Request{URL: originalURL, Alias: alias}).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusCreated)

	e.PUT("/url").
		WithJSON(update.Request{Alias: alias, NewURL: updatedURL}).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		Value("countUpdated").Number().IsEqual(1)

	testRedirect(t, alias, updatedURL)
}

func TestURLShortener_Delete(t *testing.T) {
	e := baseExpect(t)

	alias := gofakeit.Word() + gofakeit.Word()
	originalURL := gofakeit.URL()

	e.POST("/url").
		WithJSON(save.Request{URL: originalURL, Alias: alias}).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusCreated)

	e.DELETE("/url/"+alias).
		WithBasicAuth(authUser, authPassword).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		Value("countDeleted").Number().IsEqual(1)

	e.GET("/url/" + alias).
		Expect().
		Status(http.StatusNotFound)
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/url/" + alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}
