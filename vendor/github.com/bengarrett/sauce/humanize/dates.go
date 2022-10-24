package humanize

import (
	"time"
)

type Layout string

const (
	DMY   Layout = "2 Jan 2006"
	YMD   Layout = "2006 Jan 2"
	MDY   Layout = "Jan 2 2006"
	H12   Layout = "3:04 pm"
	H24   Layout = "15:04"
	DMY12 Layout = DMY + " " + H12 // 2 Jan 2006 3:04 pm
	DMY24 Layout = DMY + " " + H24 // 2 Jan 2006 15:04
	YMD12 Layout = YMD + " " + H12 // 2006 Jan 2 3:04 pm
	YMD24 Layout = YMD + " " + H24 // 2006 Jan 2 15:04
	MDY12 Layout = MDY + " " + H12 // Jan 2 2006 3:04 pm
	MDY24 Layout = MDY + " " + H24 // Jan 2 2006 15:04
)

// Format returns a formatted time string using a predefined layout.
func (l Layout) Format(t time.Time) string {
	switch l {
	case "":
		return t.Format(string(YMD24))
	case DMY, YMD, MDY, H12, H24, DMY12, DMY24, YMD12, YMD24, MDY12, MDY24:
		return t.Format(string(l))
	default:
		return ""
	}
}
