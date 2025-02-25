package main

import (
	"bundeck/internal/api"
	"bundeck/internal/db"
	"bundeck/internal/plugin"
	"bundeck/internal/settings"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"fyne.io/systray"
	_ "modernc.org/sqlite"

	"github.com/gofiber/fiber/v2"
)

//go:embed web/dist
var website embed.FS

//go:embed logo.ico
var winLogo []byte

//go:embed logo.png
var linuxLogo []byte

//go:embed logo.icns
var macLogo []byte

var dbPath = "./plugins.db"

func onReady() {
	settings := settings.LoadSettings()

	initTray(settings)

	pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"
	// Initialize SQLite database
	database, err := sql.Open("sqlite", dbPath+pragmas)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Initialize database schema
	if err := db.InitDB(database); err != nil {
		log.Fatal(err)
	}

	// Initialize dependencies
	store := db.NewPluginStore(database)
	runner, err := plugin.NewRunner()
	if err != nil {
		log.Fatal(err)
	}
	handlers := api.NewHandlers(store, runner)

	// Initialize Fiber app
	app := fiber.New()

	// API routes
	app.Post("/api/plugins", handlers.CreatePlugin)
	app.Get("/api/plugins", handlers.GetAllPlugins)
	app.Get("/api/plugins/:id/image", handlers.GetPluginImage)
	app.Put("/api/plugins/reorder", handlers.UpdatePluginOrder)
	app.Put("/api/plugins/:id/code", handlers.UpdatePluginData)
	app.Delete("/api/plugins/:id", handlers.DeletePlugin)
	app.Post("/api/plugins/:id/run", handlers.RunPlugin)

	// Plugin template routes
	app.Get("/api/plugins/templates", handlers.GetPluginTemplates)
	app.Post("/api/plugins/templates/create", handlers.CreatePluginFromTemplate)

	app.Get("/favicon*", func(c *fiber.Ctx) error {
		return c.SendFile("web/dist/favicon" + c.Params("*"))
	})

	// Serve the embedded index.html at /app
	app.Get("/*", func(c *fiber.Ctx) error {
		content, err := website.ReadFile("web/dist/index.html")
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error reading index.html")
		}
		c.Set("Content-Type", "text/html")
		return c.Send(content)
	})

	// Start server
	log.Fatal(app.Listen("0.0.0.0:" + strconv.Itoa(settings.Port)))
}

func onExit() {
	fmt.Println("closing!")
}

func main() {
	systray.Run(onReady, onExit)
}
