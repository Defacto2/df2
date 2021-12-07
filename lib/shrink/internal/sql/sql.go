package sql

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/database"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/spf13/viper"
)

var (
	ErrApprove  = errors.New("cannot shrink files as there are database records waiting for public approval")
	ErrDel      = errors.New("delete errors")
	ErrArcStore = errors.New("archive store error")
	ErrSQLComp  = errors.New("sql compress errors")
	ErrSQLDel   = errors.New("sql delete errors")
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

func month(s string) Month {
	const monthPrefix = 3
	switch strings.ToLower(s)[:monthPrefix] {
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

func Approve(cmd string) error {
	wait, err := database.Waiting()
	if err != nil {
		return fmt.Errorf("approve count: %w", err)
	}
	if wait > 0 {
		return fmt.Errorf("%d %s files waiting approval: %w", wait, cmd, ErrApprove)
	}
	return nil
}

func Init() error {
	const (
		layout   = "2-1-2006"
		minDash  = 2
		oneMonth = 730
	)

	s := viper.GetString("directory.sql")
	color.Primary.Printf("SQL directory: %s\n", s)
	c, err := ioutil.ReadDir(s)
	if err != nil {
		return fmt.Errorf("sql read directory: %w", err)
	}

	cnt, freed, inUse := 0, 0, 0
	files := []string{}
	var create time.Time

	for _, f := range c {
		if f.IsDir() {
			continue
		}
		exts := strings.Split(f.Name(), ".")
		dashes := strings.Split(exts[0], "-")
		if len(dashes) < minDash {
			continue
		}
		m := dashes[len(dashes)-minDash]
		if month(m) == non {
			continue
		}
		cnt++
		inUse += int(f.Size())
		create, err = time.Parse(layout,
			fmt.Sprintf("1-%d-%s", month(m), dashes[len(dashes)-1]))
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

	fmt.Printf("SQL found %d files using %s", cnt, humanize.Bytes(uint64(inUse)))
	if len(files) == 0 {
		fmt.Println(", but there is nothing to do.")
		return nil
	}
	fmt.Println(".")
	fmt.Printf("SQL will move %d items totaling %s, leaving %s used.\n",
		len(files), humanize.Bytes(uint64(freed)), humanize.Bytes(uint64(inUse-freed)))

	return sqlProcess(files)
}

func SaveDir() string {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	var dir string
	dir, err = os.Getwd()
	if err != nil {
		log.Fatalln("shrink saveDir failed to get the user home or the working directory:", err)
	}
	return dir
}

func Store(path, cmd, partial string) error {
	c, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("store read: %w", err)
	}

	cnt, inUse := 0, 0
	files := []string{}

	for _, f := range c {
		if f.IsDir() {
			continue
		}
		files = append(files, filepath.Join(path, f.Name()))
		cnt++
		inUse += int(f.Size())
	}

	if len(files) == 0 {
		return nil
	}
	fmt.Printf("%s found %d files using %s for backup.\n", cmd, cnt, humanize.Bytes(uint64(inUse)))

	n := time.Now()
	filename := filepath.Join(SaveDir(), fmt.Sprintf("d2-%s_%d-%02d-%02d.tar", partial, n.Year(), n.Month(), n.Day()))
	store, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("store create: %w", err)
	}
	defer store.Close()

	if errs := archive.Store(files, store); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%w: %d %s", ErrArcStore, len(errs), partial)
	}

	if errs := archive.Delete(files); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%w: %d %s", ErrDel, len(errs), strings.ToLower(partial))
	}
	fmt.Printf("%s freeing up space is complete.\n", cmd)

	return nil
}

func compress(name string, files []string) error {
	tgz, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("sql create: %w", err)
	}
	defer tgz.Close()
	if errs := archive.Compress(files, tgz); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%w: %d", ErrSQLComp, len(errs))
	}
	fmt.Println("SQL archiving is complete.")
	return nil
}

func sqlProcess(files []string) error {
	n := time.Now()
	name := filepath.Join(SaveDir(), fmt.Sprintf("d2-sql_%d-%02d-%02d.tar.gz",
		n.Year(), n.Month(), n.Day()))
	if err := compress(name, files); err != nil {
		return err
	}
	return remove(files)
}

func remove(files []string) error {
	if errs := archive.Delete(files); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%w: %d", ErrSQLDel, len(errs))
	}
	fmt.Println("SQL freeing up space is complete.")
	return nil
}
