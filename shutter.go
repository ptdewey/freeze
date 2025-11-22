package shutter

import (
	"github.com/kortschak/utter"
	"github.com/ptdewey/shutter/internal/review"
	"github.com/ptdewey/shutter/internal/snapshots"
	"github.com/ptdewey/shutter/internal/transform"
)

const version = "0.1.0"

func init() {
	utter.Config.ElideType = true
	utter.Config.SortKeys = true
}

// Snap takes any values, formats them, and creates a snapshot with the given title.
// For complex types, values are formatted using a pretty-printer.
// The last parameters can be SnapshotOptions to apply scrubbers before snapshotting.
//
//	shutter.Snap(t, "title", any(value1), any(value2), shutter.ScrubUUIDs())
//
// REFACTOR: should this take in _one_ value, and then allow options as additional inputs?
func Snap(t snapshots.T, title string, values ...any) {
	t.Helper()

	// Separate options from values
	var opts []SnapshotOption
	var actualValues []any

	for _, v := range values {
		if opt, ok := v.(SnapshotOption); ok {
			opts = append(opts, opt)
		} else {
			actualValues = append(actualValues, v)
		}
	}

	content := snapshots.FormatValues(actualValues...)

	// Apply scrubber options directly to the formatted content
	scrubbers, _ := extractOptions(opts)
	scrubbedContent := applyOptions(content, scrubbers)

	snapshots.Snap(t, title, version, scrubbedContent)
}

// SnapString takes a string value and creates a snapshot with the given title.
// Options can be provided to apply scrubbers before snapshotting.
func SnapString(t snapshots.T, title string, content string, opts ...SnapshotOption) {
	t.Helper()

	// Apply scrubber options directly to the content
	scrubbers, _ := extractOptions(opts)
	scrubbedContent := applyOptions(content, scrubbers)

	snapshots.Snap(t, title, version, scrubbedContent)
}

// SnapJSON takes a JSON string, validates it, and pretty-prints it with
// consistent formatting before snapshotting. This preserves the raw JSON
// format while ensuring valid JSON structure.
// Options can be provided to apply scrubbers and ignore patterns.
func SnapJSON(t snapshots.T, title string, jsonStr string, opts ...SnapshotOption) {
	t.Helper()

	scrubbers, ignores := extractOptions(opts)

	// Transform the JSON with ignore patterns and scrubbers
	transformConfig := &transform.Config{
		Scrubbers: toTransformScrubbers(scrubbers),
		Ignore:    toTransformIgnorePatterns(ignores),
	}

	transformedJSON, err := transform.TransformJSON(jsonStr, transformConfig)
	if err != nil {
		t.Error("failed to transform JSON:", err)
		return
	}

	snapshots.Snap(t, title, version, transformedJSON)
}

// Review launches an interactive review session to accept or reject snapshot changes.
func Review() error {
	return review.Review()
}

// AcceptAll accepts all pending snapshot changes without review.
func AcceptAll() error {
	return review.AcceptAll()
}

// RejectAll rejects all pending snapshot changes without review.
func RejectAll() error {
	return review.RejectAll()
}

// SnapshotOption represents a transformation that can be applied to snapshot content.
// Options are applied in the order they are provided.
type SnapshotOption interface {
	Apply(content string) string
}

// IgnoreOption represents a pattern for ignoring key-value pairs in JSON structures.
type IgnoreOption interface {
	ShouldIgnore(key, value string) bool
}

// extractOptions separates scrubbers and ignore patterns from options.
func extractOptions(opts []SnapshotOption) (scrubbers []SnapshotOption, ignores []IgnoreOption) {
	for _, opt := range opts {
		if ignore, ok := opt.(IgnoreOption); ok {
			ignores = append(ignores, ignore)
		} else {
			scrubbers = append(scrubbers, opt)
		}
	}
	return scrubbers, ignores
}

// applyOptions applies all scrubber options to content in sequence.
func applyOptions(content string, opts []SnapshotOption) string {
	for _, opt := range opts {
		content = opt.Apply(content)
	}
	return content
}

// scrubberAdapter adapts a SnapshotOption to the transform.Scrubber interface.
type scrubberAdapter struct {
	opt SnapshotOption
}

func (s *scrubberAdapter) Scrub(content string) string {
	return s.opt.Apply(content)
}

func toTransformScrubbers(opts []SnapshotOption) []transform.Scrubber {
	result := make([]transform.Scrubber, len(opts))
	for i, opt := range opts {
		result[i] = &scrubberAdapter{opt: opt}
	}
	return result
}

// ignoreAdapter adapts an IgnoreOption to the transform.IgnorePattern interface.
type ignoreAdapter struct {
	ignore IgnoreOption
}

func (i *ignoreAdapter) ShouldIgnore(key, value string) bool {
	return i.ignore.ShouldIgnore(key, value)
}

func toTransformIgnorePatterns(ignores []IgnoreOption) []transform.IgnorePattern {
	result := make([]transform.IgnorePattern, len(ignores))
	for i, ignore := range ignores {
		result[i] = &ignoreAdapter{ignore: ignore}
	}
	return result
}
