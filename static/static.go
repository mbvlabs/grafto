package static

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed *
var Files embed.FS

var MainCssFile = GetMainCssFile()

func GetMainCssFile() string {
	filename := "main-dev.css"

	if err := fs.WalkDir(Files, "css", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(d.Name(), "main-prod-") {
			filename = d.Name()
		}

		return nil
	}); err != nil {
		panic("could not walk files dir")
	}

	return filename
}
