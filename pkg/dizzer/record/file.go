package record

import (
	"context"
	"crypto/md5"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Sum386 returns the SHA-386 checksum value of the open file.
func Sum386(f *os.File) (string, error) {
	strong := sha512.New384()
	if _, err := io.Copy(strong, f); err != nil {
		return "", fmt.Errorf("%s: %w", f.Name(), err)
	}
	return fmt.Sprintf("%x", strong.Sum(nil)), nil
}

// SumMD5 returns the MD5 checksum value of the open file.
func SumMD5(f *os.File) (string, error) {
	weak := md5.New() //nolint: gosec
	if _, err := io.Copy(weak, f); err != nil {
		return "", fmt.Errorf("%s: %w", f.Name(), err)
	}
	return fmt.Sprintf("%x", weak.Sum(nil)), nil
}

func Determine(name string) (string, error) {
	const file = "file" // file — determine file type
	path, err := exec.LookPath(file)
	if err != nil {
		return "", err
	}
	const ten = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), ten)
	defer cancel()
	args := []string{name}
	cmd := exec.CommandContext(ctx, path, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(out))
	ss := strings.Split(s, ":")
	if len(ss) > 1 {
		return ss[1], nil
	}
	return s, nil
}
