package shrink

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
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

func SQL() error {
	const layout = "2-1-2006"
	const oneMonth = 730

	s := viper.GetString("directory.sql")
	fmt.Printf("SQL Dir: %s\n", s)
	c, err := ioutil.ReadDir(s)
	if err != nil {
		return fmt.Errorf("sql read directory: %w", err)
	}

	cnt, freed, inUse := 0, 0, 0
	files := []string{}

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
		cnt++
		inUse += int(f.Size())
		create, err := time.Parse(layout, fmt.Sprintf("1-%d-%s",
			month(m), dashes[len(dashes)-1]))
		if err != nil {
			fmt.Printf("error parsing date from %s: %s\n", f.Name(), err)
			continue
		}
		const expire = time.Hour * oneMonth * 2
		if time.Since(create) > expire {
			if filepath.Ext(f.Name()) == ".sql" {
				fmt.Printf("%s is to be moved.\n", f.Name())
			}
			files = append(files, filepath.Join(s, f.Name()))
			freed += int(f.Size())
		}
	}
	fmt.Printf("Found %d files using %s", cnt, humanize.Bytes(uint64(inUse)))
	if len(files) == 0 {
		fmt.Println(", but there is nothing to do.")
		return nil
	}
	fmt.Println(".")

	fmt.Printf("Will move %d items totaling %s, leaving %s used.\n",
		len(files), humanize.Bytes(uint64(freed)), humanize.Bytes(uint64(inUse-freed)))

	n := time.Now()
	name := filepath.Join(saveDir(), fmt.Sprintf("d2-sql_%d-%02d-%02d.tar.gz", n.Year(), n.Month(), n.Day()))
	tgz, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("sql create: %w", err)
	}
	defer tgz.Close()

	if errs := archive.Compress(files, tgz); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%d sql compress errors", len(errs))
	}
	fmt.Println("Archiving is complete.")

	if errs := archive.Delete(files); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%d sql delete errors", len(errs))
	}
	fmt.Println("Freeing up space is complete.")

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
