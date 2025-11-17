package freeze

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found")
		}
		current = parent
	}
}

func getSnapshotDir() (string, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", err
	}

	// TODO: pull this from config.
	// config should allow having snapshot dir at project root (with or w/o subdirs)
	// or in a __snapshots__ dir inside of each package dir
	snapshotDir := filepath.Join(root, "__snapshots__")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", err
	}

	return snapshotDir, nil
}

func SnapshotFileName(testName string) string {
	var result strings.Builder
	for i, r := range testName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	s := result.String()
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

func SaveSnapshot(snap *Snapshot, state string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := SnapshotFileName(snap.TestName) + "." + state
	filePath := filepath.Join(snapshotDir, fileName)

	return os.WriteFile(filePath, []byte(snap.Serialize()), 0644)
}

func ReadSnapshot(testName string, state string) (*Snapshot, error) {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	fileName := SnapshotFileName(testName) + "." + state
	filePath := filepath.Join(snapshotDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Deserialize(string(data))
}

func readAccepted(testName string) (*Snapshot, error) {
	return ReadSnapshot(testName, "accepted")
}

func readNew(testName string) (*Snapshot, error) {
	return ReadSnapshot(testName, "new")
}

func ListNewSnapshots() ([]string, error) {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, err
	}

	var newSnapshots []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".new") {
			name := strings.TrimSuffix(entry.Name(), ".new")
			newSnapshots = append(newSnapshots, name)
		}
	}

	return newSnapshots, nil
}

func AcceptSnapshot(testName string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := SnapshotFileName(testName)
	newPath := filepath.Join(snapshotDir, fileName+".new")
	acceptedPath := filepath.Join(snapshotDir, fileName+".accepted")

	data, err := os.ReadFile(newPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(acceptedPath, data, 0644); err != nil {
		return err
	}

	return os.Remove(newPath)
}

func RejectSnapshot(testName string) error {
	snapshotDir, err := getSnapshotDir()
	if err != nil {
		return err
	}

	fileName := SnapshotFileName(testName) + ".new"
	filePath := filepath.Join(snapshotDir, fileName)

	return os.Remove(filePath)
}
