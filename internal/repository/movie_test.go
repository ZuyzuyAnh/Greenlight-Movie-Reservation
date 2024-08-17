package repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"greenlight.zuyanh.net/internal/entity"
	"testing"
)

func TestInsertMovieSuccess(t *testing.T) {
	store := &MovieModel{DB: db}

	movie := entity.Movie{
		Title:   "Avengers",
		Year:    2021,
		Runtime: entity.Runtime(102),
		Genres:  []string{"actions", "superhero"},
	}
	err := store.Insert(&movie)

	require.NoError(t, err)
	assert.Equal(t, movie.Title, "Avengers")
	assert.Equal(t, movie.Year, int32(2021))
	assert.Equal(t, movie.Runtime, entity.Runtime(102))
	assert.Equal(t, movie.Genres, []string{"actions", "superhero"})

	require.NotNil(t, movie.ID)
	require.NotNil(t, movie.Version)

	assert.Greater(t, movie.ID, int64(0))
	assert.Greater(t, movie.Version, int32(0))
}
