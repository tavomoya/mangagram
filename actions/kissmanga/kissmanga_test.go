package kissmanga

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/tavomoya/mangagram/models"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func testKissMangaQueryServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./../../test/kissmanga.html")
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(file)
	}))

	return server
}

func testKissMangaReadServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./../../test/kissmanga-read.html")
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(file)
	}))

	return server
}

func TestViewManga(t *testing.T) {
	is := is.New(t)
	manga := NewKissmanga(nil)

	url := manga.ViewMangaURL
	is.Equal(url, manga.ViewMangaURL)
}

func TestQueryManga(t *testing.T) {
	is := is.New(t)

	kiss := NewKissmanga(nil)
	server := testKissMangaQueryServer()
	defer server.Close()
	kiss.ApiURL = server.URL + "/%s"

	t.Run("No manga name to query", func(t *testing.T) {
		suggestions := kiss.QueryManga("")
		is.Equal(suggestions, nil)
	})

	t.Run("Happy path", func(t *testing.T) {
		suggestions := kiss.QueryManga("Naruto")
		is.True(suggestions != nil)
		is.Equal(len(suggestions.Suggestions), 5)
		is.Equal(suggestions.Suggestions[0].Value, " Naruto ")
	})
}

func TestGetLastMangaChapter(t *testing.T) {
	is := is.New(t)

	kiss := NewKissmanga(nil)
	server := testKissMangaReadServer()
	defer server.Close()

	t.Run("No manga URL supplied", func(t *testing.T) {
		url, err := kiss.GetLastMangaChapter("")
		is.Equal(url, "")
		is.NoErr(err)
	})

	t.Run("Happy path", func(t *testing.T) {
		mangaUrl := server.URL
		expect := fmt.Sprintf(kiss.ViewMangaURL, "/chapter/manga-ng952689/chapter-700.5")
		url, err := kiss.GetLastMangaChapter(mangaUrl)
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

	kiss := NewKissmanga(&models.DatabaseConfig{
		Ctx:         context.Background(),
		MongoClient: mt.Client.Database("mangagram"),
	})
	server := testKissMangaReadServer()
	defer server.Close()

	t.Run("No manga name supplied", func(t *testing.T) {
		sub := &models.Subscription{}
		err := kiss.Subscribe(sub)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "no manga supplied"))
	})

	t.Run("No manga URL supplied", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Naruto",
		}
		err := kiss.Subscribe(sub)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "no manga supplied"))
	})

	t.Run("No Chat ID supplied", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Naruto",
			MangaURL:  server.URL,
		}
		err := kiss.Subscribe(sub)
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "no Chat supplied"))
	})

	t.Run("Error inserting subscription", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Naruto",
			MangaURL:  server.URL,
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Code:    11000,
			Message: "something went wrong",
		}))

		err := kiss.Subscribe(sub)
		is.True(err != nil)
	})

	t.Run("Happy path", func(t *testing.T) {
		sub := &models.Subscription{
			MangaName: "Naruto",
			MangaURL:  server.URL,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := kiss.Subscribe(sub)
		is.True(err != nil)
	})
}
