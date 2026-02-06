package config

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
)

// Global config instance
var Config *EnvConfig

// Env holds all environment variables loaded from .env file
type Env map[string]string

// EnvConfig holds the configuration loaded from .env file
type EnvConfig struct {
	ServerPort            string `env:"SERVER_PORT,required"`
	GitUser               string `env:"GIT_USER,required"`
	GitPAT                string `env:"GIT_PAT,required"`
	CloneFolder           string `env:"CLONE_FOLDER,required"`
	NotesRepo             string `env:"NOTES_REPO,required"`
	VoyageAPIKey          string `env:"VOYAGE_API_KEY,required"`
	OpenAiAPIKey          string `env:"OPENAI_API_KEY,required"`
	VectorStorageFolder   string `env:"VECTOR_STORAGE_FOLDER,required"`
	HardCodedAPIKeyForNow string `env:"HARD_CODED_API_KEY,required"`
}

// InitConfig loads and initializes the global config at startup
func InitConfig() error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}

	Config = &EnvConfig{}
	if err := env.Populate(Config); err != nil {
		return err
	}

	return nil
}

// LoadEnv loads environment variables with OS env vars taking priority over .env file
func LoadEnv() (Env, error) {
	env := make(Env)

	// First, try to load from .env file (if it exists)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	envFilePath := filepath.Join(filepath.Dir(cwd), ".env")
	file, err := os.Open(envFilePath)
	if err == nil {
		// .env file exists, parse it
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse key=value pairs
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

			// Expand ~ to home directory if present
			if strings.HasPrefix(value, "~") {
				value = expandTilde(value)
			}

			env[key] = value
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}
	}
	// If .env file doesn't exist, that's OK - we'll use OS environment variables

	// Override with OS environment variables (these take priority)
	for _, osEnv := range os.Environ() {
		parts := strings.SplitN(osEnv, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			env[key] = value
		}
	}

	return env, nil
}

// Populate populates a struct from environment variables using struct tags
func (e Env) Populate(envConfig interface{}) error {
	configValue := reflect.ValueOf(envConfig)
	if configValue.Kind() != reflect.Ptr || configValue.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer to a struct")
	}

	configValue = configValue.Elem()
	if configValue.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	configType := configValue.Type()
	var missingFields []string

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		fieldValue := configValue.Field(i)

		// Get the env tag
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		// Parse the tag (format: "ENV_KEY" or "ENV_KEY,required")
		parts := strings.Split(tag, ",")
		envKey := parts[0]
		isRequired := len(parts) > 1 && parts[1] == "required"

		// Get the value from environment
		value, exists := e[envKey]

		// Check if required and missing
		if isRequired && (!exists || value == "") {
			missingFields = append(missingFields, fmt.Sprintf("%s (%s)", field.Name, envKey))
			continue
		}

		// Set the field value if it's settable and we have a value
		if fieldValue.CanSet() && exists {
			fieldValue.SetString(value)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// Get retrieves a value from the environment, returning an empty string if not found
func (e Env) Get(key string) string {
	return e[key]
}

// GetRequired retrieves a required value from the environment, returning an error if not found
func (e Env) GetRequired(key string) (string, error) {
	if value, exists := e[key]; exists && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("%s not found or empty in environment", key)
}

// Set adds or updates an environment variable
func (e Env) Set(key, value string) {
	e[key] = value
}

// Has checks if a key exists in the environment
func (e Env) Has(key string) bool {
	_, exists := e[key]
	return exists
}

// Delete removes a key from the environment
func (e Env) Delete(key string) {
	delete(e, key)
}

// expandTilde expands ~ to the user's home directory
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	usr, err := user.Current()
	if err != nil {
		// If we can't get the user, return the path as-is
		return path
	}

	// Replace ~ with the home directory
	if path == "~" {
		return usr.HomeDir
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(usr.HomeDir, path[2:])
	}

	return path
}
