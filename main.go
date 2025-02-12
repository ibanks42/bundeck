package main

import (
	"bundeck/internal/api"
	"bundeck/internal/db"
	"bundeck/internal/plugin"
	"bundeck/internal/settings"
	"database/sql"
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"fyne.io/systray"

	"github.com/gofiber/fiber/v2"
	_ "modernc.org/sqlite"
)

//go:embed web/dist
var website embed.FS

//go:embed logo.png
var logo []byte

//go:embed logo.icns
var macLogo []byte

func init() {
	if runtime.GOOS == "darwin" {
		// Get path to executable
		exe, err := os.Executable()
		if err == nil {
			appPath := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
			if filepath.Base(appPath) == "BunDeck.app" {
				os.Chdir(appPath)
			}
		}
	}
}

func onReady() {
	systray.SetIcon(logo)
	systray.SetTitle("BunDeck")
	quit := systray.AddMenuItem("Exit", "Exit")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	settings := settings.LoadSettings()

	pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"
	// Initialize SQLite database
	database, err := sql.Open("sqlite", "./plugins.db"+pragmas)
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
	log.Fatal(app.Listen(":" + strconv.Itoa(settings.Port)))
}

func onExit() {
}

func main() {
	systray.Run(onReady, onExit)
}
