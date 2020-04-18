package actions

import (
	"mangagram/actions/kissmanga"
	"mangagram/actions/mangaeden"
	"mangagram/actions/manganelo"
	"mangagram/actions/mangareader"
	"mangagram/models"
)

// Available Manga Feeds:
// 1- MangaReader
// 2- Manganelo (default)
// 3- MangaEden
// 4- Kissmanga

// MangaFeedInterface defines the interface to all
// methods in the different manga sources.
type MangaFeedInterface interface {
	QueryManga(string) *models.ApiQuerySuggestions
	ViewManga() string
	Subscribe(subscription *models.Subscription) error
}

// NewMangaInterface function creates a new MangaFeedInterface interface ready
// to use.
func NewMangaInterface(src int, db *models.DatabaseConfig) MangaFeedInterface {

	switch src {
	case 1:
		return mangareader.NewMangaReader(db)
	case 2:
		return manganelo.NewManganelo(db)
	case 3:
		return mangaeden.NewMangaeden(db)
	case 4:
		return kissmanga.NewKissmanga(db)
	default:
		return nil
	}

}
