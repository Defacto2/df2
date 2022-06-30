package sys

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ErrOSFix    = errors.New("os fix could not read the file archive")
	ErrTypeOut  = errors.New("os magic file program result is empty")
	ErrWrongExt = errors.New("filename has the wrong file extension")
	ErrUnknExt  = errors.New("the archive uses an unsupported file extension")
)

const (
	// 7z,arc,ark,arj,cab,gz,lha,lzh,rar,tar,tar.gz,zip
	arjext = ".arj"
	zipext = ".zip"
)

// MagicExt uses the Linux file program to determine the src archive file type.
// The returned string will be a file separator and extension.
// Note both bzip2 and gzip archives return a .tar extension prefix.
func MagicExt(src string) (string, error) {
	prog, err := exec.LookPath("file")
	if err != nil {
		return "", fmt.Errorf("magic file type: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, "--brief", src)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("magic file type: %w", err)
	}
	if len(out) == 0 {
		return "", fmt.Errorf("magic file type: %w", ErrTypeOut)
	}

	s := strings.Split(strings.ToLower(string(out)), ",")
	magic := strings.TrimSpace(s[0])
	switch magic {
	case "7-zip archive data":
		return ".7z", nil
	case "arj archive data":
		return arjext, nil
	case "bzip2 compressed data":
		return ".tar.bz2", nil
	case "gzip compressed data":
		return ".tar.gz", nil
	case "rar archive data":
		return ".rar", nil
	case "posix tar archive":
		return ".tar", nil
	case "zip archive data":
		return zipext, nil
	default:
		return "", fmt.Errorf("no unsupport for magic file type: %q", magic)
	}
}

// Rename the filename by replacing the file extension with the ext string.
// Leaving ext empty returns the filename without a file extension.
func Rename(ext, filename string) string {
	const sep = "."
	s := strings.Split(filename, sep)
	if ext == "" && len(s) == 1 {
		return filename
	}
	if ext == "" {
		return strings.Join(s[:len(s)-1], sep)
	}
	if len(s) == 1 {
		s = append(s, ".tmp")
	}
	s[len(s)-1] = strings.Join(strings.Split(ext, sep), "")
	return strings.Join(s, sep)
}

// Readr attempts to use programs on the operating system to determine
// the archive filename and content.
func Readr(src, filename string) ([]string, string, error) {
	ext, err := MagicExt(src)
	if err != nil {
		return []string{}, "", fmt.Errorf("system reader: %w", err)
	}
	if ext != filepath.Ext(filename) {
		// retry using correct filename extension
		return []string{}, ext, fmt.Errorf("system reader: %w", ErrWrongExt)
	}
	switch ext {
	case arjext:
		return ArjReader(src)
	case zipext:
		return ZipReader(src)
	}
	return []string{}, "", fmt.Errorf("system reader: %w", ErrOSFix)
}

// Extract extracts the targets from the src file archive
// to the dest directory using an Linux archive program.
// The program used is determined by the extension of the
// provided archive filename, which maybe different to src.
func Extract(filename, src, targets, dest string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case arjext:
		return ArjExtract(src, targets, dest)
	case zipext:
		return ZipExtract(src, targets, dest)
	default:
		return ErrUnknExt
	}
}

func ArjExtract(src, targets, dest string) error {
	prog, err := exec.LookPath("arj")
	if err != nil {
		return fmt.Errorf("arj extract: %w", err)
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// arj x archive destdir/ *
	const extract = "x"
	cmd := exec.CommandContext(ctx, prog, extract, src, dest, targets)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("%s: %s", prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("%s: %w", prog, err)
	}
	return nil
}

// ZipExtract extracts the target filenames from the src zip archive
// to the dest directory using the Linux unzip program.
// Multiple filenames can be separated by spaces.
func ZipExtract(src, targets, dest string) error {
	prog, err := exec.LookPath("unzip")
	if err != nil {
		return fmt.Errorf("unzip extract: %w", err)
	}
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// [-options]
	const (
		test            = "-t"  // test archive files
		caseinsensitive = "-C"  // use case-insensitive matching
		notimestamps    = "-D"  // skip restoration of timestamps
		junkpaths       = "-j"  // junk paths, ignore directory structures
		overwrite       = "-o"  // overwrite existing files without prompting
		quiet           = "-q"  // quiet
		quieter         = "-qq" // quieter
		targetDir       = "-d"  // target directory to extract files to
	)
	// unzip [-options] file[.zip] [file(s)...] [-x files(s)] [-d exdir]
	// file[.zip]		path to the zip archive
	// [file(s)...]		optional list of archived files to process, sep by spaces.
	// [-x files(s)]	optional files to be excluded.
	// [-d exdir]		optional target directory to extract files in.
	cmd := exec.CommandContext(ctx, prog, quieter, junkpaths, overwrite, src, targets, targetDir, dest)
	cmd.Stderr = &b
	if err = cmd.Run(); err != nil {
		if b.String() != "" {
			return fmt.Errorf("%s: %s", prog, strings.TrimSpace(b.String()))
		}
		return fmt.Errorf("%s: %w", prog, err)
	}
	return nil
}

// note: limits to 999 items
func ArjReader(src string) ([]string, string, error) {
	prog, err := exec.LookPath("arj")
	if err != nil {
		return nil, "", fmt.Errorf("arj reader: %w", err)
	}

	const verboselist = "v"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, verboselist, src)
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		return nil, "", err
	}
	outs := strings.Split(string(out), "\n")
	files := []string{}
	const start = len("001) ")
	for _, s := range outs {
		if !ArjItem(s) {
			continue
		}
		files = append(files, s[start:])
	}
	// append empty value to match the other readers
	files = append(files, "")
	return files, arjext, nil
}

func ArjItem(s string) bool {
	if len(s) < 6 {
		return false
	}
	if s[3:4] != ")" {
		return false
	}
	x := s[:3]
	if _, err := strconv.Atoi(x); err != nil {
		return false
	}
	return true
}

func ZipReader(src string) ([]string, string, error) {
	prog, err := exec.LookPath("zipinfo")
	if err != nil {
		return nil, "", fmt.Errorf("zipinfo reader: %w", err)
	}

	const list = "-1"
	var b bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, list, src)
	cmd.Stderr = &b
	out, err := cmd.Output()

	files := strings.Split(string(out), "\n")
	if err != nil {
		// handle broken zips that still contain some valid files
		if b.String() != "" && len(out) > 0 {
			fmt.Print(strings.ReplaceAll(b.String(), "\n", " "))
			return files, zipext, nil
		}
		// otherwise the zipinfo threw an error
		return nil, "", fmt.Errorf("%q: %w", src, err)
	}
	if len(out) == 0 {
		return nil, "", ErrOSFix
	}
	return files, zipext, nil
}
