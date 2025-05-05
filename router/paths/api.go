package paths

import "fmt"

const apiV1Prefix = "/api/v1"

var APIHealth = register(
	Path{Name: "api.health", URL: fmt.Sprintf("%s/health", apiV1Prefix)},
)
