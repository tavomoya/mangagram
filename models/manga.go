package models

// MangaSuggestions is a struct used to
// describe a title queried by a user.
// ---------------------------------------
// Data: Refers to the encoded name used to access the manga page
// Value: Is the Title of the Manga used as a message to the User.
type MangaSuggestions struct {
	Data  string `json:"data"`
	Value string `json:"value"`
}

// ApiQuerySuggestions is a struct used to manage a list
// of manga suggestions.
// ------------------------------------------------------
// Suggestions: Slice of MangaSuggestions
type ApiQuerySuggestions struct {
	Suggestions []MangaSuggestions `json:"suggestions"`
}

// ManganeloApiResponse refers to the type of response
// that gets returned by the Manganelo API. Some of these
// properties are currently unused but I'm planning on doing
// something with them.
// ---------------------------------------------------------
// ID: Manganelo's internal ID for the Manga
// IDEncode: Encoded name of the manga, used in the Manga's page
// Name: Title of the manga
// Author: Author of the manga
// LastChapter: Last chapter of the manga published in Manganelo
type ManganeloApiResponse struct {
	ID          string `json:"id"`
	IDEncode    string `json:"id_encode"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	LastChapter string `json:"lastchapter"`
}

type MangaedenApiResponse struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Value string `json:"value"`
}

type MangaFeed struct {
	Code int
	Name string
	URL  string
}

type MangareaderApiResponse struct {
	ID           int64  `json:"id"`
	Author       string `json:"author"`
	Name         string `json:"name"`
	NameUnsigned string `json:"nameunsigned"`
}
