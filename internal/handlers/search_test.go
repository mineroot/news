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
	"github.com/mineroot/news/internal/route"
)

func TestViewSearchHandler(t *testing.T) {
	e := echo.New()
	e.GET(route.ViewSearch, nil).Name = route.ViewSearch

	post := &posts.Post{}
	m := NewMockPostsPaginator(t)
	m.EXPECT().
		FindAllByQueryWithPagination(mock.Anything, mock.Anything, mock.Anything).
		Return([]*posts.Post{post, post, post, post}, 10, nil)

	// success
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/search?page=1&q=query", nil), rec)
	require.NoError(t, handlers.ViewSearchHandler(m)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// page is too large
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/search?page=99&q=query", nil), rec)
	require.NoError(t, handlers.ViewSearchHandler(m)(c))
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/search?page=3&q=query", rec.Header().Get("Location")) // redirect to the max allowed page

	// page is invalid
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/search?page=first&q=query", nil), rec)
	require.NoError(t, handlers.ViewSearchHandler(m)(c))
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/search?q=query", rec.Header().Get("Location")) // redirect to first page

	// search query is less than 3 chars
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/search?page=first&q=ab", nil), rec)
	require.NoError(t, handlers.ViewSearchHandler(m)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Search query must be at least 3 characters")
}
