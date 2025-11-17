package freeze

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ReviewChoice int

const (
	Accept ReviewChoice = iota
	Reject
	Skip
	ToggleDiff
	Quit
)

func Review() error {
	snapshots, err := ListNewSnapshots()
	if err != nil {
		return err
	}

	if len(snapshots) == 0 {
		fmt.Println(FormatSuccess("✓ No new snapshots to review"))
		return nil
	}

	fmt.Println(FormatHeader("🐦 Freeze - Snapshot Review"))
	fmt.Printf("Found %d new snapshot(s) to review\n\n", len(snapshots))

	return reviewLoop(snapshots)
}

func reviewLoop(snapshots []string) error {
	reader := bufio.NewReader(os.Stdin)
	showDiff := false

	for i, testName := range snapshots {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(snapshots), FormatHeader(testName))

		newSnap, err := ReadSnapshot(testName, "new")
		if err != nil {
			fmt.Println(FormatError("✗ Failed to read new snapshot: " + err.Error()))
			continue
		}

		accepted, acceptErr := ReadSnapshot(testName, "accepted")

		if acceptErr == nil && showDiff {
			fmt.Println(DiffSnapshotBox(accepted, newSnap))
		} else if acceptErr == nil {
			fmt.Println(DiffSnapshotBox(accepted, newSnap))
		} else {
			fmt.Println(NewSnapshotBox(newSnap))
		}

		for {
			choice, err := askChoice(reader, i+1, len(snapshots))
			if err != nil {
				return err
			}

			switch choice {
			case Accept:
				if err := AcceptSnapshot(testName); err != nil {
					fmt.Println(FormatError("✗ Failed to accept snapshot: " + err.Error()))
				} else {
					fmt.Println(FormatSuccess("✓ Snapshot accepted"))
				}
				break
			case Reject:
				if err := RejectSnapshot(testName); err != nil {
					fmt.Println(FormatError("✗ Failed to reject snapshot: " + err.Error()))
				} else {
					fmt.Println(FormatWarning("⊘ Snapshot rejected"))
				}
				break
			case Skip:
				fmt.Println(FormatWarning("⊘ Snapshot skipped"))
				break
			case ToggleDiff:
				showDiff = !showDiff
				if acceptErr == nil {
					fmt.Println(DiffSnapshotBox(accepted, newSnap))
				} else {
					fmt.Println(NewSnapshotBox(newSnap))
				}
				continue
			case Quit:
				fmt.Println("\nReview interrupted")
				return nil
			}
			break
		}
	}

	fmt.Println("\n" + FormatSuccess("✓ Review complete"))
	return nil
}

func askChoice(reader *bufio.Reader, current, total int) (ReviewChoice, error) {
	fmt.Printf("\nOptions: [a]ccept [r]eject [s]kip [d]iff [q]uit: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return Quit, err
	}

	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "a", "accept":
		return Accept, nil
	case "r", "reject":
		return Reject, nil
	case "s", "skip":
		return Skip, nil
	case "d", "diff":
		return ToggleDiff, nil
	case "q", "quit":
		return Quit, nil
	default:
		fmt.Println(FormatWarning("Invalid option, please try again"))
		return askChoice(reader, current, total)
	}
}

func AcceptAll() error {
	snapshots, err := ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := AcceptSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(FormatSuccess("✓ Accepted %d snapshot(s)\n"), len(snapshots))
	return nil
}

func RejectAll() error {
	snapshots, err := ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := RejectSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(FormatWarning("⊘ Rejected %d snapshot(s)\n"), len(snapshots))
	return nil
}
