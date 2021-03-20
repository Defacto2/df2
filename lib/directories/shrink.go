package directories

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

type Month uint

const (
	non Month = iota
	jan
	feb
	mar
	apr
	may
	jun
	jul
	aug
	sep
	oct
	nov
	dec
)

var (
// s = viper.GetString("directory.sql")
// f = viper.GetString("directory.incoming.files")
// p = viper.GetString("directory.incoming.previews")
)

func IncFiles() error {

	return nil
}

func IncPreviews() error {

	return nil
}

func SQL() error {
	s := viper.GetString("directory.sql")
	fmt.Printf("SQL Dir: %s\n", s)
	c, err := ioutil.ReadDir(s)
	if err != nil {
		return fmt.Errorf("sql read directory: %w", err)
	}
	size, files, moves, moveSize := 0, 0, 0, 0
	const layout = "2-1-2006"
	//now, err := time.Parse(layout, time.Now().Format(layout))
	for _, f := range c {
		if f.IsDir() {
			continue
		}
		exts := strings.Split(f.Name(), ".")
		dashes := strings.Split(exts[0], "-")
		if len(dashes) < 2 {
			continue
		}
		m := dashes[len(dashes)-2]
		if month(m) == non {
			continue
		}
		files++
		size += int(f.Size())
		create, err := time.Parse(layout, fmt.Sprintf("1-%d-%s", month(m), dashes[len(dashes)-1]))
		if err != nil {
			return err
		}
		const oneMonth = 730
		const old = time.Hour * oneMonth * 2
		age := time.Since(create)
		if age > old {
			if filepath.Ext(f.Name()) == ".sql" {
				fmt.Printf("%s is to be moved.\n", f.Name())
			}
			moves++
			moveSize += int(f.Size())
		}
	}
	fmt.Printf("Found %d files using %s.\n", files, humanize.Bytes(uint64(size)))
	if moves > 0 {
		fmt.Printf("Will move %d items totaling %s, leaving %s used.\n", moves, humanize.Bytes(uint64(moveSize)), humanize.Bytes(uint64(size-moveSize)))
	}
	//
	//
	// keep two months back, current and last month
	// syntax = d2-sql-update-September-2020.sql.sha1
	// slice filename and match last two items
	return nil
}

func month(s string) Month {
	//fmt.Println(strings.ToLower(s)[:3])
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
