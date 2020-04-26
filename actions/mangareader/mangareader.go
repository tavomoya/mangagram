package mangareader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mangagram/models"
	"net/http"
	"net/url"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MangaReader is a struct used to attach
// all functionality available within this
// manga source.
type MangaReader struct {
	DB           *models.DatabaseConfig
	ApiURL       string
	ViewMangaURL string
}

// NewMangaReader function returns a pointer to
// a MangaReader struct that can be used to call
// all of its methods.
func NewMangaReader(db *models.DatabaseConfig) *MangaReader {
	return &MangaReader{
		DB:           db,
		ApiURL:       "https://mangareader.pw/search?query=%s",
		ViewMangaURL: "https://mangareader.pw/manga/%s",
	}
}

// QueryManga method receives a string that refers to the Manga name, it then
// amkes a call to the MangaReader API, and with the results it returns a
// ApiQuerySuggestions struct.
func (m *MangaReader) QueryManga(name string) *models.ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	log.Println("Thename to query: ", name)

	escapedName := url.PathEscape(name)

	path := fmt.Sprintf(m.ApiURL, escapedName)
	log.Println("the path: ", path)

	res, err := http.Get(path)
	if err != nil {
		log.Println("There was an error requesting this API: ", err)
		return nil
	}

	if res.StatusCode != 200 {
		log.Println("Status Code was not OK: ", res.StatusCode, res.Request.URL)
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("There was an error reading response body: ", err)
		return nil
	}

	suggestions := new(models.ApiQuerySuggestions)

	err = json.Unmarshal(body, &suggestions)
	if err != nil {
		log.Println("There was an error trying to unmarshal response into a struct: ", err)
		return nil
	}

	return suggestions
}

// ViewManga method returns a string with
// the Manga's URL.
func (m *MangaReader) ViewManga() string {
	return m.ViewMangaURL
}

// Subscribe method receives a subscription model, this contains information about
// a User or Group that wants to receive alerts from a certain Manga title.
// The method will save this in a 'Subscription' collection in MongoDB, as well as
// set a value for the lastChapter of the Manga title.
func (m *MangaReader) Subscribe(subscription *models.Subscription) error {

	// Validate subscription data
	if subscription.MangaName == "" || subscription.MangaURL == "" {
		log.Println("No manga supplied for subscription")
		return errors.New("No manga supplied for subscription")
	}

	if subscription.UserID < 1 || subscription.ChatID < 1 {
		log.Println("No User or Chat supplied for subscription")
		return errors.New("No User or Chat supplied for subscription")
	}

	subscription.ID = primitive.NewObjectID()

	_, err := m.DB.MongoClient.Collection("subscription").InsertOne(m.DB.Ctx, subscription)
	if err != nil {
		log.Println("There was an error creating new subscription: ", err)
		return err
	}

	return nil
}
