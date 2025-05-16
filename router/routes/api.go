package routes

import "net/http"

const (
	apiV1RoutePrefix = "/api/v1"
	apiV1NamePrefix  = "api.v1"
)

var ApiV1 = []Route{}

var Health = Route{
	Name:        apiV1NamePrefix + ".health",
	Path:        apiV1RoutePrefix + "/health",
	Method:      http.MethodGet,
	HandlerName: "AppHealth",
}
