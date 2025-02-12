package db

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	return db
}

func TestInitDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test initial migration
	err := InitDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify schema version
	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to get schema version: %v", err)
	}

	if version != len(migrations) {
		t.Errorf("Expected schema version %d, got %d", len(migrations), version)
	}

	// Verify plugins table structure
	_, err = db.Exec("INSERT INTO plugins (name, code, order_num, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		"test", "code", 1, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
}

func TestPluginStore_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := InitDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	store := NewPluginStore(db)

	// Test Create
	t.Run("Create", func(t *testing.T) {
		plugin := &Plugin{
			Name:     "Test Plugin",
			Code:     "console.log('test')",
			OrderNum: 1,
		}

		err := store.Create(plugin)
		if err != nil {
			t.Fatalf("Failed to create plugin: %v", err)
		}

		if plugin.ID == 0 {
			t.Error("Expected non-zero ID after creation")
		}
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		plugins, err := store.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all plugins: %v", err)
		}

		if len(plugins) != 1 {
			t.Errorf("Expected 1 plugin, got %d", len(plugins))
		}
	})

	// Test GetByID
	t.Run("GetByID", func(t *testing.T) {
		plugin, err := store.GetByID(1)
		if err != nil {
			t.Fatalf("Failed to get plugin by ID: %v", err)
		}

		if plugin.Name != "Test Plugin" {
			t.Errorf("Expected name 'Test Plugin', got '%s'", plugin.Name)
		}
	})

	// Test UpdateCode
	t.Run("UpdateCode", func(t *testing.T) {
		newCode := "console.log('updated')"
		newName := "Updated Plugin"
		err := store.UpdateCode(1, newCode, nil, "", newName)
		if err != nil {
			t.Fatalf("Failed to update plugin code: %v", err)
		}

		plugin, err := store.GetByID(1)
		if err != nil {
			t.Fatalf("Failed to get updated plugin: %v", err)
		}

		if plugin.Code != newCode {
			t.Errorf("Expected code '%s', got '%s'", newCode, plugin.Code)
		}

		if plugin.Name != newName {
			t.Errorf("Expected name '%s', got '%s'", newName, plugin.Name)
		}
	})

	// Test UpdateOrder
	t.Run("UpdateOrder", func(t *testing.T) {
		orders := []struct {
			ID       int `json:"id"`
			OrderNum int `json:"order_num"`
		}{
			{ID: 1, OrderNum: 2},
		}

		err := store.UpdateOrder(orders)
		if err != nil {
			t.Fatalf("Failed to update plugin order: %v", err)
		}

		plugin, err := store.GetByID(1)
		if err != nil {
			t.Fatalf("Failed to get plugin after order update: %v", err)
		}

		if plugin.OrderNum != 2 {
			t.Errorf("Expected order_num 2, got %d", plugin.OrderNum)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(1)
		if err != nil {
			t.Fatalf("Failed to delete plugin: %v", err)
		}

		_, err = store.GetByID(1)
		if err != sql.ErrNoRows {
			t.Errorf("Expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestPluginStore_ImageHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := InitDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	store := NewPluginStore(db)

	// Test Create with image
	t.Run("CreateWithImage", func(t *testing.T) {
		imageType := "image/png"
		plugin := &Plugin{
			Name:      "Image Plugin",
			Code:      "console.log('test')",
			OrderNum:  1,
			Image:     []byte("fake image data"),
			ImageType: &imageType,
		}

		err := store.Create(plugin)
		if err != nil {
			t.Fatalf("Failed to create plugin with image: %v", err)
		}

		retrieved, err := store.GetByID(plugin.ID)
		if err != nil {
			t.Fatalf("Failed to get plugin: %v", err)
		}

		if string(retrieved.Image) != string(plugin.Image) {
			t.Error("Image data mismatch")
		}

		if *retrieved.ImageType != imageType {
			t.Errorf("Expected image type %s, got %s", imageType, *retrieved.ImageType)
		}
	})

	// Test Update with image
	t.Run("UpdateWithImage", func(t *testing.T) {
		newImageType := "image/jpeg"
		newImage := []byte("new image data")

		err := store.UpdateCode(1, "new code", newImage, newImageType, "Updated Name")
		if err != nil {
			t.Fatalf("Failed to update plugin with image: %v", err)
		}

		retrieved, err := store.GetByID(1)
		if err != nil {
			t.Fatalf("Failed to get updated plugin: %v", err)
		}

		if string(retrieved.Image) != string(newImage) {
			t.Error("Updated image data mismatch")
		}

		if *retrieved.ImageType != newImageType {
			t.Errorf("Expected image type %s, got %s", newImageType, *retrieved.ImageType)
		}
	})
}
