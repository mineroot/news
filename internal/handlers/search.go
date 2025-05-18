package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/mineroot/news/internal/paging"
	"github.com/mineroot/news/internal/route"
	"github.com/mineroot/news/templates"
)

func ViewSearchHandler(repo PostsPaginator) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := strings.TrimSpace(c.QueryParam("q"))
		rq := []rune(q)
		if len(rq) < 3 {
			return render(c, http.StatusOK, templates.Error("Search query must be at least 3 characters"), "Home")
		}

		const pageSize = 4
		paginator, err := paging.NewPaginator(c.QueryParam("page"), pageSize)
		if err != nil {
			params := make(url.Values)
			params.Set("q", c.QueryParam("q"))
			loc := fmt.Sprintf("%s?%s", c.Echo().Reverse(route.ViewSearch), params.Encode())
			return c.Redirect(http.StatusTemporaryRedirect, loc)
		}

		all, total, err := repo.FindAllByQueryWithPagination(c.Request().Context(), paginator, q)
		if err != nil {
			return err
		}

		paginator.SetRealItemsCount(total)
		if paginator.NeedsRedirect() {
			params := make(url.Values)
			params.Set("q", c.QueryParam("q"))
			params.Set("page", strconv.Itoa(paginator.Page()))
			loc := fmt.Sprintf("%s?%s", c.Echo().Reverse(route.ViewSearch), params.Encode())
			return c.Redirect(http.StatusTemporaryRedirect, loc)
		}

		return render(c, http.StatusOK, templates.Home(c.Echo().Reverse, all, paginator, c.QueryParam("q")), "Home")
	}
}
