package goconf_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/goconf"
)

func cleanup[T interface {
	string | []string
}](path ...T) {
	if len(path) == 0 {
		return
	}

	var mustCleanupFileIfExists func(string) = func(path string) {
		var err error

		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			panic(fmt.Sprintf("failed to remove file at %s: %v", path, err))
		}
	}

	for _, p := range path {
		switch v := any(p).(type) {
		case string:
			mustCleanupFileIfExists(v)
		case []string:
			for _, p := range v {
				mustCleanupFileIfExists(p)
			}
		default:
			panic(fmt.Sprintf("unsupported type for cleanupFiles: %T", p))
		}
	}
}

func fileReplace(filepath string, substrOld, substrNew string) (err error) {
	var content []byte
	if content, err = os.ReadFile(filepath); err != nil {
		err = fmt.Errorf("read file: %w", err)
		return
	}

	var newContent string = strings.ReplaceAll(string(content), substrOld, substrNew)
	if err = os.WriteFile(filepath, []byte(newContent), 0644); err != nil {
		err = fmt.Errorf("write file: %w", err)
		return
	}

	return
}

// TestLoadGenericConfig will:
// 1. Try to load a config file that doesn't exist, with default behavior (reject) -> should error
// 2. Try to load a config file that doesn't exist, with create and reject behavior -> should create the file but still error
// 3. Try to load a config file that doesn't exist, with create and try behavior -> should create the file and then error because required fields are missing
// 4. Fill in the required fields, then try to load again with create and try behavior -> should load successfully
func TestLoadGenericConfig(t *testing.T) {
	type configuration struct {
		Name    string `toml:"name" validate:"required"`
		Address struct {
			Street string `toml:"street" validate:"required"`
			City   string `toml:"city" validate:"required"`
		} `toml:"address" validate:"required"`
	}

	var (
		filePath = "./generic.toml"
		cfg      configuration
		err      error
	)

	// Housekeeping
	cleanup(filePath)
	defer cleanup(filePath)

	// Test starts here \\

	// 1. No file, default behavior (reject)
	cfg, err = goconf.LoadConfig[configuration](filePath)
	assert.EqualError(t, err, fmt.Sprintf("config file does not exist: %s", filePath))
	assert.Empty(t, cfg)

	// 2. No file, create and reject
	cfg, err = goconf.LoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndReject))
	assert.EqualError(t, err, fmt.Sprintf("config file created at %s; please fill it in and try again", filePath))
	assert.Empty(t, cfg)

	// 3. There should be a file, now should fail to load because required fields are missing
	cfg, err = goconf.LoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry))
	assert.Error(t, err)
	assert.Empty(t, cfg)

	// 4. Fill in the file, then should load successfully
	assert.NoError(t, fileReplace(filePath, "name = \"\"", "name = \"John Doe\""))
	assert.NoError(t, fileReplace(filePath, "street = \"\"", "street = \"123 Main St\""))
	assert.NoError(t, fileReplace(filePath, "city = \"\"", "city = \"Anytown\""))

	cfg, err = goconf.LoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry))
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", cfg.Name)
	assert.Equal(t, "123 Main St", cfg.Address.Street)
	assert.Equal(t, "Anytown", cfg.Address.City)
}

// TestMoreComplexConfig will:
// 1. Try to load a more complex config file that doesn't exist, with create and try behavior -> should create the file and then load successfully with defaults
// 2. Modify the file to have an invalid age, then try to load again with create and try behavior -> should error because age is less than 21
func TestMoreComplexConfig(t *testing.T) {
	type configuration struct {
		Name    string `toml:"name" default:"John Doe" validate:"required"`
		Age     int    `toml:"age" default:"30" validate:"required,gte=21"`
		Address struct {
			Street string `toml:"street" default:"123 Main St" validate:"required"`
			City   string `toml:"city" default:"Anytown" validate:"required"`
		} `toml:"address" validate:"required"`
		Hobbies []string `toml:"hobbies" default:"[\"reading\", \"hiking\"]"`
	}

	var (
		filePath = "./complex.toml"
		cfg      configuration
		err      error
	)

	// Housekeeping
	cleanup(filePath)
	defer cleanup(filePath)

	// Test starts here \\

	// Load config with create and try behavior, should create the file and then load successfully with defaults
	cfg, err = goconf.LoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry))
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", cfg.Name)
	assert.Equal(t, 30, cfg.Age)
	assert.Equal(t, "123 Main St", cfg.Address.Street)
	assert.Equal(t, "Anytown", cfg.Address.City)
	assert.Equal(t, []string{"reading", "hiking"}, cfg.Hobbies)

	// Make them unable to drink in America
	assert.NoError(t, fileReplace(filePath, "age = 30", "age = 20"))

	cfg, err = goconf.LoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry))
	assert.Error(t, err)
	assert.Empty(t, cfg)
}

func TestMustLoadConfig(t *testing.T) {
	type configuration struct {
		Name string `toml:"name" validate:"required"`
	}

	var (
		filePath = "./mustload.toml"
		cfg      configuration
	)

	// Housekeeping
	cleanup(filePath)
	defer cleanup(filePath)

	// Test starts here \\

	// 1. No file, should panic
	assert.PanicsWithError(t, fmt.Sprintf("config file does not exist: %s", filePath), func() {
		goconf.MustLoadConfig[configuration](filePath)
	})

	// 2. Create file but don't fill in required fields, should panic
	assert.Panics(t, func() {
		goconf.MustLoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndReject))
	})

	cleanup(filePath) // Clean up the file created in step 2

	// 3. Create file and try to load, should panic because required fields are missing
	assert.Panics(t, func() {
		goconf.MustLoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry))
	})

	// 4. Fill in required fields, then should load successfully
	assert.NoError(t, fileReplace(filePath, "name = \"\"", "name = \"John Doe\""))
	assert.NotPanics(t, func() {
		cfg = goconf.MustLoadConfig[configuration](filePath, goconf.WithNewFileBehavior(goconf.NewFileBehaviorCreateAndTry))
	})

	assert.Equal(t, "John Doe", cfg.Name)
}
