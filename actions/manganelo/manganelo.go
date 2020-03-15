package manganelo

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"mangagram/models"
	"net/http"
	"net/url"
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
)

// Manganelo is a struct used to attach
// all functionality available within this
// manga source.
type Manganelo struct {
	ApiURL       string
	ViewMangaURL string
}

// NewManganelo function returns a pointer to
// a Manganelo struct that can be used to call
// all of its methods.
func NewManganelo() *Manganelo {
	return &Manganelo{
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
