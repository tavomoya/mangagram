package models

type MangaSuggestions struct {
	Data  string `json:"data"`
	Value string `json:"value"`
}

type ApiQuerySuggestions struct {
	Suggestions []MangaSuggestions `json:"suggestions"`
}

type ManganeloApiResponse struct {
	ID          string `json:"id"`
	IDEncode    string `json:"id_encode"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	LastChapter string `json:"lastchapter"`
}
