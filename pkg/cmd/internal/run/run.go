package run

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
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
	"github.com/Defacto2/df2/pkg/prompt"
	"github.com/Defacto2/df2/pkg/proof"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/text"
	"github.com/Defacto2/df2/pkg/zipcontent"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"github.com/hako/durafmt"
)

const (
	datal    = "datalist"
	dl       = "dl"
	htm      = "html"
	txt      = "text"
	statusOk = 200
)

var ErrFewArgs = errors.New("too few arguments given")

// Copyright returns a © Copyright year, or a range of years.
func Copyright() string {
	const initYear = 2020
	y, c := time.Now().Year(), initYear
	if y == c {
		return strconv.Itoa(c) // © 2020
	}
	return fmt.Sprintf("%s-%s", strconv.Itoa(c), time.Now().Format("06")) // © 2020-21
}

var (
	ErrArgFlag = errors.New("unknown args flag")
	ErrDZFlag  = errors.New("unknown demozoo flag")
)

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

func Apis(a arg.Apis) error {
	switch {
	case a.Refresh:
		if err := demozoo.RefreshMeta(); err != nil {
			return err
		}
	case a.Pouet:
		if err := demozoo.RefreshPouet(); err != nil {
			return err
		}
	case a.SyncDos:
		if err := syncdos(); err != nil {
			return err
		}
	case a.SyncWin:
		if err := syncwin(); err != nil {
			return err
		}
	default:
		return ErrArgFlag
	}
	return nil
}

func Demozoo(dzf arg.Demozoo) error {
	var empty []string
	r := demozoo.Request{
		All:       dzf.All,
		Overwrite: dzf.Overwrite,
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
	case dzf.Releaser != 0:
		if err := releaser(dzf.Releaser); err != nil {
			return err
		}
	case dzf.Ping != 0:
		if err := ping(dzf.Ping); err != nil {
			return err
		}
	case dzf.Download != 0:
		if err := download(dzf.Download); err != nil {
			return err
		}
	case len(dzf.Extract) == 1:
		if err := extract(dzf.Extract[0]); err != nil {
			return err
		}
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

func syncdos() error {
	var p demozoo.MsDosProducts
	if err := p.Get(); err != nil {
		return err
	}
	logs.Printf("There were %d new productions found\n", p.Finds)
	return nil
}

func syncwin() error {
	var p demozoo.WindowsProducts
	if err := p.Get(); err != nil {
		return err
	}
	logs.Printf("There were %d new productions found\n", p.Finds)
	return nil
}

func releaser(id uint) error {
	var r demozoo.Releaser
	if err := r.Get(id); err != nil {
		return err
	}
	logs.Printf("Demozoo ID %v, HTTP status %v\n", id, r.Status)
	if r.Code != statusOk {
		return nil
	}
	var p demozoo.ReleaserProducts
	if err := p.Get(id); err != nil {
		return err
	}
	if len(p.API) == 0 {
		return fmt.Errorf("%w: %s", demozoo.ErrNoRel, r.API.Name)
	}
	v := "scener"
	if r.API.IsGroup {
		v = "group"
	}
	s := fmt.Sprintf("Attempt to add the %d productions found for the %s, %s",
		len(p.API), v, r.API.Name)
	if !prompt.YN(s, true) {
		return nil
	}
	if err := demozoo.InsertProds(&p.API); err != nil {
		return err
	}
	return nil
}

func ping(id uint) error {
	var f demozoo.Product
	err := f.Get(id)
	if err != nil {
		return err
	}
	if !str.Piped() {
		logs.Printf("Demozoo ID %v, HTTP status %v\n", id, f.Status)
	}
	if err := f.API.Print(); err != nil {
		return err
	}
	return nil
}

func download(id uint) error {
	var f demozoo.Product
	if err := f.Get(id); err != nil {
		return err
	}
	logs.Printf("Demozoo ID %v, HTTP status %v\n", id, f.Status)
	f.API.Downloads()
	logs.Print("\n")
	return nil
}

func extract(src string) error {
	var empty []string
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	d, err := archive.Demozoo(src, id.String(), &empty)
	if err != nil {
		return err
	}
	logs.Println(d.String())
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
			logs.Printf("%d. %v\n", c, scanner.Text())
			continue
		}
		duration := durafmt.Parse(time.Since(t)).LimitFirstN(1)
		logs.Printf("%v %v ago  %v %s\n",
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
	if err := images.Fix(); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing text previews\n", i)
	if err := text.Fix(); err != nil {
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
	return groups.Fix()
}

func People(pf arg.People) error {
	switch {
	case pf.Cronjob, pf.Forcejob:
		force := false
		if pf.Forcejob {
			force = true
		}
		if err := people.Cronjob(force); err != nil {
			return err
		}
		return nil
	}
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

func Rename(args ...string) error {
	const wantedCount = 2
	switch len(args) {
	case 0, 1:
		return ErrFewArgs
	case wantedCount:
		// do nothing
	default:
		fmt.Println("The renaming of groups only supports two arguments, names with spaces should be quoted, for example:")
		fmt.Printf("df2 fix rename %s %q\n", args[0], strings.Join(args[1:], " "))
		return nil
	}
	oldArg, newArg := args[0], args[1]
	src, err := groups.Exact(oldArg)
	if err != nil {
		return err
	}
	if src < 1 {
		fmt.Printf("no group matches found for %q\n", oldArg)
		return nil
	}
	newName := groups.Format(newArg)
	dest, err := groups.Exact(newName)
	if err != nil {
		logs.Fatal(err)
	}
	switch dest {
	case 0:
		fmt.Printf("Will rename the %d records of %q to the new group name, %q\n", src, oldArg, newName)
	default:
		fmt.Printf("Will merge the %d records of %q into the group %q to total %d records\n", src, oldArg, newName, src+dest)
		color.Danger.Println("This cannot be undone")
	}
	if b := prompt.YN("Rename the group", false); !b {
		return nil
	}
	i, err := groups.Update(newName, oldArg)
	if err != nil {
		return err
	}
	fmt.Printf("%d records updated to use %q\n", i, newName)
	return nil
}

func TestSite(base string) error { //nolint:funlen
	urls, err := sitemap.FileList(base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d various query-string options of the list of files\n", len(urls))
	sitemap.Success.Range(urls)

	const pingCount = 10
	total, ids, err := sitemap.RandIDs(pingCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	color.Primary.Printf("\nRequesting the <title> of %d random files from %d public records\n", pingCount, total)
	sitemap.LinkSuccess.Range(urls)

	total, ids, err = sitemap.RandIDs(pingCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	color.Primary.Printf("\nRequesting the content disposition of %d random file download from %d public records\n",
		pingCount, total)
	sitemap.Success.RangeFiles(urls)

	const hideCount = 2
	total, ids, err = sitemap.RandDeleted(hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	color.Primary.Printf("\nRequesting the <title> of %d "+
		"random files from %d disabled records\n", hideCount, total)
	sitemap.LinkNotFound.Range(urls)

	total, ids, err = sitemap.RandBlocked(hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	color.Primary.Printf("\nRequesting the content disposition of %d "+
		"random file download from %d disabled records\n", hideCount, total)
	sitemap.NotFound.RangeFiles(urls)

	invalidIDs := []int{-99999999, -1, 0, 99999999}
	urls = sitemap.IDs(invalidIDs).JoinPaths(base, sitemap.File)
	invalidElms := []string{"-", "womble-bomble", "<script>", "1+%48*1"}
	for _, elm := range invalidElms {
		r, err := url.JoinPath(base, sitemap.File.String(), elm)
		if err != nil {
			return err
		}
		urls = append(urls, r)
	}
	loc, err := url.JoinPath(base, sitemap.File.String())
	if err != nil {
		return err
	}
	urls = append(urls, loc)
	color.Primary.Printf("\nRequesting the <title> of %d invalid file URLs\n", len(urls))
	sitemap.NotFound.Range(urls)

	paths, err := sitemap.AbsPaths(base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d static URLs used in the sitemap.xml\n", len(paths))
	sitemap.Success.Range(paths[:])

	html3s, err := sitemap.AbsPathsH3(base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d static URLs used by the HTML3 text mode\n", len(html3s))
	sitemap.Success.Range(html3s)

	return nil
}
