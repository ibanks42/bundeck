package api

import (
	"bundeck/internal/db"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// PluginsFS is the embedded filesystem from main package that contains plugin templates
var PluginsFS fs.FS

// readPluginFile attempts to read a file from the embedded filesystem
func readPluginFile(path string) ([]byte, error) {
	return fs.ReadFile(PluginsFS, path)
}

// PluginStore interface for database operations
type PluginStore interface {
	Create(plugin *db.Plugin) error
	GetAll() ([]db.Plugin, error)
	GetByID(id int) (*db.Plugin, error)
	UpdateCode(id int, code string, image []byte, imageType string, name string, runContinuously bool, intervalSeconds int) error
	UpdateOrder(orders []struct {
		ID       int `json:"id"`
		OrderNum int `json:"order_num"`
	}) error
	Delete(id int) error
}

type PluginResponse struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Code            string  `json:"code"`
	OrderNum        int     `json:"order_num"`
	Image           *string `json:"image"`
	ImageType       *string `json:"image_type"`
	RunContinuously bool    `json:"run_continuously"`
	IntervalSeconds int     `json:"interval_seconds"`
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

	// Get run continuously and interval fields
	runContinuously := false
	if len(form.Value["run_continuously"]) > 0 {
		runContinuously, _ = strconv.ParseBool(form.Value["run_continuously"][0])
	}

	intervalSeconds := 0
	if len(form.Value["interval_seconds"]) > 0 {
		intervalSeconds, _ = strconv.Atoi(form.Value["interval_seconds"][0])
	}

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
		Name:            name,
		Code:            code,
		OrderNum:        orderNum,
		Image:           imageData,
		ImageType:       &imageType,
		RunContinuously: runContinuously,
		IntervalSeconds: intervalSeconds,
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
				ID:              dbPlugins[i].ID,
				Name:            dbPlugins[i].Name,
				Code:            dbPlugins[i].Code,
				OrderNum:        dbPlugins[i].OrderNum,
				Image:           &dataUrl,
				ImageType:       dbPlugins[i].ImageType,
				RunContinuously: dbPlugins[i].RunContinuously,
				IntervalSeconds: dbPlugins[i].IntervalSeconds,
			})
		} else {
			plugins = append(plugins, PluginResponse{
				ID:              dbPlugins[i].ID,
				Name:            dbPlugins[i].Name,
				Code:            dbPlugins[i].Code,
				OrderNum:        dbPlugins[i].OrderNum,
				Image:           nil,
				ImageType:       nil,
				RunContinuously: dbPlugins[i].RunContinuously,
				IntervalSeconds: dbPlugins[i].IntervalSeconds,
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

	// Get run continuously and interval fields
	runContinuously := false
	if len(form.Value["run_continuously"]) > 0 {
		runContinuously, _ = strconv.ParseBool(form.Value["run_continuously"][0])
	}

	intervalSeconds := 0
	if len(form.Value["interval_seconds"]) > 0 {
		intervalSeconds, _ = strconv.Atoi(form.Value["interval_seconds"][0])
	}

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

	if err := h.store.UpdateCode(id, code, imageData, imageType, name, runContinuously, intervalSeconds); err != nil {
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
	templatesPath := "list.json"
	data, err := readPluginFile(templatesPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read plugin templates",
		})
	}

	// Parse templates - now structured by category
	var categorizedTemplates map[string]map[string]interface{}
	if err := json.Unmarshal(data, &categorizedTemplates); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse plugin templates",
		})
	}

	// Convert to flat array as expected by frontend
	templates := []map[string]interface{}{}
	for _, categoryData := range categorizedTemplates {
		if plugins, ok := categoryData["plugins"].([]interface{}); ok {
			for _, plugin := range plugins {
				if pluginMap, ok := plugin.(map[string]interface{}); ok {
					templates = append(templates, pluginMap)
				}
			}
		}
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
	templatesPath := "list.json"
	data, err := readPluginFile(templatesPath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read plugin templates",
		})
	}

	// Parse templates - now structured by category
	var categorizedTemplates map[string]map[string]interface{}
	if err := json.Unmarshal(data, &categorizedTemplates); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse plugin templates",
		})
	}

	// Find the requested template
	var selectedTemplate map[string]interface{}
	templateFound := false

	for _, categoryData := range categorizedTemplates {
		if plugins, ok := categoryData["plugins"].([]interface{}); ok {
			for _, plugin := range plugins {
				if pluginMap, ok := plugin.(map[string]interface{}); ok {
					if id, ok := pluginMap["id"].(string); ok && id == body.TemplateID {
						selectedTemplate = pluginMap
						templateFound = true
						break
					}
				}
			}
			if templateFound {
				break
			}
		}
	}

	if !templateFound {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Template not found",
		})
	}

	// Read the template source file
	sourcePath := selectedTemplate["file"].(string)
	sourceContent, err := readPluginFile(sourcePath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read template source",
		})
	}

	// Replace variables in the source content
	content := string(sourceContent)

	// Process the variables
	for key, value := range body.Variables {
		var stringValue string

		// Handle different variable types
		switch v := value.(type) {
		case []interface{}:
			// Handle various array types
			if len(v) > 0 {
				items := make([]string, len(v))

				// Determine the type of array based on the first element
				switch v[0].(type) {
				case bool:
					// Boolean array
					for i, item := range v {
						boolVal, ok := item.(bool)
						if !ok {
							return c.Status(http.StatusBadRequest).JSON(fiber.Map{
								"error": fmt.Sprintf("Invalid boolean value in array for variable %s", key),
							})
						}
						items[i] = fmt.Sprintf("%v", boolVal)
					}
					stringValue = fmt.Sprintf("[%s]", strings.Join(items, ", "))
				case float64:
					// Number array (JSON numbers come as float64)
					for i, item := range v {
						numVal, ok := item.(float64)
						if !ok {
							return c.Status(http.StatusBadRequest).JSON(fiber.Map{
								"error": fmt.Sprintf("Invalid number value in array for variable %s", key),
							})
						}

						// Use integer format if it's a whole number
						if numVal == float64(int(numVal)) {
							items[i] = fmt.Sprintf("%d", int(numVal))
						} else {
							items[i] = fmt.Sprintf("%g", numVal)
						}
					}
					stringValue = fmt.Sprintf("[%s]", strings.Join(items, ", "))
				default:
					// String array (or mixed, default to strings)
					for i, item := range v {
						strVal, ok := item.(string)
						if !ok {
							// Convert to string if it's not already
							strVal = fmt.Sprintf("%v", item)
						}
						items[i] = fmt.Sprintf("%q", strVal)
					}
					stringValue = fmt.Sprintf("[%s]", strings.Join(items, ", "))
				}
			} else {
				// Empty array
				stringValue = "[]"
			}
		case bool:
			// Boolean value
			stringValue = fmt.Sprintf("%v", v)
		case string:
			// String value
			stringValue = fmt.Sprintf("%q", v)
		case float64:
			// Number value (JSON numbers are decoded as float64)
			if float64(int(v)) == v {
				// If it's a whole number, format as integer
				stringValue = fmt.Sprintf("%d", int(v))
			} else {
				stringValue = fmt.Sprintf("%g", v)
			}
		default:
			// Other types
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

	// Get the run_continuously and interval_seconds values if they were provided
	runContinuously := false
	intervalSeconds := 0
	if c.Get("run_continuously") == "true" {
		runContinuously = true
	}
	if intervalVal, err := strconv.Atoi(c.Get("interval_seconds")); err == nil {
		intervalSeconds = intervalVal
	}

	// Create a new plugin
	plugin := &db.Plugin{}

	// Safely get the title
	if title, ok := selectedTemplate["title"].(string); ok {
		plugin.Name = title
	} else if name, ok := selectedTemplate["name"].(string); ok {
		plugin.Name = name
	} else {
		// Fallback to ID if neither title nor name is available
		if id, ok := selectedTemplate["id"].(string); ok {
			plugin.Name = id
		} else {
			plugin.Name = "Plugin from template"
		}
	}

	plugin.Code = content
	plugin.OrderNum = -1 // Will be last in order
	plugin.RunContinuously = runContinuously
	plugin.IntervalSeconds = intervalSeconds

	if err := h.store.Create(plugin); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(plugin)
}
