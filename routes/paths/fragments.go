package paths

import "fmt"

const FragmentPrefix = "/fragments"

var LoadCsrfToken = register(
	Path{
		Name: "fragments.load_csrf_token",
		URL:  fmt.Sprintf("%s/load-csrf", FragmentPrefix),
	},
)
