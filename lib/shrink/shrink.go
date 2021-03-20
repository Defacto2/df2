package shrink

import (
	"log"
	"os"
	"os/user"
	"strings"
)

func month(s string) Month {
	switch strings.ToLower(s)[:3] {
	case "jan":
		return jan
	case "feb":
		return feb
	case "mar":
		return mar
	case "apr":
		return apr
	case "may":
		return may
	case "jun":
		return jun
	case "jul":
		return jul
	case "aug":
		return aug
	case "sep":
		return sep
	case "oct":
		return oct
	case "nov":
		return nov
	case "dec":
		return dec
	default:
		return non
	}
}

func saveDir() string {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln("shrink saveDir failed to get the user home or the working directory:", err)
	}
	return dir
}
