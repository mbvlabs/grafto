//go:build unit
// +build unit

package router_test

import (
	"testing"
)

func TestSetupRoutes(t *testing.T) {
	// ctx := context.Background()
	//
	// logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	// 	Level: slog.LevelInfo,
	// }))
	// slog.SetDefault(logger)
	//
	// r, _ := routes.NewRoutes(handlers.Handlers{}, nil).SetupRoutes(ctx)
	//
	// allRegisteredRoutesDeclared := true
	// for _, route := range r.Routes() {
	// 	if strings.Contains(route.Name, "github") {
	// 		continue
	// 	}
	//
	// 	found := false
	// 	for _, path := range paths.GetAllPaths() {
	// 		if path.Name == route.Name {
	// 			found = true
	// 			break
	// 		}
	// 	}
	// 	if !found {
	// 		slog.Error("TestSetupRoutes", "route", route.Name)
	// 		allRegisteredRoutesDeclared = false
	// 		break
	// 	}
	// }
	//
	// if !allRegisteredRoutesDeclared {
	// 	t.Error("Not all registered routes were defined")
	// }
}
