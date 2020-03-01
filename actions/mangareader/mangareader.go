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

type MangaReader struct {
	ApiURL       string
	ViewMangaURL string
}

func NewMangaReader() *MangaReader {
	return &MangaReader{
		ApiURL:       "https://mangareader.pw/search?query=%s",
		ViewMangaURL: "https://mangareader.pw/search?query=%s",
	}
}

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

func (m *MangaReader) ViewManga() string {
	return m.ViewMangaURL
}
