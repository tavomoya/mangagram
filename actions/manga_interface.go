package actions

import (
	"mangagram/actions/manganelo"
	"mangagram/actions/mangareader"
	"mangagram/models"
)

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
	default:
		return nil
	}

}
