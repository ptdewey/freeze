package freeze

import (
	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
	"github.com/ptdewey/freeze/internal/pretty"
)

type Snapshot = files.Snapshot

type DiffLine = diff.DiffLine

const (
	DiffShared = diff.DiffShared
	DiffOld    = diff.DiffOld
	DiffNew    = diff.DiffNew
)

func Deserialize(raw string) (*Snapshot, error) {
	return files.Deserialize(raw)
}

func SaveSnapshot(snap *Snapshot, state string) error {
	return files.SaveSnapshot(snap, state)
}

func ReadSnapshot(testName string, state string) (*Snapshot, error) {
	return files.ReadSnapshot(testName, state)
}

func SnapshotFileName(testName string) string {
	return files.SnapshotFileName(testName)
}

func Histogram(old, new string) []DiffLine {
	return diff.Histogram(old, new)
}

func NewSnapshotBox(snap *Snapshot) string {
	return pretty.NewSnapshotBox(snap)
}

func DiffSnapshotBox(oldSnap, newSnap *Snapshot) string {
	diffLines := convertDiffLines(diff.Histogram(oldSnap.Content, newSnap.Content))
	return pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)
}
