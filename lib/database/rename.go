package database

import "regexp"

// StripChars removes incompatible characters used for groups and author names.
func StripChars(s string) string {
	r := regexp.MustCompile(`[^A-Za-zÀ-ÖØ-öø-ÿ0-9\-,& ]`)
	return r.ReplaceAllString(s, "")
}

// StripStart removes non-alphanumeric characters from the start of the string.
func StripStart(s string) string {
	r := regexp.MustCompile(`[A-Za-z0-9À-ÖØ-öø-ÿ]`)
	f := r.FindStringIndex(s)
	if f == nil {
		return ""
	}
	if f[0] != 0 {
		return s[f[0]:]
	}
	return s
}

// TrimSP removes duplicate spaces from a string.
func TrimSP(s string) string {
	r := regexp.MustCompile(`\s+`)
	return r.ReplaceAllString(s, " ")
}
