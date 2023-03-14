package run

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/config"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/Defacto2/df2/pkg/images"
	"github.com/Defacto2/df2/pkg/people"
	"github.com/Defacto2/df2/pkg/prompt"
	"github.com/Defacto2/df2/pkg/proof"
	"github.com/Defacto2/df2/pkg/sitemap"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/Defacto2/df2/pkg/text"
	"github.com/Defacto2/df2/pkg/zipcontent"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"go.uber.org/zap"
)

const (
	datal    = "datalist"
	dl       = "dl"
	htm      = "html"
	txt      = "text"
	statusOk = 200
)

var (
	ErrToFew  = errors.New("too few arguments given")
	ErrArg    = errors.New("unknown args flag")
	ErrDZFlag = errors.New("unknown demozoo flag")
)

func Data(w io.Writer, d database.Flags) error {
	switch {
	case d.CronJob:
		return d.Run(w)
	case d.Tables == "all":
		return d.DB(w)
	default:
		return d.ExportTable(w)
	}
}

func APIs(w io.Writer, l *zap.SugaredLogger, a arg.Apis) error {
	switch {
	case a.Refresh:
		return demozoo.RefreshMeta(w, l)
	case a.Pouet:
		return demozoo.RefreshPouet(w, l)
	case a.SyncDos:
		return syncdos(w)
	case a.SyncWin:
		return syncwin(w)
	default:
		return ErrArg
	}
}

func Demozoo(w io.Writer, log *zap.SugaredLogger, dzf arg.Demozoo) error {
	var empty []string
	r := demozoo.Request{
		All:       dzf.All,
		Overwrite: dzf.Overwrite,
	}
	switch {
	case dzf.New, dzf.All:
		return r.Queries(w, log)
	case dzf.ID != "":
		return r.Query(w, log, dzf.ID)
	case dzf.Releaser != 0:
		return releaser(w, dzf.Releaser)
	case dzf.Ping != 0:
		return ping(w, dzf.Ping)
	case dzf.Download != 0:
		return download(w, dzf.Download)
	case len(dzf.Extract) == 1:
		return extract(w, dzf.Extract[0])
	case len(dzf.Extract) > 1: // limit to the first 2 flags
		d, err := archive.Demozoo(w, dzf.Extract[0], dzf.Extract[1], &empty)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, d)
		return nil
	default:
		return ErrDZFlag
	}
}

func syncdos(w io.Writer) error {
	var p demozoo.MsDosProducts
	if err := p.Get(w); err != nil {
		return err
	}
	fmt.Fprintf(w, "There were %d new productions found\n", p.Finds)
	return nil
}

func syncwin(w io.Writer) error {
	var p demozoo.WindowsProducts
	if err := p.Get(w); err != nil {
		return err
	}
	fmt.Fprintf(w, "There were %d new productions found\n", p.Finds)
	return nil
}

func releaser(w io.Writer, id uint) error {
	var r demozoo.Releaser
	if err := r.Get(id); err != nil {
		return err
	}
	fmt.Fprintf(w, "Demozoo ID %v, HTTP status %v\n", id, r.Status)
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
	return demozoo.InsertProds(w, &p.API)
}

func ping(w io.Writer, id uint) error {
	var f demozoo.Product
	err := f.Get(id)
	if err != nil {
		return err
	}
	if !str.Piped() {
		fmt.Fprintf(w, "Demozoo ID %v, HTTP status %v\n", id, f.Status)
	}
	return f.API.Print(w)
}

func download(w io.Writer, id uint) error {
	var f demozoo.Product
	if err := f.Get(id); err != nil {
		return err
	}
	fmt.Fprintf(w, "Demozoo ID %v, HTTP status %v\n", id, f.Status)
	f.API.Downloads(w)
	fmt.Fprintln(w)
	return nil
}

func extract(w io.Writer, src string) error {
	var empty []string
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	d, err := archive.Demozoo(w, src, id.String(), &empty)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, d)
	return nil
}

func Groups(w io.Writer, gpf arg.Group) error {
	switch {
	case gpf.Cronjob, gpf.Forcejob:
		force := false
		if gpf.Forcejob {
			force = true
		}
		if err := groups.Cronjob(w, force); err != nil {
			return err
		}
		return nil
	}
	arg.FilterFlag(w, groups.Wheres(), "filter", gpf.Filter)
	req := groups.Request{Filter: gpf.Filter, Counts: gpf.Counts, Initialisms: gpf.Init, Progress: gpf.Progress}
	switch gpf.Format {
	case datal, dl, "d":
		return req.DataList(w, "")
	case htm, "h", "":
		return req.HTML(w, "")
	case txt, "t":
		if _, err := req.Print(w); err != nil {
			return err
		}
	}
	return nil
}

func New(w io.Writer, l *zap.SugaredLogger) error {
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
	if err := newDZ.Queries(w, l); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. scan for new proof submissions\n", i)
	newProof := proof.Request{
		Overwrite:   false,
		AllProofs:   false,
		HideMissing: false,
	}
	if err := newProof.Queries(w, l); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. scan for empty archives\n", i)
	if err := zipcontent.Fix(w, l, true); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing images\n", i)
	if err := images.Fix(w, l); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. generate missing text previews\n", i)
	if err := text.Fix(w); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. fix demozoo data conflicts\n", i)
	if err := demozoo.Fix(w); err != nil {
		return err
	}
	i++
	color.Info.Printf("%d. fix malformed database entries\n", i)
	if err := database.Fix(w, l); err != nil {
		return err
	}
	return groups.Fix(w)
}

func People(w io.Writer, pf arg.People) error {
	switch {
	case pf.Cronjob, pf.Forcejob:
		force := false
		if pf.Forcejob {
			force = true
		}
		if err := people.Cronjob(w, force); err != nil {
			return err
		}
		return nil
	}
	fmtflags := [7]string{datal, htm, txt, dl, "d", "h", "t"}
	arg.FilterFlag(w, people.Filters(), "filter", pf.Filter)
	var req people.Request
	if arg.FilterFlag(w, fmtflags, "format", pf.Format); pf.Format != "" {
		req = people.Request{Filter: pf.Filter, Progress: pf.Progress}
	}
	switch pf.Format {
	case datal, dl, "d":
		return people.DataList(w, "", req)
	case htm, "h", "":
		return people.HTML(w, "", req)
	case txt, "t":
		return people.Print(w, req)
	}
	return nil
}

func Rename(w io.Writer, args ...string) error {
	const wantedCount = 2
	switch len(args) {
	case 0, 1:
		return ErrToFew
	case wantedCount:
		// do nothing
	default:
		fmt.Fprintln(w, "The renaming of groups only supports two arguments, "+
			"names with spaces should be quoted, for example:")
		fmt.Fprintf(w, "df2 fix rename %s %q\n", args[0], strings.Join(args[1:], " "))
		return nil
	}
	oldArg, newArg := args[0], args[1]
	src, err := groups.Exact(w, oldArg)
	if err != nil {
		return err
	}
	if src < 1 {
		fmt.Fprintf(w, "no group matches found for %q\n", oldArg)
		return nil
	}
	newName := groups.Format(newArg)
	dest, err := groups.Exact(w, newName)
	if err != nil {
		return err
	}
	switch dest {
	case 0:
		fmt.Fprintf(w, "Will rename the %d records of %q to the new group name, %q\n", src, oldArg, newName)
	default:
		fmt.Fprintf(w, "Will merge the %d records of %q into the group %q to total %d records\n",
			src, oldArg, newName, src+dest)
		color.Danger.Println("This cannot be undone")
	}
	if b := prompt.YN("Rename the group", false); !b {
		return nil
	}
	i, err := groups.Update(w, newName, oldArg)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d records updated to use %q\n", i, newName)
	return nil
}

func TestSite(w io.Writer, base string) error { //nolint:funlen
	urls, err := sitemap.FileList(base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d various query-string options of the list of files\n", len(urls))
	sitemap.Success.Range(w, urls)

	const pingCount = 10
	total, ids, err := sitemap.RandIDs(w, pingCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	color.Primary.Printf("\nRequesting the <title> of %d random files from %d public records\n", pingCount, total)
	sitemap.LinkSuccess.Range(w, urls)

	total, ids, err = sitemap.RandIDs(w, pingCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	color.Primary.Printf("\nRequesting the content disposition of %d random file download from %d public records\n",
		pingCount, total)
	sitemap.Success.RangeFiles(w, urls)

	const hideCount = 2
	total, ids, err = sitemap.RandDeleted(w, hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	color.Primary.Printf("\nRequesting the <title> of %d "+
		"random files from %d disabled records\n", hideCount, total)
	sitemap.LinkNotFound.Range(w, urls)

	total, ids, err = sitemap.RandBlocked(w, hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	color.Primary.Printf("\nRequesting the content disposition of %d "+
		"random file download from %d disabled records\n", hideCount, total)
	sitemap.NotFound.RangeFiles(w, urls)

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
	sitemap.NotFound.Range(w, urls)

	paths, err := sitemap.AbsPaths(base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d static URLs used in the sitemap.xml\n", len(paths))
	sitemap.Success.Range(w, paths[:])

	html3s, err := sitemap.AbsPathsH3(w, base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d static URLs used by the HTML3 text mode\n", len(html3s))
	sitemap.Success.Range(w, html3s)

	return nil
}
