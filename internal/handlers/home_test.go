package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mineroot/news/internal/handlers"
	"github.com/mineroot/news/internal/posts"
)

func TestViewHomeHandler(t *testing.T) {
	e := echo.New()

	post := &posts.Post{}
	m := NewMockPostsPaginator(t)
	m.EXPECT().
		FindAllByQueryWithPagination(mock.Anything, mock.Anything, mock.Anything).
		Return([]*posts.Post{post, post, post, post}, 10, nil)

	// success
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	require.NoError(t, handlers.ViewHomeHandler(m)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// page is too large
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/?page=99", nil), rec)
	require.NoError(t, handlers.ViewHomeHandler(m)(c))
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/?page=3", rec.Header().Get("Location")) // redirect to the max allowed page

	// page is invalid
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/?page=first", nil), rec)
	require.NoError(t, handlers.ViewHomeHandler(m)(c))
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location")) // redirect to first page
}
