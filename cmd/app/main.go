package main

import (
	"log"
	"os"

	"github.com/MBvisti/grafto/controllers"
	"github.com/MBvisti/grafto/pkg/mail"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/MBvisti/grafto/routes"
	"github.com/MBvisti/grafto/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	v := views.NewViews()

	router := echo.New()
	router.Renderer = v

	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	conn := database.SetupDatabaseConnection(os.Getenv("DATABASE_URL"))
	db := database.New(conn)
	controllers := controllers.NewController(*db)

	postmark := mail.NewPostmark("332f9c6b-aac6-426d-b070-8eadaccce28c")
	mailClient := mail.NewMail(&postmark)

	err := mailClient.Send("confirm_password", mail.ConfirmPassword{Token: "yoyoyoyoyoy"})
	if err != nil {
		panic(err)
	}

	log.Print(err)

	server := routes.NewServer(router, v, controllers)

	server.Start()
}
