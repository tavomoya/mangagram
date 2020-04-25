package mangaeden

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mangagram/models"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mangaeden is a struct used to attach
// all functionality available within this
// manga source
type Mangaeden struct {
	DB           *models.DatabaseConfig
	ApiURL       string
	ViewMangaURL string
}

// NewMangaeden function returns a pointer to a Mangaeden
// struct that can be used to call all of its methods
func NewMangaeden(db *models.DatabaseConfig) *Mangaeden {
	return &Mangaeden{
		DB:           db,
		ApiURL:       "https://mangaeden.com/ajax/search-manga/?term=%s",
		ViewMangaURL: "https://mangaeden.com%s",
	}
}

// ViewManga method returns a string with
// the Manga's URL
func (m *Mangaeden) ViewManga() string {
	return m.ViewMangaURL
}

// QueryManga method receives a string that refers to the Manga name, it then
// makes a call to the Mangaeden API, and with the results it returns a
// pointer to a ApiQuerySuggestions struct
func (m *Mangaeden) QueryManga(name string) *models.ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	log.Println("Thename to query: ", name)

	escapedName := url.QueryEscape(name)

	path := fmt.Sprintf(m.ApiURL, escapedName)

	log.Println("the path: ", path)

	res, err := http.Get(path)
	if err != nil {
		log.Println("There was an error requesting Mangaeden's API: ", err)
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("There was an error reading Mangaeden's response body: ", err)
		return nil
	}

	mangas := make([]*models.MangaedenApiResponse, 0)

	err = json.Unmarshal(body, &mangas)
	if err != nil {
		log.Println("There was an error unmarshlling Mangaeden's JSON response: ", err)
		return nil
	}

	suggestions := new(models.ApiQuerySuggestions)

	for _, manga := range mangas {

		// Let's filter the mangas to only English
		// Mangaeden returns both in english and Italian
		// but I won't use the Italian versions. (yet)
		if strings.Contains(manga.URL, "it-manga") {
			continue
		}

		s := models.MangaSuggestions{
			Data:  manga.URL,
			Value: manga.Value,
		}

		suggestions.Suggestions = append(suggestions.Suggestions, s)
	}

	return suggestions
}

func (m *Mangaeden) getLastMagaChapter(mangaURL string) (string, error) {

	if mangaURL == "" {
		log.Println("No manga supplied")
		return "", nil
	}

	page, err := goquery.NewDocument(mangaURL)
	if err != nil {
		log.Println("There was an error getting the manga page: ", err)
		return "", err
	}

	lastChapter, _ := page.Find("a.chapterLink").First().Attr("href")

	return fmt.Sprintf(m.ViewMangaURL, lastChapter), nil
}

// Subscribe method receives a subscription model, this contains information
// about a User or Group that wants to receive alerts from a certain Manga title.
// The method will save this in a 'Subscription' collection in MongoDB, as well as
// set a value for the lastChapter of the Manga title.
func (m *Mangaeden) Subscribe(subscription *models.Subscription) error {

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

	subscription.LastChapterURL, _ = m.getLastMagaChapter(subscription.MangaURL)

	_, err := m.DB.MongoClient.Collection("subscription").InsertOne(m.DB.Ctx, subscription)
	if err != nil && !strings.Contains(err.Error(), "subscription_unq") {
		log.Println("There was an error creating new subscription: ", err)
		return err
	}

	return nil

}
