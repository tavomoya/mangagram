package manganelo

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"mangagram/models"
	"net/http"
	"net/url"

	strip "github.com/grokify/html-strip-tags-go"
)

type Manganelo struct {
	ApiURL       string
	ViewMangaURL string
}

func NewManganelo() *Manganelo {
	return &Manganelo{
		ApiURL:       "https://manganelo.com/getstorysearchjson",
		ViewMangaURL: "https://manganelo.com/manga/%s",
	}
}

func (m *Manganelo) ViewManga() string {
	return m.ViewMangaURL
}

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
		s.Value = strip.StripTags(manga.Name)

		suggestions.Suggestions = append(suggestions.Suggestions, s)
	}

	return suggestions
}
