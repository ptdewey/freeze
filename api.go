package freeze

import (
	"github.com/ptdewey/freeze/internal/api"
)

type Snapshot = api.Snapshot

type DiffLine = api.DiffLine

const (
	DiffShared = api.DiffShared
	DiffOld    = api.DiffOld
	DiffNew    = api.DiffNew
)

func Deserialize(raw string) (*Snapshot, error) {
	return api.Deserialize(raw)
}

func SaveSnapshot(snap *Snapshot, state string) error {
	return api.SaveSnapshot(snap, state)
}

func ReadSnapshot(testName string, state string) (*Snapshot, error) {
	return api.ReadSnapshot(testName, state)
}

func SnapshotFileName(testName string) string {
	return api.SnapshotFileName(testName)
}

func Histogram(old, new string) []DiffLine {
	return api.Histogram(old, new)
}

func NewSnapshotBox(snap *Snapshot) string {
	return api.NewSnapshotBox(snap)
}

func NewSnapshotBoxFunc(snap *Snapshot) string {
	return api.NewSnapshotBoxFunc(snap)
}

func DiffSnapshotBox(old, new *Snapshot) string {
	return api.DiffSnapshotBox(old, new)
}
