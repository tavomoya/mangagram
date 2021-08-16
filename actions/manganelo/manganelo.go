package manganelo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tavomoya/mangagram/models"

	"github.com/PuerkitoBio/goquery"
	strip "github.com/grokify/html-strip-tags-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Manganelo is a struct used to attach
// all functionality available within this
// manga source.
type Manganelo struct {
	DB           *models.DatabaseConfig
	ApiURL       string
	ViewMangaURL string
}

// NewManganelo function returns a pointer to
// a Manganelo struct that can be used to call
// all of its methods.
func NewManganelo(db *models.DatabaseConfig) *Manganelo {
	return &Manganelo{
		DB:           db,
		ApiURL:       "https://manganelo.com/getstorysearchjson",
		ViewMangaURL: "https://manganelo.com/manga/%s",
	}
}

// ViewManga method returns a string with
// the Manga's URL.
func (m *Manganelo) ViewManga() string {
	return m.ViewMangaURL
}

// QueryManga method receives a string that refers to the Manga name, it then
// amkes a call to the Manganelo API, and with the results it returns a
// ApiQuerySuggestions struct.
func (m *Manganelo) QueryManga(name string) *models.ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	res, err := http.PostForm(m.ApiURL, url.Values{"searchword": {name}})
	if err != nil {
		log.Println("There was an error requesting this API: ", err)
		return nil
	}

	if res.StatusCode != http.StatusOK {
		log.Println("Status code was not OK: ", res.StatusCode, res.Request.Body)
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("There was an error reading response body: ", err)
		return nil
	}

	mangas := make([]*models.ManganeloApiResponse, 0)

	err = json.Unmarshal(body, &mangas)
	if err != nil {
		log.Println("There was an error trying to unmarshal response into struct: ", err)
		return nil
	}

	suggestions := new(models.ApiQuerySuggestions)

	for _, manga := range mangas {
		s := models.MangaSuggestions{}

		s.Data = manga.IDEncode
		s.Value = strings.Title(strip.StripTags(manga.Name))

		suggestions.Suggestions = append(suggestions.Suggestions, s)
	}

	return suggestions
}

// Subscribe method receives a subscription model, this contains information about
// a User or Group that wants to receive alerts from a certain Manga title.
// The method will save this in a 'Subscription' collection in MongoDB, as well as
// set a value for the lastChapter of the Manga title.
func (m *Manganelo) Subscribe(subscription *models.Subscription) error {

	// Validate subscription data
	if subscription.MangaName == "" || subscription.MangaURL == "" {
		log.Println("No manga supplied for subscription")
		return errors.New("no manga supplied for subscription")
	}

	if subscription.ChatID == 0 {
		log.Println("No Chat supplied for subscription")
		return errors.New("no Chat supplied for subscription")
	}

	subscription.ID = primitive.NewObjectID()
	subscription.MangaFeed = 2

	subscription.LastChapterURL, _ = m.GetLastMangaChapter(subscription.MangaURL)

	_, err := m.DB.MongoClient.Collection("subscription").InsertOne(m.DB.Ctx, subscription)
	if err != nil && !strings.Contains(err.Error(), "subscription_unq") {
		log.Println("There was an error creating new subscription: ", err)
		return err
	}

	return nil
}

// GetLastMangaChapter method receives the URL to a manga title and returns
// the last chapter published in this URL. An error might be returned if
// no URL is supplied or if it cannot connect to the URL
func (m *Manganelo) GetLastMangaChapter(titleURL string) (string, error) {

	if titleURL == "" {
		log.Println("No title supplied")
		return "", nil
	}

	req, err := http.NewRequest("GET", titleURL, nil)
	if err != nil {
		log.Println("There was an error requesting Kissmanga page: ", err)
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error calling HTTP URL: ", err)
		return "", err
	}

	defer res.Body.Close()

	page, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("There was an error getting the page: ", err)
		return "", err
	}

	lastChaperUrl, _ := page.Find("a.chapter-name").First().Attr("href")

	return lastChaperUrl, nil
}
