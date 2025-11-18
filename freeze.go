package freeze

import (
	"fmt"
	"runtime"

	"github.com/kortschak/utter"
	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
	"github.com/ptdewey/freeze/internal/pretty"
	"github.com/ptdewey/freeze/internal/review"
)

// TODO: probably make this (and other things) configurable
func init() {
	utter.Config.ElideType = true
}

func SnapString(t testingT, content string) {
	t.Helper()
	snap(t, content)
}

func Snap(t testingT, values ...any) {
	t.Helper()
	content := formatValues(values...)
	snap(t, content)
}

func SnapWithTitle(t testingT, title string, values ...any) {
	t.Helper()
	content := formatValues(values...)
	snapWithTitle(t, title, content)
}

func SnapFunc(t testingT, values ...any) {
	t.Helper()
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		fullName := fn.Name()
		parts := len(fullName) - 1
		for i := len(fullName) - 1; i >= 0; i-- {
			if fullName[i] == '.' {
				parts = i + 1
				break
			}
		}
		funcName = fullName[parts:]
	}
	content := formatValues(values...)
	snapWithTitle(t, funcName, content)
}

func getFunctionName() string {
	pc, _, _, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	fullName := fn.Name()
	parts := len(fullName) - 1
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == '.' {
			parts = i + 1
			break
		}
	}

	return fullName[parts:]
}

func snap(t testingT, content string) {
	t.Helper()
	testName := t.Name()
	snapWithTitle(t, testName, content)
}

func snapWithTitle(t testingT, title string, content string) {
	t.Helper()

	_, filePath, _, _ := runtime.Caller(2)

	snapshot := &files.Snapshot{
		Name:     title,
		FilePath: filePath,
		Content:  content,
	}

	accepted, err := files.ReadAccepted(title)
	if err == nil {
		if accepted.Content == content {
			return
		}

		if err := files.SaveSnapshot(snapshot, "new"); err != nil {
			t.Error("failed to save snapshot:", err)
			return
		}

		diffLines := convertDiffLines(diff.Histogram(accepted.Content, snapshot.Content))
		fmt.Println(pretty.DiffSnapshotBox(accepted, snapshot, diffLines))
		t.Error("snapshot mismatch - run 'freeze review' to update")
		return
	}

	if err := files.SaveSnapshot(snapshot, "new"); err != nil {
		t.Error("failed to save snapshot:", err)
		return
	}

	fmt.Println(pretty.NewSnapshotBox(snapshot))
	t.Error("new snapshot created - run 'freeze review' to accept")
}

func convertDiffLines(diffLines []diff.DiffLine) []pretty.DiffLine {
	result := make([]pretty.DiffLine, len(diffLines))
	for i, dl := range diffLines {
		result[i] = pretty.DiffLine{
			Number: dl.Number,
			Line:   dl.Line,
			Kind:   pretty.DiffKind(dl.Kind),
		}
	}
	return result
}

func formatValues(values ...any) string {
	var result string
	for _, v := range values {
		result += formatValue(v)
	}
	return result
}

func formatValue(v any) string {
	return utter.Sdump(v)
}

// DOCS:
func Review() error {
	return review.Review()
}

func AcceptAll() error {
	return review.AcceptAll()
}

func RejectAll() error {
	return review.RejectAll()
}

type testingT interface {
	Helper()
	Skip(...any)
	Skipf(string, ...any)
	SkipNow()
	Name() string
	Error(...any)
	Log(...any)
	Cleanup(func())
}
