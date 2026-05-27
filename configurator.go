package goconf

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

// createDefaultConfig creates a new config file at the given path with default values for the struct fields.
func createDefaultConfig[ConfigType any](path string, opts _options) (err error) {
	var cfg ConfigType

	// 1. Apply struct defaults
	if err = defaults.Set(&cfg); err != nil {
		err = fmt.Errorf("set defaults: %w", err)
		return
	}

	// NOTE: Do NOT validate here.
	// The default config is allowed to be "invalid" from a required-fields POV;
	// it's just a template for the user to fill in.
	// Validation happens in LoadConfig() when we actually load the file.

	// 2. Create / truncate the file
	var file *os.File
	if file, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
		err = fmt.Errorf("create config file: %w", err)
		return
	}

	defer file.Close()

	// 3. Encode as TOML
	var encoder *toml.Encoder = toml.NewEncoder(file)
	encoder.Indent = strings.Repeat(" ", opts.indentSpaces)
	if err = encoder.Encode(cfg); err != nil {
		err = fmt.Errorf("encode toml: %w", err)
	}

	return
}

// LoadConfig loads a TOML config file into a struct of type ConfigType.
func LoadConfig[ConfigType any](path string, options ...Option) (config ConfigType, err error) {
	var opts _options
	if opts, err = applyOptions(options); err != nil {
		err = fmt.Errorf("apply options: %w", err)
		return
	}

	// Apply struct defaults BEFORE loading TOML (so TOML overrides)
	if err = defaults.Set(&config); err != nil {
		err = fmt.Errorf("set defaults: %w", err)
		return
	}

	// If it doesn't exist, handle that
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			switch opts.newFileBehavior {
			case NewFileBehaviorReject:
				err = fmt.Errorf("config file does not exist: %s", path)
				return
			case NewFileBehaviorCreateAndReject, NewFileBehaviorCreateAndTry:
				if err = createDefaultConfig[ConfigType](path, opts); err != nil {
					err = fmt.Errorf("create default config: %w", err)
					return
				}

				if opts.newFileBehavior == NewFileBehaviorCreateAndReject {
					err = fmt.Errorf("config file created at %s; please fill it in and try again", path)
					return
				}
			default:
				err = fmt.Errorf("invalid new file behavior: %d", opts.newFileBehavior)
			}
		}
	}

	// Decode TOML file into struct
	if _, err = toml.DecodeFile(path, &config); err != nil {
		err = fmt.Errorf("decode toml: %w", err)
		return
	}

	// Validate required fields
	if err = validator.New(validator.WithRequiredStructEnabled()).Struct(config); err != nil {
		err = fmt.Errorf("validate config: %w", err)
		config = *new(ConfigType) // zero out config on validation
		return
	}

	return
}

// MustLoadConfig is a helper that panics if LoadConfig returns an error.
// Useful for cases where you want to load a config at startup and want the program to fail fast if the config is invalid.
func MustLoadConfig[ConfigType any](path string, options ...Option) (config ConfigType) {
	var err error

	if config, err = LoadConfig[ConfigType](path, options...); err != nil {
		panic(err)
	}

	return
}
