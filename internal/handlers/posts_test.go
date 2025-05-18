package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mineroot/news/internal/handlers"
	"github.com/mineroot/news/internal/posts"
	"github.com/mineroot/news/internal/route"
)

func TestViewPostHandler(t *testing.T) {
	e := echo.New()
	oid := bson.NewObjectID()

	m := NewMockPostFinder(t)
	m.EXPECT().FindById(mock.Anything, oid).Return(&posts.Post{ID: oid}, nil)
	m.EXPECT().FindById(mock.Anything, mock.Anything).Return(nil, nil)

	// success
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/posts/"+oid.Hex(), nil), rec)
	c.SetParamNames("id")
	c.SetParamValues(oid.Hex())
	require.NoError(t, handlers.ViewPostHandler(m)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// not found
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/posts/invalid_id", nil), rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid_id")
	err := handlers.ViewPostHandler(m)(c)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}

func TestCreatePostHandler(t *testing.T) {
	e := echo.New()
	e.GET(route.ViewPost, nil).Name = route.ViewPost

	createdOid := bson.NewObjectID()
	m := NewMockPostCreator(t)
	m.EXPECT().Create(mock.Anything, mock.Anything).Return(&posts.Post{ID: createdOid}, nil)
	validation := validator.New(validator.WithRequiredStructEnabled())

	// success
	rec := httptest.NewRecorder()
	f := make(url.Values)
	f.Set("title", "Post Title")
	f.Set("content", "Some content")
	req := httptest.NewRequest(http.MethodPost, "/posts/new", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, rec)
	require.NoError(t, handlers.CreatePostHandler(m, validation)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/posts/"+createdOid.Hex(), rec.Header().Get("HX-Push-Url"))

	// validation errors
	rec = httptest.NewRecorder()
	f = make(url.Values)
	f.Set("title", "") // empty title
	f.Set("content", "Some content")
	req = httptest.NewRequest(http.MethodPost, "/posts/new", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c = e.NewContext(req, rec)
	require.NoError(t, handlers.CreatePostHandler(m, validation)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Header().Get("HX-Push-Url")) // assert empty header value
}

func TestViewUpdatePostFormHandler(t *testing.T) {
	e := echo.New()
	oid := bson.NewObjectID()

	m := NewMockPostFinder(t)
	m.EXPECT().FindById(mock.Anything, oid).Return(&posts.Post{ID: oid}, nil)
	m.EXPECT().FindById(mock.Anything, mock.Anything).Return(nil, nil)

	// success
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/posts/"+oid.Hex()+"/edit", nil), rec)
	c.SetParamNames("id")
	c.SetParamValues(oid.Hex())
	require.NoError(t, handlers.ViewUpdatePostFormHandler(m)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// not found
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/posts/invalid_id/edit", nil), rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid_id")
	err := handlers.ViewUpdatePostFormHandler(m)(c)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}

func TestUpdatePostHandler(t *testing.T) {
	e := echo.New()
	e.GET(route.ViewPost, nil).Name = route.ViewPost

	oidToUpdate := bson.NewObjectID()
	m := NewMockPostUpdater(t)
	m.EXPECT().UpdateById(mock.Anything, oidToUpdate, mock.Anything, mock.Anything).Return(&posts.Post{ID: oidToUpdate}, nil)
	m.EXPECT().UpdateById(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	validation := validator.New(validator.WithRequiredStructEnabled())

	// success
	rec := httptest.NewRecorder()
	f := make(url.Values)
	f.Set("title", "New Post Title")
	f.Set("content", "New Some content")
	req := httptest.NewRequest(http.MethodPost, "/posts/"+oidToUpdate.Hex()+"/edit", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(oidToUpdate.Hex())
	require.NoError(t, handlers.UpdatePostHandler(m, validation)(c))
	assert.Equal(t, "/posts/"+oidToUpdate.Hex(), rec.Header().Get("HX-Push-Url"))
	assert.Equal(t, http.StatusOK, rec.Code)

	// validation errors
	rec = httptest.NewRecorder()
	f = make(url.Values)
	f.Set("title", "") // empty title
	f.Set("content", "New Some content")
	req = httptest.NewRequest(http.MethodPost, "/posts/"+oidToUpdate.Hex()+"/edit", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(oidToUpdate.Hex())
	require.NoError(t, handlers.UpdatePostHandler(m, validation)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Header().Get("HX-Push-Url")) // assert empty header value

	// not found
	oidNotFound := bson.NewObjectID()
	rec = httptest.NewRecorder()
	f = make(url.Values)
	f.Set("title", "New Post Title")
	f.Set("content", "New Some content")
	req = httptest.NewRequest(http.MethodPost, "/posts/"+oidNotFound.Hex()+"/edit", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(oidNotFound.Hex())
	err := handlers.UpdatePostHandler(m, validation)(c)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}

func TestDeletePostHandler(t *testing.T) {
	e := echo.New()
	e.GET(route.ViewHome, nil).Name = route.ViewHome

	oidToDelete := bson.NewObjectID()
	m := NewMockPostDeleter(t)
	m.EXPECT().DeleteById(mock.Anything, oidToDelete).Return(&posts.Post{ID: oidToDelete}, nil)
	m.EXPECT().DeleteById(mock.Anything, mock.Anything).Return(nil, nil)

	// success
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodPost, "/posts/"+oidToDelete.Hex()+"/delete", nil), rec)
	c.SetParamNames("id")
	c.SetParamValues(oidToDelete.Hex())
	require.NoError(t, handlers.DeletePostHandler(m)(c))
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))

	// not found
	oidNotFound := bson.NewObjectID()
	rec = httptest.NewRecorder()
	c = e.NewContext(httptest.NewRequest(http.MethodGet, "/posts/"+oidNotFound.Hex()+"/delete", nil), rec)
	c.SetParamNames("id")
	c.SetParamValues(oidNotFound.Hex())
	err := handlers.DeletePostHandler(m)(c)
	var httpErr *echo.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}
