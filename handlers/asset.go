package handlers

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/maypok86/otter"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/routes/paths"
	"gopkg.in/yaml.v2"
)

const (
	sitemapCacheKey = "assets.sitemap"
	robotsCacheKey  = "assets.robots"
	weekInHours     = 168
	threeInHours    = 72
)

type Assets struct {
	sitemapCache otter.Cache[string, Sitemap]
	assetsCache  otter.Cache[string, string]
}

func newAssets() Assets {
	sitemapCacheBuilder, err := otter.NewBuilder[string, Sitemap](1)
	if err != nil {
		panic(err)
	}

	sitemapCache, err := sitemapCacheBuilder.WithTTL(threeInHours).Build()
	if err != nil {
		panic(err)
	}

	robotsCacheBuilder, err := otter.NewBuilder[string, string](1)
	if err != nil {
		panic(err)
	}

	robotsCache, err := robotsCacheBuilder.WithTTL(weekInHours).Build()
	if err != nil {
		panic(err)
	}

	return Assets{sitemapCache, robotsCache}
}

func (a Assets) Robots(c echo.Context) error {
	if value, ok := a.assetsCache.Get(robotsCacheKey); ok {
		return c.String(http.StatusOK, string(value))
	}

	type robotsTxt struct {
		UserAgent string `yaml:"User-agent"`
		Allow     string `yaml:"Allow"`
		Sitemap   string `yaml:"Sitemap"`
	}

	robots, err := yaml.Marshal(robotsTxt{
		UserAgent: "*",
		Allow:     "/",
		Sitemap: fmt.Sprintf(
			"%s%s",
			config.Cfg.GetFullDomain(),
			paths.Sitemap.URL,
		),
	})
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, string(robots))
}

func (a Assets) Sitemap(c echo.Context) error {
	if value, ok := a.sitemapCache.Get(sitemapCacheKey); ok {
		return c.XML(http.StatusOK, value)
	}

	sitemap, err := createSitemap(c)
	if err != nil {
		return err
	}

	if ok := a.sitemapCache.Set(sitemapCacheKey, sitemap); !ok {
		slog.ErrorContext(
			c.Request().Context(),
			"could not set sitemap cache",
			"error",
			err,
		)
	}

	return c.XML(http.StatusOK, sitemap)
}

type URL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	ChangeFreq string   `xml:"changefreq"`
	LastMod    string   `xml:"lastmod,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URL     []URL    `xml:"url"`
}

func createSitemap(c echo.Context) (Sitemap, error) {
	baseUrl := config.Cfg.GetFullDomain()

	var urls []URL

	urls = append(urls, URL{
		Loc:        baseUrl,
		ChangeFreq: "monthly",
		LastMod:    "2024-10-22T09:43:09+00:00",
		Priority:   "1",
	})

	routes := c.Echo().Routes()
	for _, r := range routes {
		n := paths.Path{Name: r.Name, URL: r.Path}
		switch n {
		case paths.About:
			urls = append(urls, URL{
				Loc: fmt.Sprintf(
					"%s%s",
					baseUrl,
					r.Path,
				),
				ChangeFreq: "monthly",
			})
		}
	}

	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URL:   urls,
	}

	return sitemap, nil
}
