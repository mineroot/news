package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mineroot/news/internal/paging"
	"github.com/mineroot/news/templates"
)

func ViewHomeHandler(repo PostsPaginator) echo.HandlerFunc {
	return func(c echo.Context) error {
		const pageSize = 4
		paginator, err := paging.NewPaginator(c.QueryParam("page"), pageSize)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}

		all, total, err := repo.FindAllByQueryWithPagination(c.Request().Context(), paginator, "")
		if err != nil {
			return err
		}

		paginator.SetRealItemsCount(total)
		if paginator.NeedsRedirect() {
			return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/?page=%d", paginator.Page()))
		}

		return render(c, http.StatusOK, templates.Home(c.Echo().Reverse, all, paginator, ""), "Home")
	}
}
