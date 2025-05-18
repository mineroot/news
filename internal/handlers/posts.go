package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mineroot/news/internal/posts"
	"github.com/mineroot/news/internal/route"
	"github.com/mineroot/news/templates"
)

func ViewPostHandler(repo PostFinder) echo.HandlerFunc {
	return func(c echo.Context) error {
		post, err := findPostFromId(c.Request().Context(), repo, c.Param("id"))
		if err != nil {
			return err
		}

		return render(c, http.StatusOK, templates.Post(c.Echo().Reverse, post), post.Title)
	}
}

func ViewCreatePostFormHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		actionUrl := c.Echo().Reverse(route.CreatePost)
		return render(c, http.StatusOK, templates.PostForm(actionUrl, "", "", ""), "Create new post")
	}
}

type postRequest struct {
	Title   string `form:"title" validate:"required,max=100"`
	Content string `form:"content" validate:"required,max=50000"`
}

func CreatePostHandler(repo PostCreator, validate *validator.Validate) echo.HandlerFunc {
	return func(c echo.Context) error {
		req, err := bindAndValidateRequest[postRequest](c, validate)
		if err != nil {
			fp, _ := c.FormParams()
			return render(
				c,
				http.StatusOK,
				templates.PostForm(c.Echo().Reverse(route.CreatePost), fp.Get("title"), fp.Get("content"), err.Error()),
				"Create new post",
			)
		}

		now := time.Now()
		post, err := repo.Create(c.Request().Context(), &posts.Post{
			Title:   req.Title,
			Content: req.Content,
			Created: now,
			Updated: now,
		})
		if err != nil {
			return err
		}

		c.Response().Header().Set("HX-Push-Url", c.Echo().Reverse(route.ViewPost, post.ID.Hex()))
		return render(c, http.StatusOK, templates.Post(c.Echo().Reverse, post), post.Title)
	}
}

func ViewUpdatePostFormHandler(repo PostFinder) echo.HandlerFunc {
	return func(c echo.Context) error {
		post, err := findPostFromId(c.Request().Context(), repo, c.Param("id"))
		if err != nil {
			return err
		}
		actionUrl := c.Echo().Reverse(route.UpdatePost, post.ID.Hex())

		return render(c, http.StatusOK, templates.PostForm(actionUrl, post.Title, post.Content, ""), "Update post")
	}
}

func UpdatePostHandler(repo PostUpdater, validate *validator.Validate) echo.HandlerFunc {
	return func(c echo.Context) error {
		oid, err := bson.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		req, err := bindAndValidateRequest[postRequest](c, validate)
		if err != nil {
			fp, _ := c.FormParams()
			return render(
				c,
				http.StatusOK,
				templates.PostForm(c.Echo().Reverse(route.UpdatePost, oid), fp.Get("title"), fp.Get("content"), err.Error()),
				"Update post",
			)
		}
		post, err := repo.UpdateById(c.Request().Context(), oid, req.Title, req.Content)
		if err != nil {
			return err
		}
		if post == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		c.Response().Header().Set("HX-Push-Url", c.Echo().Reverse(route.ViewPost, post.ID.Hex()))
		return render(c, http.StatusOK, templates.Post(c.Echo().Reverse, post), post.Title)
	}
}

func DeletePostHandler(repo PostDeleter) echo.HandlerFunc {
	return func(c echo.Context) error {
		oid, err := bson.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		post, err := repo.DeleteById(c.Request().Context(), oid)
		if err != nil {
			return err
		}
		if post == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return c.Redirect(http.StatusSeeOther, c.Echo().Reverse(route.ViewHome))
	}
}

func findPostFromId(ctx context.Context, repo PostFinder, id string) (*posts.Post, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}
	post, err := repo.FindById(ctx, oid)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}

	return post, nil
}
