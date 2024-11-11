package cfgloader

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTempEnvFile creates a temporary .env file with the given content.
func createTempEnvFile(t *testing.T, envContent string) string {
	content := []byte(envContent)
	tmpfile, err := os.CreateTemp("", ".env")
	require.NoError(t, err, "Failed to create temp file")

	_, err = tmpfile.Write(content)
	require.NoError(t, err, "Failed to write to temp file")

	err = tmpfile.Close()
	require.NoError(t, err, "Failed to close temp file")

	return tmpfile.Name()
}

func TestLoadConfigFromENVSuccess(t *testing.T) {
	type TestConfig struct {
		Host        string   `env:"TEST_HOST"`
		Port        int      `env:"TEST_PORT"`
		Debug       bool     `env:"TEST_DEBUG"`
		IntArray    []int    `env:"TEST_INT_ARRAY"`
		StringArray []string `env:"TEST_STRING_ARRAY"`
	}

	// Create a temporary .env file
	tmpEnvPath := createTempEnvFile(t, "TEST_HOST=localhost\nTEST_PORT=8080\nTEST_DEBUG=true\nTEST_INT_ARRAY=1,2,3\nTEST_STRING_ARRAY=one,two,three\n")

	// Change working directory to where the .env file is
	originalWd, _ := os.Getwd()
	err := os.Chdir(filepath.Dir(tmpEnvPath))
	require.NoError(t, err, "Failed to change working directory")
	defer func() {
		err := os.Chdir(originalWd)
		require.Nil(t, err)
	}()

	err = os.Setenv("ENV", "test")
	require.NoError(t, err, "Failed to set ENV variable")

	// Rename the temp file to .env
	filename := ".env"
	env := os.Getenv("ENV")
	if env != "" {
		filename += fmt.Sprintf(".%s", env)
	}
	err = os.Rename(tmpEnvPath, filename)
	require.NoError(t, err, fmt.Sprintf("Failed to rename temp file to %s", filename))
	defer func() {
		err := os.Remove(filename)
		require.Nil(t, err)
	}()

	// Test successful loading
	config, err := LoadConfigFromEnv[TestConfig]()
	require.NoError(t, err, "Failed to load config")
	require.NotNil(t, config, "Config should not be nil")

	// Check if values are correctly loaded
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, true, config.Debug)
	assert.Equal(t, []int{1, 2, 3}, config.IntArray)
	assert.Equal(t, []string{"one", "two", "three"}, config.StringArray)
}

func TestLoadConfigFromEnvSuccess(t *testing.T) {
	type TestConfig struct {
		Host        string   `env:"TEST_HOST"`
		Port        int      `env:"TEST_PORT"`
		Debug       bool     `env:"TEST_DEBUG"`
		IntArray    []int    `env:"TEST_INT_ARRAY"`
		StringArray []string `env:"TEST_STRING_ARRAY"`
	}

	// Create a temporary .env file
	tmpEnvPath := createTempEnvFile(t, "TEST_HOST=localhost\nTEST_PORT=8080\nTEST_DEBUG=true\nTEST_INT_ARRAY=1,2,3\nTEST_STRING_ARRAY=one,two,three\n")

	// Change working directory to where the .env file is
	originalWd, _ := os.Getwd()
	err := os.Chdir(filepath.Dir(tmpEnvPath))
	require.NoError(t, err, "Failed to change working directory")
	defer func() {
		err := os.Chdir(originalWd)
		require.Nil(t, err)
	}()

	// Rename the temp file to .env
	err = os.Rename(tmpEnvPath, ".env")
	require.NoError(t, err, "Failed to rename temp file to .env")
	defer func() {
		err := os.Remove(".env")
		require.Nil(t, err)
	}()

	// Test successful loading
	config, err := LoadConfigFromEnv[TestConfig]()
	require.NoError(t, err, "Failed to load config")
	require.NotNil(t, config, "Config should not be nil")

	// Check if values are correctly loaded
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, true, config.Debug)
	assert.Equal(t, []int{1, 2, 3}, config.IntArray)
	assert.Equal(t, []string{"one", "two", "three"}, config.StringArray)
}

func TestMissingVariable(t *testing.T) {
	type TestConfig struct {
		Host  string `env:"TEST_HOST"`
		Port  int    `env:"TEST_PORT"`
		Empty string `env:"TEST_EMPTY"`
	}

	// Create a temporary .env file
	tmpEnvPath := createTempEnvFile(t, "TEST_HOST=localhost\nTEST_PORT=8080\n")

	// Change working directory to where the .env file is
	originalWd, _ := os.Getwd()
	err := os.Chdir(filepath.Dir(tmpEnvPath))
	require.NoError(t, err, "Failed to change working directory")
	defer func() {
		err := os.Chdir(originalWd)
		require.Nil(t, err)
	}()

	// Rename the temp file to .env
	err = os.Rename(tmpEnvPath, ".env")
	require.NoError(t, err, "Failed to rename temp file to .env")
	defer func() {
		err := os.Remove(".env")
		require.Nil(t, err)
	}()

	// Test missing variable
	_, err = LoadConfigFromEnv[TestConfig]()
	require.Error(t, err, "Should return error")
}

func TestNestedVariable(t *testing.T) {
	type TestConfig struct {
		Host string `env:"TEST_HOST"`
		Port int    `env:"TEST_PORT"`
	}
	type NestedConfig struct {
		Test TestConfig
	}

	// Create a temporary .env file
	tmpEnvPath := createTempEnvFile(t, "TEST_HOST=localhost\nTEST_PORT=8080\n")

	// Change working directory to where the .env file is
	originalWd, _ := os.Getwd()
	err := os.Chdir(filepath.Dir(tmpEnvPath))
	require.NoError(t, err, "Failed to change working directory")
	defer func() {
		err := os.Chdir(originalWd)
		require.Nil(t, err)
	}()

	// Rename the temp file to .env
	err = os.Rename(tmpEnvPath, ".env")
	require.NoError(t, err, "Failed to rename temp file to .env")
	defer func() {
		err := os.Remove(".env")
		require.Nil(t, err)
	}()

	// Test nested variable
	config, err := LoadConfigFromEnv[NestedConfig]()
	require.NoError(t, err, "Failed to load config")
	require.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, "localhost", config.Test.Host)
	assert.Equal(t, 8080, config.Test.Port)
}
