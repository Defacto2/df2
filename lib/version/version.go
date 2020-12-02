package version

// Code generated by go generate; ignore.
// See /lib/root.go for the version flag.

// Build and version information
type Build struct {
	// Commit Git SHA
	Commit string
	// Date in RFC3339
	Date string
	// Version of Defacto2
	Version string
}

// B holds the build and version information.
var B = Build{
	Commit:  "n/a",
	Date:    "2020-12-03T08:57:17+11:00",
	Version: "v1.1.12",
}
