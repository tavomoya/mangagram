package kissmanga

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tavomoya/mangagram/models"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Kissmanga is a struct to attach
// all functionality available within this
// manga source
type Kissmanga struct {
	DB           *models.DatabaseConfig
	ApiURL       string
	ViewMangaURL string
}

// NewKissmanga function returns a pointer to a Kissmanga
// struct that can be used to call all of its methods
func NewKissmanga(db *models.DatabaseConfig) *Kissmanga {
	return &Kissmanga{
		DB:           db,
		ApiURL:       "https://kissmanga.org/Search/SearchSuggest?keyword=%s",
		ViewMangaURL: "https://kissmanga.org%s",
	}
}

// ViewManga method returns a string with
// the Manga's URL
func (k *Kissmanga) ViewManga() string {
	return k.ViewMangaURL
}

// QueryManga method receives a string that refers to the Manga name, it then
// makes a call to the Kissmanga API, and with the results it returns a
// pointer to a ApiQuerySuggestions struct
func (k *Kissmanga) QueryManga(name string) *models.ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	log.Println("Thename to query: ", name)

	escapedName := url.PathEscape(name)

	path := fmt.Sprintf(k.ApiURL, escapedName)
	log.Println("the path: ", path)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Println("There was an error requesting Kissmanga page: ", err)
		return nil
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error calling HTTP URL: ", err)
		return nil
	}

	defer res.Body.Close()

	page, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("There was an error getting suggestions from Kissmanga's API: ", err)
		return nil
	}

	suggestions := new(models.ApiQuerySuggestions)

	page.Find("a.item_search_link").Each(func(idx int, s *goquery.Selection) {

		mangaURL, _ := s.Attr("href")
		manga := models.MangaSuggestions{
			Data:  mangaURL,
			Value: s.Text(),
		}

		suggestions.Suggestions = append(suggestions.Suggestions, manga)
	})

	return suggestions
}

// GetLastMangaChapter method receives the URL to a manga title and returns
// the last chapter published in this URL. An error might be returned if
// no URL is supplied or if it cannot connect to the URL
func (k *Kissmanga) GetLastMangaChapter(mangaURL string) (string, error) {
	if mangaURL == "" {
		log.Println("No manga supplied")
		return "", nil
	}

	req, err := http.NewRequest("GET", mangaURL, nil)
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
		log.Println("there was an error getting the manga page: ", err)
		return "", err
	}

	lastChapter, _ := page.Find("div.listing div div h3 a").First().Attr("href")

	return fmt.Sprintf(k.ViewMangaURL, lastChapter), nil
}

// Subscribe method receives a subscription model, this contains information
// about a User or Group that wants to receive alerts from a certain Manga title.
// The method will save this in a 'Subscription' collection in MongoDB, as well as
// set a value for the lastChapter of the Manga title.
func (k *Kissmanga) Subscribe(subscription *models.Subscription) error {

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
	subscription.MangaFeed = 4

	subscription.LastChapterURL, _ = k.GetLastMangaChapter(subscription.MangaURL)

	_, err := k.DB.MongoClient.Collection("subscription").InsertOne(k.DB.Ctx, subscription)
	if err != nil && !strings.Contains(err.Error(), "subscription_unq") {
		log.Println("There was an error creating new subscription: ", err)
		return err
	}

	return nil
}
