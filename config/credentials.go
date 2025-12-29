package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// CredentialResolver handles environment variable substitution and credentials file loading.
type CredentialResolver struct {
	// envVars stores environment variables (from .env files or system)
	envVars map[string]string

	// credentials stores credentials loaded from credentials files
	credentials map[string]CredentialSource
}

// NewCredentialResolver creates a new credential resolver.
// It automatically loads system environment variables.
func NewCredentialResolver() *CredentialResolver {
	cr := &CredentialResolver{
		envVars:     make(map[string]string),
		credentials: make(map[string]CredentialSource),
	}

	// Load system environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			cr.envVars[parts[0]] = parts[1]
		}
	}

	return cr
}

// LoadEnvFile loads environment variables from a .env file.
// Format: KEY=value (one per line, # for comments).
func (cr *CredentialResolver) LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open env file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("%s:%d: invalid format (expected KEY=value)", path, lineNum)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		cr.envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file %s: %w", path, err)
	}

	return nil
}

// LoadCredentialsFile loads a credentials YAML file.
func (cr *CredentialResolver) LoadCredentialsFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read credentials file %s: %w", path, err)
	}

	var credConfig CredentialsConfig
	if err := yaml.Unmarshal(data, &credConfig); err != nil {
		return fmt.Errorf("failed to parse credentials file %s: %w", path, err)
	}

	// Merge credentials
	for name, cred := range credConfig.Credentials {
		cr.credentials[name] = cred
	}

	return nil
}

// Resolve resolves placeholders in a connection string.
// Supports:
// - ${VAR_NAME} - environment variable
// - ${VAR_NAME:-default} - environment variable with default
// - @credentials:name - reference to credentials file
func (cr *CredentialResolver) Resolve(value string) (string, error) {
	// Handle credentials file reference
	if strings.HasPrefix(value, "@credentials:") {
		sourceName := strings.TrimPrefix(value, "@credentials:")
		cred, exists := cr.credentials[sourceName]
		if !exists {
			return "", fmt.Errorf("credential source '%s' not found", sourceName)
		}
		return cred.Connection, nil
	}

	// Handle environment variable placeholders
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	result := value
	matches := re.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		placeholder := match[0] // ${VAR_NAME} or ${VAR_NAME:-default}
		varExpr := match[1]     // VAR_NAME or VAR_NAME:-default

		// Check for default value syntax
		var varName, defaultValue string
		if strings.Contains(varExpr, ":-") {
			parts := strings.SplitN(varExpr, ":-", 2)
			varName = parts[0]
			defaultValue = parts[1]
		} else {
			varName = varExpr
		}

		// Get value from environment
		varValue, exists := cr.envVars[varName]
		if !exists {
			if defaultValue != "" {
				// Use default value
				result = strings.Replace(result, placeholder, defaultValue, 1)
			} else {
				return "", fmt.Errorf("environment variable '%s' not set and no default provided", varName)
			}
		} else {
			result = strings.Replace(result, placeholder, varValue, 1)
		}
	}

	return result, nil
}

// Sanitize removes sensitive information from strings (for logging).
// Replaces actual credentials with placeholders.
func (cr *CredentialResolver) Sanitize(message string) string {
	result := message

	// Sanitize known credential patterns
	patterns := []struct {
		regex       string
		replacement string
	}{
		// Password in connection strings
		{`:[^:@]+@`, `:***@`},
		// API keys
		{`(key|token|secret)=[^&\s]+`, `$1=***`},
		// Bearer tokens
		{`Bearer\s+[^\s]+`, `Bearer ***`},
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern.regex)
		result = re.ReplaceAllString(result, pattern.replacement)
	}

	return result
}

// GetEnvVar returns an environment variable value.
func (cr *CredentialResolver) GetEnvVar(name string) (string, bool) {
	value, exists := cr.envVars[name]
	return value, exists
}

// SetEnvVar sets an environment variable (useful for testing).
func (cr *CredentialResolver) SetEnvVar(name, value string) {
	cr.envVars[name] = value
}
