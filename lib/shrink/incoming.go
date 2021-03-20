package shrink

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/df2/lib/archive"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

// cd /opt/daily-defacto2/ROOT/incoming/user_submissions/files

func Files() {
	s := viper.GetString("directory.incoming.files")
	fmt.Printf("Incoming files directory: %s\n", s)
	if err := store(s, "Incoming", "incoming-files"); err != nil {
		logs.Danger(err)
		return
	}
	fmt.Println("Incoming storage is complete.")
}

func Previews() {
	s := viper.GetString("directory.incoming.previews")
	fmt.Printf("Previews incoming directory: %s\n", s)
	if err := store(s, "Previews", "incoming-preview"); err != nil {
		logs.Danger(err)
		return
	}
	fmt.Println("Previews storage is complete.")
}

func store(path string, cmd string, partial string) error {
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
		//fmt.Printf("%s has no files for backup.\n", cmd)
		return nil
	}
	fmt.Printf("%s found %d files using %s for backup.\n", cmd, cnt, humanize.Bytes(uint64(inUse)))

	n := time.Now()
	filename := filepath.Join(saveDir(), fmt.Sprintf("d2-%s_%d-%02d-%02d.tar", partial, n.Year(), n.Month(), n.Day()))
	store, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("store create: %w", err)
	}
	defer store.Close()

	if errs := archive.Store(files, store); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%d %s errors", len(errs), partial)
	}

	if errs := archive.Delete(files); errs != nil {
		for i, err := range errs {
			fmt.Printf("error #%d: %s\n", i+1, err)
		}
		return fmt.Errorf("%d %s delete errors", len(errs), strings.ToLower(partial))
	}
	fmt.Printf("%s freeing up space is complete.\n", cmd)

	return nil
}
