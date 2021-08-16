package mangaeden

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/tavomoya/mangagram/models"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func testQueryServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("term")

		if query == "err" {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if query == "badjson" {
			res := `[{"url": 4}]`
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(res))
			return
		}

		res := `
		[
			{
				"label": "boku no hero academia - IT",
				"url": "/it-manga/boku-no-hero-academia/",
				"value": "boku no hero academia"
			},
			{
				"label": "Boku no Hero Academia - EN",
				"url": "/en-manga/boku-no-hero-academia/",
				"value": "Boku no Hero Academia"
			},
			{
				"label": "Vigilante: Boku no Hero Academia Illegals - EN",
				"url": "/en-manga/vigilante-boku-no-hero-academia-illegals/",
				"value": "Vigilante: Boku no Hero Academia Illegals"
			},
			{
				"label": "Boku no Hero Academia Smash!! - EN",
				"url": "/en-manga/boku-no-hero-academia-smash/",
				"value": "Boku no Hero Academia Smash!!"
			},
			{
				"label": "boku no hero true hero - EN",
				"url": "/en-manga/boku-no-hero-true-hero/",
				"value": "boku no hero true hero"
			}
		]
		`

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(res))
	}))

	return server
}

func testMangaedenReadServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./../../test/mangaeden-reader.html")
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(file)
	}))

	return server
}

func TestViewManga(t *testing.T) {
	is := is.New(t)
	manga := NewMangaeden(nil)

	url := manga.ViewMangaURL
	is.Equal(url, manga.ViewMangaURL)
}

func TestQueryManga(t *testing.T) {
	is := is.New(t)
	manga := NewMangaeden(nil)
	server := testQueryServer()
	defer server.Close()

	manga.ApiURL = server.URL + "/?term=%s"

	t.Run("No manga name", func(t *testing.T) {
		suggestions := manga.QueryManga("")
		is.Equal(suggestions, nil)
	})

	t.Run("API error, non-200 response", func(t *testing.T) {
		suggestions := manga.QueryManga("err")
		is.Equal(suggestions, nil)
	})

	t.Run("Parsing error, incorrect JSON", func(t *testing.T) {
		suggestions := manga.QueryManga("badjson")
		is.Equal(suggestions, nil)
	})

	t.Run("Happy path", func(t *testing.T) {
		suggestions := manga.QueryManga("boku no hero")
		is.True(suggestions != nil)
		is.Equal(len(suggestions.Suggestions), 4)
		is.Equal(suggestions.Suggestions[0].Value, "Boku no Hero Academia")
	})
}

func TestGetLastMangaChapter(t *testing.T) {
	is := is.New(t)

	manga := NewMangaeden(nil)
	server := testMangaedenReadServer()
	defer server.Close()

	t.Run("No manga URL supplied", func(t *testing.T) {
		url, err := manga.GetLastMangaChapter("")
		is.Equal(url, "")
		is.NoErr(err)
	})

	t.Run("Happy path", func(t *testing.T) {
		mangaUrl := server.URL
		expect := "https://mangaeden.com/en/en-manga/boku-no-hero-academia/279/1/"
		url, err := manga.GetLastMangaChapter(mangaUrl)
		is.Equal(url, expect)
		is.NoErr(err)
	})
}

func TestSubscribe(t *testing.T) {
	is := is.New(t)
	opts := &mtest.Options{}
	opts.ClientType(mtest.Mock)
	opts.CollectionName("subscription")
	opts.DatabaseName("mangagram")
	opts.ShareClient(true)

	mt := mtest.New(t, opts)
	defer mt.Close()

	manga := NewMangaeden(&models.DatabaseConfig{
		Ctx:         context.Background(),
		MongoClient: mt.Client.Database("mangagram"),
	})
	server := testMangaedenReadServer()
	defer server.Close()

	t.Run("No manga name supplied", func(t *testing.T) {
		sub := &models.Subscription{}
		err := manga.Subscribe(sub)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "no manga supplied"))
	})

	t.Run("No manga URL supplied", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Naruto",
		}
		err := manga.Subscribe(sub)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "no manga supplied"))
	})

	t.Run("No Chat ID supplied", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Tokyo Ghoul",
			MangaURL:  server.URL,
		}
		err := manga.Subscribe(sub)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "no Chat supplied"))
	})

	t.Run("Error inserting subscription", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Boku no Hero Academia",
			MangaURL:  server.URL,
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Code:    11000,
			Message: "something went wrong",
		}))

		err := manga.Subscribe(sub)
		is.True(err != nil)
	})

	t.Run("Happy path", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Boku no Hero Academia",
			MangaURL:  server.URL,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := manga.Subscribe(sub)
		is.True(err != nil)
	})
}
