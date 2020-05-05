package models

// MangaSuggestions is a struct used to
// describe a title queried by a user.
type MangaSuggestions struct {
	// Refers to the encoded name
	// used to access the manga page
	Data string `json:"data"`

	// Title of the manga used as a
	// message to the user
	Value string `json:"value"`
}

// ApiQuerySuggestions is a struct used to manage a list
// of manga suggestions.
type ApiQuerySuggestions struct {
	Suggestions []MangaSuggestions `json:"suggestions"`
}

// ManganeloApiResponse refers to the type of response
// that gets returned by the Manganelo API. Some of these
// properties are currently unused but I'm planning on doing
// something with them.
type ManganeloApiResponse struct {
	// Manganelo's internal ID for the manga
	ID string `json:"id"`

	// Encoded name of the manga, used in the
	// Manga's page
	IDEncode string `json:"id_encode"`

	// Title of the manga
	Name string `json:"name"`

	// Author of the manga
	Author string `json:"author"`

	// Last chapter published
	LastChapter string `json:"lastchapter"`
}

// MangaedenApiResponse refers to the type of response
// that gets returned by the Mangaeden's API.
type MangaedenApiResponse struct {
	// (unused) Mangaeden's internal label
	Label string `json:"label"`

	// Manga URL in Mangaeden
	URL string `json:"url"`

	// Title of the manga
	Value string `json:"value"`
}

// MangaFeed is a struct used to
// define feed's information internally.
type MangaFeed struct {
	// Assigned identifier for the feed
	Code int

	// Name of the feed
	Name string

	// Feed's URL
	URL string
}

// MangareaderApiResponse refers to the type of response
// that gets returned by the Manga Reader's API.
type MangareaderApiResponse struct {
	// Manga Reader's internal ID
	ID int64 `json:"id"`

	// Author of the manga
	Author string `json:"author"`

	// Title of the manga
	Name string `json:"name"`

	// Path to the manga URL
	NameUnsigned string `json:"nameunsigned"`
}
