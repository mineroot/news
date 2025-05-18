package paging_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mineroot/news/internal/paging"
)

func TestNewPaginator_Valid(t *testing.T) {
	p, err := paging.NewPaginator("2", 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, p.Page())
	assert.Equal(t, 10, p.Size())
	assert.Equal(t, 0, p.PagesCount())
	assert.False(t, p.NeedsRedirect())
}

func TestNewPaginator_DefaultPage(t *testing.T) {
	p, err := paging.NewPaginator("", 5)
	assert.NoError(t, err)
	assert.Equal(t, 1, p.Page())
}

func TestNewPaginator_InvalidPage(t *testing.T) {
	_, err := paging.NewPaginator("abc", 5)
	assert.Error(t, err)

	_, err = paging.NewPaginator("0", 5)
	assert.Error(t, err)
}

func TestNewPaginator_InvalidSize(t *testing.T) {
	_, err := paging.NewPaginator("1", 0)
	assert.Error(t, err)
}

func TestSetRealItemsCount_AdjustsPage(t *testing.T) {
	p, _ := paging.NewPaginator("5", 3)
	p.SetRealItemsCount(10) // pagesCount should be ceil(10 / 3) = 4
	assert.Equal(t, 4, p.PagesCount())
	assert.Equal(t, 4, p.Page())
	assert.True(t, p.NeedsRedirect())
}

func TestSetRealItemsCount_ZeroItems(t *testing.T) {
	p, _ := paging.NewPaginator("5", 10)
	p.SetRealItemsCount(0)
	assert.Equal(t, 1, p.Page())
	assert.Equal(t, 1, p.PagesCount())
	assert.True(t, p.NeedsRedirect())
}

func TestSetRealItemsCount_ValidRange(t *testing.T) {
	p, _ := paging.NewPaginator("2", 5)
	p.SetRealItemsCount(20) // pagesCount should be ceil(10 / 3) = 4
	assert.Equal(t, 4, p.PagesCount())
	assert.Equal(t, 2, p.Page())
	assert.False(t, p.NeedsRedirect())
}
