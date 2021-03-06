package actions

import (
	"github.com/tavomoya/mangagram/actions/kissmanga"
	"github.com/tavomoya/mangagram/actions/mangadex"
	"github.com/tavomoya/mangagram/actions/mangaeden"
	"github.com/tavomoya/mangagram/actions/manganelo"
	"github.com/tavomoya/mangagram/actions/mangareader"
	"github.com/tavomoya/mangagram/models"
)

// Available Manga Feeds:
// 1- MangaReader (default)
// 2- Manganelo
// 3- MangaEden
// 4- Kissmanga
// 5- Mangadex

// MangaFeedInterface defines the interface to all
// methods in the different manga sources.
type MangaFeedInterface interface {
	QueryManga(string) *models.ApiQuerySuggestions
	ViewManga() string
	Subscribe(subscription *models.Subscription) error
	GetLastMangaChapter(string) (string, error)
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
	case 5:
		return mangadex.NewMangadex(db)
	default:
		return nil
	}

}
