package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	na "github.com/mbvlabs/grafto/assets"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/handlers"
	"github.com/mbvlabs/grafto/routes/paths"
	"github.com/mbvlabs/grafto/static"
)

func assetsRoutes(
	router *echo.Echo,
	handlers handlers.Assets,
) {
	router.GET(paths.Robots.URL, func(c echo.Context) error {
		return handlers.Robots(c)
	}).Name = paths.Robots.Name

	router.GET(paths.Sitemap.URL, func(c echo.Context) error {
		return handlers.Sitemap(c)
	}).Name = paths.Sitemap.Name

	// css
	router.GET(paths.BootstrapGrid.URL, func(c echo.Context) error {
		stylesheet, err := static.Files.ReadFile(
			fmt.Sprintf("css/%s", "bootstrap-v5_3_3.min.css"),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=604800, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			c.Response().
				Header().
				Set("ETag", "v5.3.3")
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.BootstrapGrid.Name

	router.GET(paths.MainCss.URL, func(c echo.Context) error {
		file := strings.Split(paths.MainCss.URL, "/")[2]
		stylesheet, err := static.Files.ReadFile(
			fmt.Sprintf("css/%s", file),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			hash := strings.Split(file, "-")
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=86400, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			c.Response().
				Header().
				Set("ETag", hash[2])
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.MainCss.Name

	router.GET(paths.NewCss.URL, func(c echo.Context) error {
		stylesheet, err := na.Files.ReadFile(
			"css/styles.css",
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=86400, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			// c.Response().
			// 	Header().
			// 	Set("ETag", hash[2])
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.NewCss.Name

	router.GET(paths.NormalizeCss.URL, func(c echo.Context) error {
		stylesheet, err := na.Files.ReadFile(
			"css/normalize.css",
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=86400, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			// c.Response().
			// 	Header().
			// 	Set("ETag", hash[2])
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.NormalizeCss.Name

	router.GET(paths.LayoutCss.URL, func(c echo.Context) error {
		stylesheet, err := na.Files.ReadFile(
			"css/layout.css",
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=86400, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			// c.Response().
			// 	Header().
			// 	Set("ETag", hash[2])
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.LayoutCss.Name

	router.GET(paths.BaseCss.URL, func(c echo.Context) error {
		stylesheet, err := na.Files.ReadFile(
			"css/base.css",
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=86400, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			// c.Response().
			// 	Header().
			// 	Set("ETag", hash[2])
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.BaseCss.Name

	router.GET(paths.NavCss.URL, func(c echo.Context) error {
		stylesheet, err := na.Files.ReadFile(
			"css/nav.css",
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=86400, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			// c.Response().
			// 	Header().
			// 	Set("ETag", hash[2])
		}

		return c.Blob(http.StatusOK, "text/css", stylesheet)
	}).Name = paths.NavCss.Name

	// htmx
	router.GET(paths.HtmxJS.URL, func(c echo.Context) error {
		file := strings.Split(paths.HtmxJS.URL, "/")[2]
		script, err := static.Files.ReadFile(
			fmt.Sprintf("js/%s", file),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			hash := strings.Split(file, "-")
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=604800, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			c.Response().
				Header().
				Set("ETag", hash[1])
		}

		return c.Blob(http.StatusOK, "text/javascript", script)
	}).Name = paths.HtmxJS.Name

	// alpine.js
	router.GET(paths.AlpineJS.URL, func(c echo.Context) error {
		file := strings.Split(paths.AlpineJS.URL, "/")[2]
		script, err := static.Files.ReadFile(
			fmt.Sprintf("js/%s", file),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			hash := strings.Split(file, "-")
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=604800, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
			c.Response().
				Header().
				Set("ETag", hash[1])
		}

		return c.Blob(http.StatusOK, "text/javascript", script)
	}).Name = paths.AlpineJS.Name

	// favicons/images
	router.GET(paths.Favicon.URL, func(c echo.Context) error {
		img, err := static.Files.ReadFile(
			fmt.Sprintf("images/%s", strings.Split(paths.Favicon.URL, "/")[2]),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=604800, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
		}

		return c.Blob(http.StatusOK, "image/png", img)
	}).Name = paths.Favicon.Name

	router.GET(paths.Favicon16.URL, func(c echo.Context) error {
		img, err := static.Files.ReadFile(
			fmt.Sprintf("images/%s", strings.Split(paths.Favicon16.URL, "/")[2]),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=604800, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
		}

		return c.Blob(http.StatusOK, "image/png", img)
	}).Name = paths.Favicon16.Name

	router.GET(paths.Favicon32.URL, func(c echo.Context) error {
		img, err := static.Files.ReadFile(
			fmt.Sprintf("images/%s", strings.Split(paths.Favicon32.URL, "/")[2]),
		)
		if err != nil {
			return err
		}

		if config.Cfg.Environment == config.PROD_ENVIRONMENT {
			c.Response().
				Header().
				Set("Cache-Control", "public, max-age=604800, immutable")
			c.Response().
				Header().
				Set("Vary", "Accept-Encoding")
		}

		return c.Blob(http.StatusOK, "image/png", img)
	}).Name = paths.Favicon32.Name
}
