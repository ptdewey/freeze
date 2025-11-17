package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ptdewey/freeze/internal/files"
)

func TestSnapshotFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestMyFunction", "test_my_function"},
		{"test_another_one", "test_another_one"},
		{"TestCamelCase", "test_camel_case"},
		{"TestWithNumbers123", "test_with_numbers123"},
		{"TestABC", "test_a_b_c"},
		{"test", "test"},
		{"TEST", "t_e_s_t"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := files.SnapshotFileName(tt.input)
			if result != tt.expected {
				t.Errorf("SnapshotFileName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSerializeDeserialize(t *testing.T) {
	snap := &files.Snapshot{
		Version: "1.0.0",
		Name:    "TestExample",
		Content: "test content\nmultiline",
	}

	serialized := snap.Serialize()
	expected := "---\nversion: 1.0.0\ntest_name: TestExample\n---\ntest content\nmultiline"
	if serialized != expected {
		t.Errorf("Serialize():\nexpected:\n%s\n\ngot:\n%s", expected, serialized)
	}

	deserialized, err := files.Deserialize(serialized)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if deserialized.Version != snap.Version {
		t.Errorf("Version mismatch: %s != %s", deserialized.Version, snap.Version)
	}
	if deserialized.Name != snap.Name {
		t.Errorf("Name mismatch: %s != %s", deserialized.Name, snap.Name)
	}
	if deserialized.Content != snap.Content {
		t.Errorf("Content mismatch: %s != %s", deserialized.Content, snap.Content)
	}
}

func TestDeserializeInvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing separators", "no separators here"},
		{"only one separator", "---\nno closing separator"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := files.Deserialize(tt.input)
			if err == nil {
				t.Error("expected error for invalid format")
			}
		})
	}
}

func TestDeserializeValidFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantVer     string
		wantTest    string
		wantContent string
	}{
		{
			"simple",
			"---\nversion: 1.0\ntest_name: Test\n---\ncontent",
			"1.0",
			"Test",
			"content",
		},
		{
			"multiline content",
			"---\nversion: 0.1\ntest_name: MyTest\n---\nline1\nline2\nline3",
			"0.1",
			"MyTest",
			"line1\nline2\nline3",
		},
		{
			"with extra fields",
			"---\nversion: 1.0\ntest_name: Test\nextra: ignored\n---\ncontent",
			"1.0",
			"Test",
			"content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap, err := files.Deserialize(tt.input)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if snap.Version != tt.wantVer {
				t.Errorf("Version = %s, want %s", snap.Version, tt.wantVer)
			}
			if snap.Name != tt.wantTest {
				t.Errorf("Name = %s, want %s", snap.Name, tt.wantTest)
			}
			if snap.Content != tt.wantContent {
				t.Errorf("Content = %s, want %s", snap.Content, tt.wantContent)
			}
		})
	}
}

func TestSaveAndReadSnapshot(t *testing.T) {
	snap := &files.Snapshot{
		Version: "0.1.0",
		Name:    "TestSaveRead",
		Content: "saved content",
	}

	if err := files.SaveSnapshot(snap, "test"); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	read, err := files.ReadSnapshot("TestSaveRead", "test")
	if err != nil {
		t.Fatalf("ReadSnapshot failed: %v", err)
	}

	if read.Content != snap.Content {
		t.Errorf("Content mismatch: %s != %s", read.Content, snap.Content)
	}
	if read.Version != snap.Version {
		t.Errorf("Version mismatch: %s != %s", read.Version, snap.Version)
	}

	cleanupSnapshot(t, "TestSaveRead", "test")
}

func TestReadSnapshotNotFound(t *testing.T) {
	_, err := files.ReadSnapshot("NonExistentTest", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent snapshot")
	}
}

func TestAcceptSnapshot(t *testing.T) {
	newSnap := &files.Snapshot{
		Version: "0.1.0",
		Name:    "TestAccept",
		Content: "new content to accept",
	}

	if err := files.SaveSnapshot(newSnap, "new"); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	if err := files.AcceptSnapshot("TestAccept"); err != nil {
		t.Fatalf("AcceptSnapshot failed: %v", err)
	}

	accepted, err := files.ReadSnapshot("TestAccept", "accepted")
	if err != nil {
		t.Fatalf("ReadSnapshot failed: %v", err)
	}

	if accepted.Content != newSnap.Content {
		t.Errorf("Content mismatch: %s != %s", accepted.Content, newSnap.Content)
	}

	_, err = files.ReadSnapshot("TestAccept", "new")
	if err == nil {
		t.Error("expected error: .new file should be deleted after accept")
	}

	cleanupSnapshot(t, "TestAccept", "accepted")
}

func TestRejectSnapshot(t *testing.T) {
	snap := &files.Snapshot{
		Version: "0.1.0",
		Name:    "TestReject",
		Content: "content to reject",
	}

	if err := files.SaveSnapshot(snap, "new"); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	if err := files.RejectSnapshot("TestReject"); err != nil {
		t.Fatalf("RejectSnapshot failed: %v", err)
	}

	_, err := files.ReadSnapshot("TestReject", "new")
	if err == nil {
		t.Error("expected error: .new file should be deleted after reject")
	}
}

func cleanupSnapshot(t *testing.T, testName, state string) {
	t.Helper()

	root, err := os.Getwd()
	if err != nil {
		t.Logf("cleanup: failed to get cwd: %v", err)
		return
	}

	for root != "/" && root != "" {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}

	fileName := files.SnapshotFileName(testName) + "." + state
	filePath := filepath.Join(root, "__snapshots__", fileName)
	_ = os.Remove(filePath)
}
