package mangareader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mangagram/models"
	"net/http"
	"net/url"
)

// MangaReader is a struct used to attach
// all functionality available within this
// manga source.
type MangaReader struct {
	ApiURL       string
	ViewMangaURL string
}

// NewMangaReader function returns a pointer to
// a MangaReader struct that can be used to call
// all of its methods.
func NewMangaReader() *MangaReader {
	return &MangaReader{
		ApiURL:       "https://mangareader.pw/search?query=%s",
		ViewMangaURL: "https://mangareader.pw/search?query=%s",
	}
}

// QueryManga method receives a string that refers to the Manga name, it then
// amkes a call to the MangaReader API, and with the results it returns a
// ApiQuerySuggestions struct.
func (m *MangaReader) QueryManga(name string) *models.ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	escapedName := url.QueryEscape(name)

	path := fmt.Sprintf(m.ApiURL, escapedName)

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
