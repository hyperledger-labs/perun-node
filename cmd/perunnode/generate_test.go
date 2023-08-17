package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSessionConfig(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "session_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Now you can test the function with a clean temporary directory
	err = generateSessionConfig()
	require.NoError(t, err)

	// List of aliases for testing
	aliasesList := []string{"alice", "bob", "api"}

	// Verify the directories and files were created as expected
	for _, alias := range aliasesList {
		dirPath := filepath.Join(tempDir, alias)
		require.DirExists(t, dirPath)

		require.DirExists(t, filepath.Join(dirPath, "database"))
		require.FileExists(t, filepath.Join(dirPath, "idprovider.yaml"))
		require.DirExists(t, filepath.Join(dirPath, "keystore"))

		keystoreFiles, err := ioutil.ReadDir(filepath.Join(dirPath, "keystore"))
		require.NoError(t, err)
		require.Len(t, keystoreFiles, 2)
		for _, file := range keystoreFiles {
			require.True(t, strings.HasPrefix(file.Name(), "UTC"))
		}

		require.FileExists(t, filepath.Join(dirPath, "session.yaml"))
	}
}
