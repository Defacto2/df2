### Importer, adding a new package for a group

1. Create a new package in the `importer` directory. The package name should match the group filename used in the `.nfo` file.
   `\importer\newgroup\newgroup.go`

2. Create a Name const for the group name.
   `const Name = "New Group"`

3. Create a `NfoDate(body)` function that returns the date format used in the `.nfo` file.

```go
func NfoDate(body string) ( //nolint:nonamedreturns
	year int, month time.Month, day int,
) {
	if body == "" {
		return 0, 0, 0
	}

	// DATE : 25.05.2018
	// DATE : 27-01-2015
	rx := regexp.MustCompile(`DATE : (\d\d)[\.|\-](\d\d)[\.|\-](\d\d\d\d)`)
	f := rx.FindStringSubmatch(body)
	const ddmmyy = 4
	if len(f) == ddmmyy {
		y, _ := strconv.Atoi(f[3])
		m, _ := strconv.Atoi(f[2])
		d, _ := strconv.Atoi(f[1])
		return y, time.Month(m), d
	}
	return 0, 0, 0
}
```

4. Implement `Name` result in the `Group(key)` within `importer.go` file.

```go
func Group(key string) string {
	s := PathGroup(key)
	switch strings.ToLower(s) {
	case "newgroup":
		return newgroup.Name
	}
	return s
}
```

5. Implement `NfoDate()` results in the `ReadNfo()` function found within `record\record.go` package.

```go
func (dl *Download) ReadNfo(body, group string) error {
	var (
		m    time.Month
		y, d int
	)
	switch strings.ToLower(group) {
	case "":
		return ErrGroup
	case "newgroup":
		y, m, d = newgroup.NfoDate(body)
	}
	...
}
```

#### If dates are not working, check the rar extracted directories for any `file_id.diz` files. If found, add the following case to the `UseDIZ()` function.

```go
func UseDIZ(g, base string) bool {
	switch g {
	case `newgroup.nfo`:
      return false
   }
	return strings.ToLower(base) == record.FileID
}
```