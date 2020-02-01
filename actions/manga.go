package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const ApiURL string = "https://mangareader.pw/search?query=%s"
const ViewMangaURL = "https://mangareader.pw/manga/%s"

type ApiQuerySuggestions struct {
	Suggestions []MangaSuggestion `json:"suggestions"`
}

type MangaSuggestion struct {
	Data  string `json:"data"`
	Value string `json:"value"`
}

func QueryManga(name string) *ApiQuerySuggestions {

	if name == "" {
		return nil
	}

	escapedName := url.QueryEscape(name)

	path := fmt.Sprintf(ApiURL, escapedName)

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

	suggestions := new(ApiQuerySuggestions)

	err = json.Unmarshal(body, &suggestions)
	if err != nil {
		log.Println("There was an error trying to unmarshal response into a struct: ", err)
		return nil
	}

	return suggestions
}
