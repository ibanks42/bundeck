package api

import (
	"bundeck/internal/db"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// PluginStore interface for database operations
type PluginStore interface {
	Create(plugin *db.Plugin) error
	GetAll() ([]db.Plugin, error)
	GetByID(id int) (*db.Plugin, error)
	UpdateCode(id int, code string, image []byte, imageType string, name string) error
	UpdateOrder(orders []struct {
		ID       int `json:"id"`
		OrderNum int `json:"order_num"`
	}) error
	Delete(id int) error
}

type PluginResponse struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	OrderNum  int     `json:"order_num"`
	Image     *string `json:"image"`
	ImageType *string `json:"image_type"`
}

// Runner interface for plugin execution
type Runner interface {
	Run(id int, code string) (string, error)
}

type Handlers struct {
	store  PluginStore
	runner Runner
}

func NewHandlers(store PluginStore, runner Runner) *Handlers {
	return &Handlers{
		store:  store,
		runner: runner,
	}
}

func (h *Handlers) CreatePlugin(c *fiber.Ctx) error {
	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get form fields
	name := form.Value["name"][0]
	code := form.Value["code"][0]
	orderNum, _ := strconv.Atoi(form.Value["order_num"][0])

	var imageData []byte
	var imageType string

	// Handle image upload if present
	if files := form.File["image"]; len(files) > 0 {
		file := files[0]

		// Validate file type
		if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid file type. Only images are allowed.",
			})
		}

		// Open uploaded file
		f, err := file.Open()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process image",
			})
		}
		defer f.Close()

		// Read file data
		imageData, err = io.ReadAll(f)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read image data",
			})
		}

		imageType = file.Header.Get("Content-Type")
	}

	plugin := &db.Plugin{
		Name:      name,
		Code:      code,
		OrderNum:  orderNum,
		Image:     imageData,
		ImageType: &imageType,
	}

	if err := h.store.Create(plugin); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(plugin)
}

func (h *Handlers) GetAllPlugins(c *fiber.Ctx) error {
	dbPlugins, err := h.store.GetAll()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	var plugins []PluginResponse

	// Convert image data to base64 for JSON response
	for i := range dbPlugins {
		if len(dbPlugins[i].Image) > 0 {
			base := base64.StdEncoding.EncodeToString(dbPlugins[i].Image)
			dataUrl := fmt.Sprintf("data:%s;base64,%s", *dbPlugins[i].ImageType, base)
			plugins = append(plugins, PluginResponse{
				ID:        dbPlugins[i].ID,
				Name:      dbPlugins[i].Name,
				Code:      dbPlugins[i].Code,
				OrderNum:  dbPlugins[i].OrderNum,
				Image:     &dataUrl,
				ImageType: dbPlugins[i].ImageType,
			})
		} else {
			plugins = append(plugins, PluginResponse{
				ID:        dbPlugins[i].ID,
				Name:      dbPlugins[i].Name,
				Code:      dbPlugins[i].Code,
				OrderNum:  dbPlugins[i].OrderNum,
				Image:     nil,
				ImageType: nil,
			})
		}
	}

	return c.JSON(plugins)
}

// Add a new handler to serve plugin images
func (h *Handlers) GetPluginImage(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid plugin ID",
		})
	}

	plugin, err := h.store.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Plugin not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(plugin.Image) == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "No image found",
		})
	}

	if plugin.ImageType == nil {
		c.Set("Content-Type", "application/octet-stream")
	} else {
		c.Set("Content-Type", *plugin.ImageType)
	}
	return c.Send(plugin.Image)
}

func (h *Handlers) UpdatePluginData(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid plugin ID",
		})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Get form fields
	code := form.Value["code"][0]
	name := form.Value["name"][0]

	var imageData []byte
	var imageType string

	// Handle image upload if present
	if files := form.File["image"]; len(files) > 0 {
		file := files[0]

		// Validate file type
		if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid file type. Only images are allowed.",
			})
		}

		// Open uploaded file
		f, err := file.Open()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process image",
			})
		}
		defer f.Close()

		// Read file data
		imageData, err = io.ReadAll(f)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read image data",
			})
		}

		imageType = file.Header.Get("Content-Type")
	}

	if err := h.store.UpdateCode(id, code, imageData, imageType, name); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Plugin not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	row, err := h.store.GetByID(id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(row)
}

func (h *Handlers) UpdatePluginOrder(c *fiber.Ctx) error {
	var orders []struct {
		ID       int `json:"id"`
		OrderNum int `json:"order_num"`
	}
	if err := c.BodyParser(&orders); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.store.UpdateOrder(orders); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(http.StatusOK)
}

func (h *Handlers) DeletePlugin(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid plugin ID",
		})
	}

	if err := h.store.Delete(id); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Plugin not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(http.StatusOK)
}

func (h *Handlers) RunPlugin(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid plugin ID",
		})
	}

	plugin, err := h.store.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Plugin not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := h.runner.Run(id, plugin.Code)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendString(result)
}

// GetPluginTemplates returns the list of available plugin templates
func (h *Handlers) GetPluginTemplates(c *fiber.Ctx) error {
	// Read templates from plugins/list.json
	templatesPath := "plugins/list.json"
	data, err := os.ReadFile(templatesPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read plugin templates",
		})
	}

	// Parse templates
	var templates []map[string]interface{}
	if err := json.Unmarshal(data, &templates); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse plugin templates",
		})
	}

	return c.JSON(templates)
}

// CreatePluginFromTemplate creates a new plugin from a template
func (h *Handlers) CreatePluginFromTemplate(c *fiber.Ctx) error {
	// Parse request body
	var body struct {
		TemplateID string                 `json:"templateId"`
		Variables  map[string]interface{} `json:"variables"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Read templates
	templatesPath := "plugins/list.json"
	data, err := os.ReadFile(templatesPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read plugin templates",
		})
	}

	var templates []map[string]interface{}
	if err := json.Unmarshal(data, &templates); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse plugin templates",
		})
	}

	// Find the template
	var template map[string]interface{}
	for _, t := range templates {
		if t["id"].(string) == body.TemplateID {
			template = t
			break
		}
	}
	if template == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Template not found",
		})
	}

	// Read the template source file
	sourcePath := filepath.Join("plugins", template["file"].(string))
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read template source",
		})
	}

	// Replace variables in the source content
	content := string(sourceContent)
	for key, value := range body.Variables {
		var stringValue string
		switch v := value.(type) {
		case []interface{}:
			items := make([]string, len(v))
			for i, item := range v {
				items[i] = fmt.Sprintf("%q", item)
			}
			stringValue = fmt.Sprintf("[%s]", strings.Join(items, ", "))
		case string:
			stringValue = fmt.Sprintf("%q", v)
		case float64: // JSON numbers are decoded as float64
			if float64(int(v)) == v {
				// If it's a whole number, format as integer
				stringValue = fmt.Sprintf("%d", int(v))
			} else {
				stringValue = fmt.Sprintf("%g", v)
			}
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		// Create a more precise regex pattern that matches the exact variable declaration
		pattern := fmt.Sprintf(`(const\s+%s\s*=\s*)([^;]+)(;)`, regexp.QuoteMeta(key))
		re := regexp.MustCompile(pattern)
		if !re.MatchString(content) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Variable %s not found in template", key),
			})
		}
		content = re.ReplaceAllString(content, "${1}"+stringValue+"${3}")
	}

	// Create a new plugin
	plugin := &db.Plugin{
		Name:     template["title"].(string),
		Code:     content,
		OrderNum: 0, // Will be last in order
	}

	if err := h.store.Create(plugin); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(plugin)
}
