package api

import (
	"bundeck/internal/db"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type mockPluginStore struct {
	plugins map[int]*db.Plugin
	nextID  int
}

func newMockPluginStore() *mockPluginStore {
	return &mockPluginStore{
		plugins: make(map[int]*db.Plugin),
		nextID:  1,
	}
}

func (m *mockPluginStore) Create(plugin *db.Plugin) error {
	plugin.ID = m.nextID
	m.nextID++
	m.plugins[plugin.ID] = plugin
	return nil
}

func (m *mockPluginStore) GetAll() ([]db.Plugin, error) {
	var plugins []db.Plugin
	for _, p := range m.plugins {
		plugins = append(plugins, *p)
	}
	return plugins, nil
}

func (m *mockPluginStore) GetByID(id int) (*db.Plugin, error) {
	plugin, ok := m.plugins[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return plugin, nil
}

func (m *mockPluginStore) UpdateCode(id int, code string, image []byte, imageType string, name string) error {
	plugin, ok := m.plugins[id]
	if !ok {
		return sql.ErrNoRows
	}
	plugin.Code = code
	plugin.Image = image
	if imageType != "" {
		plugin.ImageType = &imageType
	}
	plugin.Name = name
	return nil
}

func (m *mockPluginStore) UpdateOrder(orders []struct {
	ID       int `json:"id"`
	OrderNum int `json:"order_num"`
},
) error {
	for _, order := range orders {
		plugin, ok := m.plugins[order.ID]
		if !ok {
			return fmt.Errorf("plugin not found: %d", order.ID)
		}
		plugin.OrderNum = order.OrderNum
	}
	return nil
}

func (m *mockPluginStore) Delete(id int) error {
	if _, ok := m.plugins[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.plugins, id)
	return nil
}

type mockRunner struct {
	output string
	err    error
}

func (m *mockRunner) Run(id int, code string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.output, nil
}

func setupTest() (*fiber.App, *mockPluginStore, *mockRunner) {
	store := newMockPluginStore()
	runner := &mockRunner{output: "test output"}
	handlers := NewHandlers(store, runner)

	app := fiber.New()
	app.Post("/api/plugins", handlers.CreatePlugin)
	app.Get("/api/plugins", handlers.GetAllPlugins)
	app.Get("/api/plugins/:id/image", handlers.GetPluginImage)
	app.Put("/api/plugins/reorder", handlers.UpdatePluginOrder)
	app.Put("/api/plugins/:id/code", handlers.UpdatePluginData)
	app.Delete("/api/plugins/:id", handlers.DeletePlugin)
	app.Post("/api/plugins/:id/run", handlers.RunPlugin)

	return app, store, runner
}

// Create a small PNG file (1x1 transparent pixel)
var testPNGData = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
	0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x00, 0x00, 0x05,
	0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
	0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func createMultipartRequest(t *testing.T, fields map[string]string, image []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		err := writer.WriteField(key, value)
		if err != nil {
			t.Fatalf("Failed to write field: %v", err)
		}
	}

	if image != nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "image", "test.png"))
		h.Set("Content-Type", "image/png")
		part, err := writer.CreatePart(h)
		if err != nil {
			t.Fatalf("Failed to create image part: %v", err)
		}
		_, err = part.Write(image)
		if err != nil {
			t.Fatalf("Failed to write image data: %v", err)
		}
	}

	writer.Close()
	return body, writer.FormDataContentType()
}

func TestHandlers_CreatePlugin(t *testing.T) {
	app, store, _ := setupTest()

	t.Run("Valid Plugin Creation", func(t *testing.T) {
		fields := map[string]string{
			"name":      "Test Plugin",
			"code":      "console.log('test')",
			"order_num": "1",
		}
		body, contentType := createMultipartRequest(t, fields, testPNGData)

		req := httptest.NewRequest("POST", "/api/plugins", body)
		req.Header.Set("Content-Type", contentType)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusCreated {
			respBody, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", fiber.StatusCreated, resp.StatusCode, string(respBody))
		}

		var plugin db.Plugin
		err = json.NewDecoder(resp.Body).Decode(&plugin)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if plugin.Name != fields["name"] {
			t.Errorf("Expected name %s, got %s", fields["name"], plugin.Name)
		}

		// Verify plugin was stored
		stored, err := store.GetByID(plugin.ID)
		if err != nil {
			t.Fatalf("Failed to get stored plugin: %v", err)
		}
		if stored.Name != fields["name"] {
			t.Errorf("Stored plugin name mismatch: expected %s, got %s", fields["name"], stored.Name)
		}
	})

	t.Run("Invalid Form Data", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/plugins", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "multipart/form-data")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
		}
	})
}

func TestHandlers_GetAllPlugins(t *testing.T) {
	app, store, _ := setupTest()

	// Add test plugins
	store.Create(&db.Plugin{Name: "Plugin 1", Code: "code1", OrderNum: 1})
	store.Create(&db.Plugin{Name: "Plugin 2", Code: "code2", OrderNum: 2})

	req := httptest.NewRequest("GET", "/api/plugins", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	var plugins []db.Plugin
	err = json.NewDecoder(resp.Body).Decode(&plugins)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestHandlers_GetPluginImage(t *testing.T) {
	app, store, _ := setupTest()

	imageType := "image/png"
	plugin := &db.Plugin{
		Name:      "Image Plugin",
		Code:      "code",
		OrderNum:  1,
		Image:     testPNGData,
		ImageType: &imageType,
	}
	store.Create(plugin)

	t.Run("Valid Image Request", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/plugins/%d/image", plugin.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
		}

		if resp.Header.Get("Content-Type") != imageType {
			t.Errorf("Expected content type %s, got %s", imageType, resp.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if !bytes.Equal(body, testPNGData) {
			t.Error("Image data mismatch")
		}
	})

	t.Run("Invalid Plugin ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/plugins/999/image", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
		}
	})
}

func TestHandlers_UpdatePluginData(t *testing.T) {
	app, store, _ := setupTest()

	// Create test plugin
	plugin := &db.Plugin{
		Name:     "Test Plugin",
		Code:     "old code",
		OrderNum: 1,
	}
	store.Create(plugin)

	t.Run("Valid Update", func(t *testing.T) {
		fields := map[string]string{
			"name": "Updated Plugin",
			"code": "new code",
		}
		body, contentType := createMultipartRequest(t, fields, testPNGData)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/plugins/%d/code", plugin.ID), body)
		req.Header.Set("Content-Type", contentType)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", fiber.StatusOK, resp.StatusCode, string(respBody))
		}

		updated, err := store.GetByID(plugin.ID)
		if err != nil {
			t.Fatalf("Failed to get updated plugin: %v", err)
		}

		if updated.Name != fields["name"] {
			t.Errorf("Expected name %s, got %s", fields["name"], updated.Name)
		}

		if updated.Code != fields["code"] {
			t.Errorf("Expected code %s, got %s", fields["code"], updated.Code)
		}
	})
}

func TestHandlers_DeletePlugin(t *testing.T) {
	app, store, _ := setupTest()

	// Create test plugin
	plugin := &db.Plugin{
		Name:     "Test Plugin",
		Code:     "code",
		OrderNum: 1,
	}
	store.Create(plugin)

	t.Run("Valid Delete", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/plugins/%d", plugin.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
		}

		_, err = store.GetByID(plugin.ID)
		if err != sql.ErrNoRows {
			t.Error("Plugin was not deleted")
		}
	})

	t.Run("Invalid Plugin ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/plugins/999", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
		}
	})
}

func TestHandlers_RunPlugin(t *testing.T) {
	app, store, runner := setupTest()

	// Create test plugin
	plugin := &db.Plugin{
		Name:     "Test Plugin",
		Code:     "console.log('test')",
		OrderNum: 1,
	}
	store.Create(plugin)

	t.Run("Successful Run", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/plugins/%d/run", plugin.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedOutput := "test output"
		if string(body) != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, string(body))
		}
	})

	t.Run("Run Error", func(t *testing.T) {
		runner.err = fmt.Errorf("run error")
		defer func() { runner.err = nil }()

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/plugins/%d/run", plugin.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestHandlers_UpdatePluginOrder(t *testing.T) {
	app, store, _ := setupTest()

	// Create test plugins
	plugin1 := &db.Plugin{Name: "Plugin 1", OrderNum: 1}
	plugin2 := &db.Plugin{Name: "Plugin 2", OrderNum: 2}
	store.Create(plugin1)
	store.Create(plugin2)

	t.Run("Valid Order Update", func(t *testing.T) {
		orders := []struct {
			ID       int `json:"id"`
			OrderNum int `json:"order_num"`
		}{
			{ID: plugin1.ID, OrderNum: 2},
			{ID: plugin2.ID, OrderNum: 1},
		}

		body, err := json.Marshal(orders)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("PUT", "/api/plugins/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
		}

		// Verify order update
		p1, _ := store.GetByID(plugin1.ID)
		p2, _ := store.GetByID(plugin2.ID)

		if p1.OrderNum != 2 || p2.OrderNum != 1 {
			t.Error("Plugin order was not updated correctly")
		}
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/plugins/reorder", strings.NewReader("invalid"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to test request: %v", err)
		}

		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
		}
	})
}
