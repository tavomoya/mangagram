package manganelo

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
		err := r.ParseForm()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		for key, val := range r.PostForm {
			if key == "searchword" {
				for _, s := range val {
					if s == "err" {
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}

					if s == "badjson" {
						res := `[{"id": 4}]`
						rw.Header().Set("Content-Type", "application/json")
						rw.WriteHeader(http.StatusOK)
						rw.Write([]byte(res))
						return
					}
				}

			}
		}

		res := `
			[
				{
					"id": "4029",
					"id_encode": "read_tokyo_ghoul_manga_online_free4",
					"name": "<span style=\"color: #FF530D;font-weight: bold;\">tokyo<\/span> <span style=\"color: #FF530D;font-weight: bold;\">ghoul<\/span>",
					"nameunsigned": "read_tokyo_ghoul_manga_online_free4",
					"lastchapter": "Chapter 145",
					"image": "https:\/\/avt.mkklcdnv6temp.com\/35\/x\/3-1583469087.jpg",
					"author": "Ishida Sui",
					"link_story": "https:\/\/readmanganato.com\/manga-od955386"
				},
				{
					"id": "569",
					"id_encode": "read_tokyo_ghoulre",
					"name": "<span style=\"color: #FF530D;font-weight: bold;\">tokyo<\/span> <span style=\"color: #FF530D;font-weight: bold;\">ghoul<\/span>:re",
					"nameunsigned": "read_tokyo_ghoulre",
					"lastchapter": "Vol.16 Chapter 179: Song of the Goat",
					"image": "https:\/\/avt.mkklcdnv6temp.com\/21\/m\/1-1583464542.jpg",
					"author": "Ishida Sui",
					"link_story": "https:\/\/readmanganato.com\/manga-er951926"
				},
				{
					"id": "21539",
					"id_encode": "tokyo_ghoul_redrawn",
					"name": "<span style=\"color: #FF530D;font-weight: bold;\">tokyo<\/span> <span style=\"color: #FF530D;font-weight: bold;\">ghoul<\/span>: Redrawn",
					"nameunsigned": "tokyo_ghoul_redrawn",
					"lastchapter": "vol.1 ch.1 : Tragedy",
					"image": "https:\/\/avt.mkklcdnv6temp.com\/23\/d\/14-1583490571.jpg",
					"author": "Ishida Sui",
					"link_story": "https:\/\/readmanganato.com\/manga-vn972896"
				},
				{
					"id": "8037",
					"id_encode": "toukyou_kushu_jack",
					"name": "Toukyou Kushu Jack",
					"nameunsigned": "toukyou_kushu_jack",
					"lastchapter": "vol.1 ch.7",
					"image": "https:\/\/avt.mkklcdnv6temp.com\/12\/b\/6-1583474597.jpg",
					"author": "Ishida Sui",
					"link_story": "https:\/\/readmanganato.com\/manga-cl959394"
				},
				{
					"id": "11227",
					"id_encode": "read_tokyo_ghoul_oneshot",
					"name": "<span style=\"color: #FF530D;font-weight: bold;\">tokyo<\/span> <span style=\"color: #FF530D;font-weight: bold;\">ghoul<\/span> (Oneshot)",
					"nameunsigned": "read_tokyo_ghoul_oneshot",
					"lastchapter": "ch.0 : [Oneshot]",
					"image": "https:\/\/avt.mkklcdnv6temp.com\/23\/g\/8-1583479061.jpg",
					"author": "ISHIDA Sui",
					"link_story": "https:\/\/readmanganato.com\/manga-lb962584"
				}
			]
		`

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(res))
	}))

	return server
}

func testManganeloReadServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./../../test/manganelo-read.html")
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(file)
	}))

	return server
}

func TestViewManga(t *testing.T) {
	is := is.New(t)
	manga := NewManganelo(nil)

	url := manga.ViewMangaURL
	is.Equal(url, manga.ViewMangaURL)
}

func TestQueryManga(t *testing.T) {
	is := is.New(t)
	manga := NewManganelo(nil)
	server := testQueryServer()
	defer server.Close()

	manga.ApiURL = server.URL

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
		suggestions := manga.QueryManga("tokyo_ghoul")
		is.True(suggestions != nil)
		is.Equal(len(suggestions.Suggestions), 5)
		is.Equal(suggestions.Suggestions[0].Value, "Tokyo Ghoul")
	})
}

func TestGetLastMangaChapter(t *testing.T) {
	is := is.New(t)

	manga := NewManganelo(nil)
	server := testManganeloReadServer()
	defer server.Close()

	t.Run("No manga URL supplied", func(t *testing.T) {
		url, err := manga.GetLastMangaChapter("")
		is.Equal(url, "")
		is.NoErr(err)
	})

	t.Run("Happy path", func(t *testing.T) {
		mangaUrl := server.URL
		expect := "https://readmanganato.com/manga-od955386/chapter-145"
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

	manga := NewManganelo(&models.DatabaseConfig{
		Ctx:         context.Background(),
		MongoClient: mt.Client.Database("mangagram"),
	})
	server := testManganeloReadServer()
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
			MangaName: "Tokyo Ghoul",
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
			MangaName: "Tokyo Ghoul",
			MangaURL:  server.URL,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := manga.Subscribe(sub)
		is.True(err != nil)
	})
}
