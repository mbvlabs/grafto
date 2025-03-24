package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/routes/paths"
)

type Resource struct{}

func newResource() Resource {
	return Resource{}
}

func (r Resource) Sitemap(c echo.Context) error {
	sitemap, err := createSitemap(c)
	if err != nil {
		return err
	}

	return c.XML(http.StatusOK, sitemap)
}

type URL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	ChangeFreq string   `xml:"changefreq"`
	LastMod    string   `xml:"lastmod"`
	Priority   string   `xml:"priority"`
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
