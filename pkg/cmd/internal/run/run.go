package run

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/config"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/Defacto2/df2/pkg/proof"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/text"
	"github.com/Defacto2/df2/pkg/zipcontent"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"github.com/hako/durafmt"
)

const (
	datal = "datalist"
	dl    = "dl"
	htm   = "html"
	txt   = "text"
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

var ErrDZFlag = errors.New("unknown demozoo flag")

func Data(dbf database.Flags) error {
	switch {
	case dbf.CronJob:
		if err := dbf.Run(); err != nil {
			return err
		}
	case dbf.Tables == "all":
		if err := dbf.DB(); err != nil {
			return err
		}
	default:
		if err := dbf.ExportTable(); err != nil {
			return err
		}
	}
	return nil
}

func Demozoo(dzf arg.Demozoo) error { //nolint:funlen
	var empty []string
	r := demozoo.Request{
		All:       dzf.All,
		Overwrite: dzf.Overwrite,
		Refresh:   dzf.Refresh,
		Simulate:  dzf.Simulate,
	}
	switch {
	case dzf.New, dzf.All:
		if err := r.Queries(); err != nil {
			return err
		}
	case dzf.ID != "":
		if err := r.Query(dzf.ID); err != nil {
			return err
		}
	case dzf.Refresh:
		if err := demozoo.RefreshMeta(); err != nil {
			return err
		}
	case dzf.Pouet:
		return demozoo.ErrNA
	case dzf.Sync:
		return demozoo.ErrNA
		// if err := demozoo.Sync(); err != nil {
		// 	return err
		// }
	case dzf.Ping != 0:
		f, err := demozoo.Fetch(dzf.Ping)
		if err != nil {
			return err
		}
		if !str.Piped() {
			logs.Printf("Demozoo ID %v, HTTP status %v\n", dzf.Ping, f.Status)
		}
		if err := f.API.Print(); err != nil {
			return err
		}
	case dzf.Download != 0:
		f, err := demozoo.Fetch(dzf.Download)
		if err != nil {
			return err
		}
		logs.Printf("Demozoo ID %v, HTTP status %v\n", dzf.Download, f.Status)
		f.API.Downloads()
		logs.Print("\n")
	case len(dzf.Extract) == 1:
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		d, err := archive.Demozoo(dzf.Extract[0], id.String(), &empty)
		if err != nil {
			return err
		}
		logs.Println(d.String())
	case len(dzf.Extract) > 1: // limit to the first 2 flags
		d, err := archive.Demozoo(dzf.Extract[0], dzf.Extract[1], &empty)
		if err != nil {
			return err
		}
		logs.Println(d.String())
	default:
		return ErrDZFlag
	}
	return nil
}

func Groups(gpf arg.Group) error {
	switch {
	case gpf.Cronjob, gpf.Forcejob:
		force := false
		if gpf.Forcejob {
			force = true
		}
		if err := groups.Cronjob(force); err != nil {
			return err
		}
		return nil
	}
	arg.FilterFlag(groups.Wheres(), "filter", gpf.Filter)
	req := groups.Request{Filter: gpf.Filter, Counts: gpf.Counts, Initialisms: gpf.Init, Progress: gpf.Progress}
	switch gpf.Format {
	case datal, dl, "d":
		if err := req.DataList(""); err != nil {
			return err
		}
	case htm, "h", "":
		if err := req.HTML(""); err != nil {
			return err
		}
	case txt, "t":
		if _, err := req.Print(); err != nil {
			return err
		}
	}
	return nil
}

func Log() error {
	logs.Printf("%v%v %v\n",
		color.Cyan.Sprint("log file"),
		color.Red.Sprint(":"),
		logs.Filepath(logs.Filename))
	f, err := os.Open(logs.Filepath(logs.Filename))
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	c := 0
	scanner.Text()
	const maxSplit = 5
	for scanner.Scan() {
		c++
		s := strings.SplitN(scanner.Text(), " ", maxSplit)
		t, err := time.Parse("2006/01/02 15:04:05", strings.Join(s[0:2], " "))
		if err != nil {
			fmt.Printf("%d. %v\n", c, scanner.Text())
			continue
		}
		duration := durafmt.Parse(time.Since(t)).LimitFirstN(1)
		fmt.Printf("%v %v ago  %v %s\n",
			color.Secondary.Sprintf("%d.", c),
			duration, color.Info.Sprint(s[2]),
			strings.Join(s[3:], " "))
	}
	return scanner.Err()
}

func New() error {
	i := 0
	color.Primary.Println("Scans for new submissions and record cleanup")
	config.Check()
	i++
	color.Info.Printf("%d. scan for new demozoo submissions\n", i)
	newDZ := demozoo.Request{
		All:       false,
		Overwrite: false,
		Refresh:   false,
		Simulate:  false,
	}
	if err := newDZ.Queries(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. scan for new proof submissions\n", i)
	newProof := proof.Request{
		Overwrite:   false,
		AllProofs:   false,
		HideMissing: false,
	}
	if err := newProof.Queries(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. scan for empty archives\n", i)
	if err := zipcontent.Fix(true); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing images\n", i)
	if err := images.Fix(false); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing text previews\n", i)
	if err := text.Fix(false); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. fix demozoo data conflicts\n", i)
	if err := demozoo.Fix(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. fix malformed database entries\n", i)
	if err := database.Fix(); err != nil {
		return err
	}
	return groups.Fix(false)
}

func People(pf arg.People) error {
	fmtflags := [7]string{datal, htm, txt, dl, "d", "h", "t"}
	arg.FilterFlag(people.Filters(), "filter", pf.Filter)
	var req people.Request
	if arg.FilterFlag(fmtflags, "format", pf.Format); pf.Format != "" {
		req = people.Request{Filter: pf.Filter, Progress: pf.Progress}
	}
	switch pf.Format {
	case datal, dl, "d":
		if err := people.DataList("", req); err != nil {
			return err
		}
	case htm, "h", "":
		if err := people.HTML("", req); err != nil {
			return err
		}
	case txt, "t":
		if err := people.Print(req); err != nil {
			return err
		}
	}
	return nil
}
