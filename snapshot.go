package freeze

import (
	"fmt"
	"strings"
)

type Snapshot struct {
	Version  string
	TestName string
	Content  string
}

func (s *Snapshot) Serialize() string {
	header := fmt.Sprintf("---\nversion: %s\ntest_name: %s\n---\n", s.Version, s.TestName)
	return header + s.Content
}

func Deserialize(raw string) (*Snapshot, error) {
	parts := strings.SplitN(raw, "---\n", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid snapshot format")
	}

	header := parts[1]
	content := parts[2]

	snap := &Snapshot{
		Content: content,
	}

	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			continue
		}

		key, value := kv[0], kv[1]
		switch key {
		case "version":
			snap.Version = value
		case "test_name":
			snap.TestName = value
		}
	}

	return snap, nil
}
