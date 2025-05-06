package views

import (
	"fmt"

	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/views/internal/components"
)

func WithTitle(title string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.Title = fmt.Sprintf("%s | %s", title, hd.Title)
	}
}

func WithDescription(description string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.Description = description
	}
}

func WithImage(image string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.Image = image
	}
}

func WithSlug(slug string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.Slug = fmt.Sprintf("%v%s",
			config.Cfg.GetFullDomain(),
			slug,
		)
	}
}

func WithMetaType(metaType string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.MetaType = metaType
	}
}

func WithStyles(filename string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.StylesheetHref = filename
	}
}

func WithExtraMeta(content, name, property string) components.HeadDataOption {
	return func(hd *components.HeadData) {
		hd.ExtraMeta = append(hd.ExtraMeta, components.MetaContent{
			Content:  content,
			Name:     name,
			Property: property,
		})
	}
}
