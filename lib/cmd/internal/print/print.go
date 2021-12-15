package print

import (
	"fmt"
	"strconv"
	"time"
)

// Copyright returns a © Copyright year, or a range of years.
func Copyright() string {
	const initYear = 2020
	y, c := time.Now().Year(), initYear
	if y == c {
		return strconv.Itoa(c) // © 2020
	}
	return fmt.Sprintf("%s-%s", strconv.Itoa(c), time.Now().Format("06")) // © 2020-21
}
