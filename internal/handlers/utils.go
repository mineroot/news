package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/mineroot/news/templates"
)

func ErrorHandler(logger *zerolog.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
		}

		switch code {
		case http.StatusNotFound:
			_ = render(c, http.StatusNotFound, templates.Error("Page not found"), "Page not found")
		default:
			_ = render(c, http.StatusInternalServerError, templates.Error("Internal server error"), "Internal server error")
			logger.Error().
				Str("uri", c.Request().RequestURI).
				Err(err).
				Msg("server error")
		}
	}
}

func bindAndValidateRequest[T any](c echo.Context, validate *validator.Validate) (*T, error) {
	var req T
	if err := c.Bind(&req); err != nil {
		return nil, errors.New("invalid input")
	}
	if err := validate.Struct(req); err != nil {
		var errs validator.ValidationErrors
		if !errors.As(err, &errs) {
			panic("must be unreachable")
		}
		b := strings.Builder{}
		b.WriteString("Validation errors:\n")
		for _, fieldErr := range errs {
			b.WriteString(fmt.Sprintf("invalid '%s' value: %s %s\n", fieldErr.Field(), fieldErr.Tag(), fieldErr.Param()))
		}
		return nil, errors.New(b.String())
	}

	return &req, nil
}

func render(c echo.Context, statusCode int, t templ.Component, title string) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	search := c.QueryParam("q")
	if err := templates.Layout(c.Echo().Reverse, title, search).Render(
		templ.WithChildren(c.Request().Context(), t),
		buf,
	); err != nil {
		return err
	}

	return c.HTML(statusCode, buf.String())
}
