package mangadex

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/tavomoya/mangagram/models"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mangaeden is a struct used to attach
// all functionality available within this
// manga source
type Mangadex struct {
	DB           *models.DatabaseConfig
	ApiURL       string
	ViewMangaURL string
	client       *http.Client
}

// NewMangaeden function returns a pointer to a Mangaeden
// struct that can be used to call all of its methods
func NewMangadex(db *models.DatabaseConfig) *Mangadex {
	jar, _ := cookiejar.New(nil)

	return &Mangadex{
		DB:           db,
		ApiURL:       "https://mangadex.org/search?title=%s",
		ViewMangaURL: "https://mangadex.org%s",
		client:       &http.Client{Jar: jar},
	}
}

// ViewManga method returns a string with
// the Manga's URL
func (m *Mangadex) ViewManga() string {
	return m.ViewMangaURL
}

// QueryManga method receives a string that refers to the Manga name, it then
// makes a call to the Mangadex API, and with the results it returns a
// pointer to a ApiQuerySuggestions struct
func (m *Mangadex) QueryManga(name string) *models.ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	log.Println("Thename to query: ", name)

	// Login
	err := m.login()
	if err != nil {
		return nil
	}

	escapedName := url.PathEscape(name)

	path := fmt.Sprintf(m.ApiURL, escapedName)
	log.Println("the path: ", path)

	req := m.getRequest(path)

	res, err := m.client.Do(req)
	if err != nil {
		log.Println("Error getting to search path: ", err)
		return nil
	}

	defer res.Body.Close()

	page, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("There was an error getting to the search pahe: ", err)
		return nil
	}

	suggestions := new(models.ApiQuerySuggestions)
	page.Find("div a.manga_title").Each(func(idx int, selection *goquery.Selection) {
		mangaTitle, _ := selection.Attr("title")
		mangaURL, _ := selection.Attr("href")

		manga := models.MangaSuggestions{
			Data:  mangaURL,
			Value: mangaTitle,
		}

		suggestions.Suggestions = append(suggestions.Suggestions, manga)
	})

	return suggestions
}

func (m *Mangadex) GetLastMangaChapter(mangaURL string) (string, error) {

	if mangaURL == "" {
		log.Println("No manga supplied")
		return "", nil
	}

	// Login
	err := m.login()
	if err != nil {
		return "", nil
	}

	req := m.getRequest(mangaURL)

	res, err := m.client.Do(req)
	if err != nil {
		log.Println("Error getting to search path: ", err)
		return "", nil
	}

	defer res.Body.Close()

	page, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("There was an error trying to get to the manga page: ", err)
		return "", err
	}
	lastChapterURL := ""
	page.Find("div.chapter-row").EachWithBreak(func(idx int, sel *goquery.Selection) bool {

		lang, _ := sel.Attr("data-lang")
		if lang == "1" {
			lastChapterURL, _ = sel.Find("a.text-truncate").First().Attr("href")
			return false
		}
		return true
	})

	return fmt.Sprintf(m.ViewMangaURL, lastChapterURL), nil
}

// Subscribe method receives a subscription model, this contains information
// about a User or Group that wants to receive alerts from a certain Manga title.
// The method will save this in a 'Subscription' collection in MongoDB, as well as
// set a value for the lastChapter of the Manga title.
func (m *Mangadex) Subscribe(subscription *models.Subscription) error {

	// Validate subscription data
	if subscription.MangaName == "" || subscription.MangaURL == "" {
		log.Println("No manga supplied for subscription")
		return errors.New("No manga supplied for subscription")
	}

	if subscription.ChatID == 0 {
		log.Println("No Chat supplied for subscription")
		return errors.New("No Chat supplied for subscription")
	}

	subscription.ID = primitive.NewObjectID()

	subscription.LastChapterURL, _ = m.GetLastMangaChapter(subscription.MangaURL)

	_, err := m.DB.MongoClient.Collection("subscription").InsertOne(m.DB.Ctx, subscription)
	if err != nil && !strings.Contains(err.Error(), "subscription_unq") {
		log.Println("There was an error creating new subscription: ", err)
		return err
	}

	return nil
}

func (m *Mangadex) login() error {

	loginURL := fmt.Sprintf(m.ViewMangaURL, "/ajax/actions.ajax.php?function=login")
	username := os.Getenv("MANGADEX_USERNAME")
	password := os.Getenv("MANGADEX_PASSWORD")

	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)
	writer.WriteField("login_username", username)
	writer.WriteField("login_password", password)

	req, err := http.NewRequest("POST", loginURL, data)
	if err != nil {
		log.Println("Error creating login request: ", err)
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "mangadex-api/4.0.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	res, err := m.client.Do(req)
	if err != nil {
		log.Println("Error logging into Mangadex => ", err)
		return err
	}

	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading mangadex response: ", err)
		return err
	}

	return nil
}

func (m *Mangadex) getRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating login request: ", err)
		return nil
	}
	req.Header.Set("User-Agent", "mangadex-api/4.0.0")

	return req
}
