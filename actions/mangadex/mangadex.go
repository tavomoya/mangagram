package mangadex

import (
	"errors"
	"fmt"
	"log"
	"net/url"
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
}

// NewMangaeden function returns a pointer to a Mangaeden
// struct that can be used to call all of its methods
func NewMangadex(db *models.DatabaseConfig) *Mangadex {
	return &Mangadex{
		DB:           db,
		ApiURL:       "https://mangadex.org/search?title=%s",
		ViewMangaURL: "https://mangadex.org%s",
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

	escapedName := url.PathEscape(name)

	path := fmt.Sprintf(m.ApiURL, escapedName)
	log.Println("the path: ", path)

	page, err := goquery.NewDocument(path)
	if err != nil {
		log.Println("There was an error getting to the search pahe: ", err)
		return nil
	}

	suggestions := new(models.ApiQuerySuggestions)
	page.Find("div a.manga_title").Each(func(idx int, selection *goquery.Selection) {

		mangaTitle, _ := selection.Attr("title")
		mangaURL, _ := selection.Attr("href")

		manga := models.MangaSuggestions{
			Data:  fmt.Sprintf(m.ViewMangaURL, mangaURL),
			Value: mangaTitle,
		}

		suggestions.Suggestions = append(suggestions.Suggestions, manga)
	})

	return suggestions
}

func (m *Mangadex) getLastMangaChapter(mangaURL string) (string, error) {

	if mangaURL == "" {
		log.Println("No manga supplied")
		return "", nil
	}

	page, err := goquery.NewDocument(mangaURL)
	if err != nil {
		log.Println("There was an error trying to get to the manga page: ", err)
		return "", err
	}

	lastChapter, _ := page.Find("div[data-lang=1] a.text-truncate").First().Attr("href")

	return fmt.Sprintf(m.ViewMangaURL, lastChapter), nil
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

	subscription.LastChapterURL, _ = m.getLastMangaChapter(subscription.MangaURL)

	_, err := m.DB.MongoClient.Collection("subscription").InsertOne(m.DB.Ctx, subscription)
	if err != nil && !strings.Contains(err.Error(), "subscription_unq") {
		log.Println("There was an error creating new subscription: ", err)
		return err
	}

	return nil
}
