package actions

import (
	"mangagram/actions/manganelo"
	"mangagram/actions/mangareader"
	"mangagram/models"
)

type MangaFeedInterface interface {
	QueryManga(string) *models.ApiQuerySuggestions
	ViewManga() string
}

func NewMangaInterface(src int) MangaFeedInterface {

	switch src {
	case 1:
		return mangareader.NewMangaReader()
	case 2:
		return manganelo.NewManganelo()
	default:
		return nil
	}

}
