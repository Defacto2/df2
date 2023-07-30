// Package arg has structures used to store the flag values used by the
// cobra.Command methods.
package arg

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gookit/color"
)

var (
	ErrFlagValue = errors.New("value for flag is not valid")
	ErrNoCmd     = errors.New("no command argument was provided")
)

// Targets are the file sections to clean.
func Targets() []string {
	return []string{
		"all",
		"download",
		"emulation",
		"image",
	}
}

// Persistent global flags.
type Persistent struct {
	Panic   bool // Enable panic errors to help debug.
	ASCII   bool // Ascii is placeholder for Cobra to store the PersistentFlag value*
	Quiet   bool // Quiet is placeholder for Cobra to store the PersistentFlag value*
	Version bool // Version is placeholder for Cobra to store the PersistentFlag value*
	// * but the quiet flag is handled by main.go.
}

// APIs synchronization flags.
type APIs struct {
	Refresh bool // Refresh empty fields in the database with data from the API.
	Pouet   bool // Pouet sync local files with pouet ids linked on demozoo.
	SyncDos bool // SyncDos scan demozoo for missing local msdos bbstros and cracktros.
	SyncWin bool // SyncWin scan demozoo for missing local windows bbstros and cracktros.
}

// Approve records flags.
type Approve struct {
	Verbose bool // Verbose display the records that are being approved.
}

// Clean orphan file flags.
type Clean struct {
	Delete   bool   // Delete erase the orphan files.
	Humanise bool   // Humanise display the file sizes in human readable format.
	MakeDirs bool   // MakeDirs generate uuid directories.
	Target   string // Target is the type of file to clean.
}

// Demozoo synchronization flags.
type Demozoo struct {
	All       bool     // All scans all demozoo records.
	Overwrite bool     // Overwrite all existing assets.
	New       bool     // New scans for new demozoo submissions.
	ID        string   // ID auto-generated id or a uuid.
	Extract   []string // Extracts and parses an archived file.
	Ping      uint     // Ping fetches and displays the demozoo api response.
	Download  uint     // Download fetches and saves the demozoo api response.
	Releaser  uint     // Releaser add to the local files all the productions of a demozoo scener.
}

// Env flags.
type Env struct {
	Init bool // Init creates the configuration directories.
}

// Group flags.
type Group struct {
	Counts   bool   // Counts display the file totals for each group.
	Cronjob  bool   // Cronjob run the group command as a cronjob.
	Forcejob bool   // Forcejob run the group command as a cronjob ignoring any conditions.
	Init     bool   // Init displays the acronyms and initalialisms for each group.
	Progress bool   // Progress display the progress bar.
	Filter   string // Filter groups by tags.
	Format   string // Format the output.
}

// Import flags.
type Import struct {
	Insert bool // Insert the found text files metadata into the database.
	Limit  uint // Limit the number of found text files to import.
}

// People flags.
type People struct {
	Cronjob  bool   // Cronjob run the people command as a cronjob.
	Forcejob bool   // Forcejob run the people command as a cronjob ignoring any conditions.
	Progress bool   // Progress display the progress bar.
	Filter   string // Filter people by roles.
	Format   string // Format the output.
}

// Prods flags.
type Proof struct {
	ID          string // ID or uuid of a single proof.
	All         bool   // All scans for all proofs, not just new uploads.
	HideMissing bool   // HideMissing hide proofs that are missing their file download.
	Overwrite   bool   // Overwrite all existing images.
}

// Recent flags.
type Recent struct {
	Compress bool // Compress removes insignificant whitespace characters from the output.
	Limit    uint // Limit the number of recent records to display.
}

// TestSite flags.
type TestSite struct {
	LocalHost bool // LocalHost runs the tests to target a developer, Docker setup.
}

// ZipCmmt flags.
type ZipCmmt struct {
	Stdout  bool // Stdout writes any found zip comment to the stdout.
	Unicode bool // Unicode attempts to convert any CP-437 encoded comments to Unicode.
	OW      bool // OW overwrites any existing save zip comments, otherwise they're skipped.
}

// Invalid returns instructions for invalid command arguments and exits with an error code.
func Invalid(w io.Writer, arg string, args ...string) error {
	if arg == "" {
		return ErrNoCmd
	}
	if w == nil {
		w = io.Discard
	}
	s := ""
	if len(args) == 0 {
		s = fmt.Sprintf("%s %s", color.Warn.Sprint("invalid command"),
			color.Bold.Sprintf("\"%s\"", arg))
	}
	if len(args) > 0 {
		s = fmt.Sprintf("%s %s",
			color.Warn.Sprint("invalid command"),
			color.Bold.Sprintf("\"%s %s\"", arg, args[0]))
	}
	s += fmt.Sprint("\n" + color.Warn.Sprint("please use one of the Available Commands shown above"))
	fmt.Fprintln(w, s)
	return nil
}

// FilterFlag compairs the value of the filter flag against the list of slice values.
func FilterFlag(w io.Writer, t any, flag, val string) error {
	if val == "" {
		return nil
	}
	if w == nil {
		w = io.Discard
	}
	if t, ok := t.([]string); ok {
		sup := false
		for _, value := range t {
			if value == val || (val == value[:1]) {
				sup = true
				break
			}
		}
		if sup {
			return nil
		}
		fmt.Fprintf(w, "%s %s\n%s %s\n", color.Warn.Sprintf("unsupported --%s flag value", flag),
			color.Bold.Sprintf("%q", val),
			color.Warn.Sprint("available flag values"),
			color.Primary.Sprint(strings.Join(t, ",")))
		return ErrFlagValue
	}
	return nil
}

// CleanOpts returns a string of sorted options.
func CleanOpts(a ...string) string {
	const opts = "\noptions: "
	if len(a) == 0 {
		return opts + "MISSING"
	}
	sort.Strings(a)
	return opts + strings.Join(a, ", ")
}
