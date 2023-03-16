package run

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/configger"
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
	ErrToFew   = errors.New("too few arguments given")
	ErrArg     = errors.New("unknown args flag")
	ErrNothing = errors.New("had nothing to do")

	ErrDB  = errors.New("database handle pointer cannot be nil")
	ErrZap = errors.New("zap logger cannot be nil")
)

func Data(db *sql.DB, w io.Writer, d database.Flags) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	switch {
	case d.CronJob:
		return d.Run(db, w)
	case d.Tables == "all":
		return d.DB(db, w)
	default:
		return d.ExportTable(db, w)
	}
}

func APIs(db *sql.DB, w io.Writer, l *zap.SugaredLogger, a arg.APIs) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if l == nil {
		return ErrZap
	}
	switch {
	case a.Refresh:
		return demozoo.RefreshMeta(db, w, l)
	case a.Pouet:
		return demozoo.RefreshPouet(db, w, l)
	case a.SyncDos:
		return syncdos(db, w)
	case a.SyncWin:
		return syncwin(db, w)
	}
	return nil
}

func Demozoo(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg configger.Config, dz arg.Demozoo) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if l == nil {
		return ErrZap
	}
	var empty []string
	r := demozoo.Request{
		All:       dz.All,
		Overwrite: dz.Overwrite,
	}
	switch {
	case dz.New, dz.All:
		return r.Queries(db, w, l, cfg)
	case dz.ID != "":
		return r.Query(db, w, l, cfg, dz.ID)
	case dz.Releaser != 0:
		return releaser(db, w, dz.Releaser)
	case dz.Ping != 0:
		return ping(w, dz.Ping)
	case dz.Download != 0:
		return download(w, dz.Download)
	case len(dz.Extract) == 1:
		return extract(db, w, cfg, dz.Extract[0])
	case len(dz.Extract) > 1: // limit to the first 2 flags
		d, err := archive.Demozoo(db, w, cfg, dz.Extract[0], dz.Extract[1], &empty)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, d)
		return nil
	default:
		return fmt.Errorf("demozoo %w", ErrNothing)
	}
}

func syncdos(db *sql.DB, w io.Writer) error {
	var p demozoo.MsDosProducts
	if err := p.Get(db, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "There were %d new productions found\n", p.Finds)
	return nil
}

func syncwin(db *sql.DB, w io.Writer) error {
	var p demozoo.WindowsProducts
	if err := p.Get(db, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "There were %d new productions found\n", p.Finds)
	return nil
}

func releaser(db *sql.DB, w io.Writer, id uint) error {
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
	return demozoo.InsertProds(db, w, &p.API)
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

func extract(db *sql.DB, w io.Writer, cfg configger.Config, src string) error {
	var empty []string
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	d, err := archive.Demozoo(db, w, cfg, src, id.String(), &empty)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, d)
	return nil
}

func Groups(db *sql.DB, w io.Writer, directory string, gpf arg.Group) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	switch {
	case gpf.Cronjob, gpf.Forcejob:
		force := false
		if gpf.Forcejob {
			force = true
		}
		if err := groups.Cronjob(db, w, directory, force); err != nil {
			return err
		}
		return nil
	}
	arg.FilterFlag(w, groups.Wheres(), "filter", gpf.Filter)
	req := groups.Request{Filter: gpf.Filter, Counts: gpf.Counts, Initialisms: gpf.Init, Progress: gpf.Progress}
	switch gpf.Format {
	case datal, dl, "d":
		return req.DataList(db, w, "", directory)
	case htm, "h", "":
		return req.HTML(db, w, "", directory)
	case txt, "t":
		if _, err := req.Print(db, w); err != nil {
			return err
		}
		return nil
	}
	return ErrNothing
}

func New(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg configger.Config) error {
	if db == nil {
		return ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if l == nil {
		return ErrZap
	}
	i := 0
	s := color.Primary.Sprint("Scans for new submissions and record cleanup")
	fmt.Fprintln(w, s)
	i++
	s = color.Info.Sprintf("%d. scan for new demozoo submissions\n", i)
	fmt.Fprintln(w, s)
	newDZ := demozoo.Request{
		All:       false,
		Overwrite: false,
		Refresh:   false,
	}
	if err := newDZ.Queries(db, w, l, cfg); err != nil {
		return err
	}
	i++
	s = color.Info.Sprintf("%d. scan for new proof submissions\n", i)
	fmt.Fprintln(w, s)
	newProof := proof.Request{
		Overwrite:   false,
		AllProofs:   false,
		HideMissing: false,
	}
	if err := newProof.Queries(db, w, l, cfg); err != nil {
		return err
	}
	i++
	s = color.Info.Sprintf("%d. scan for empty archives\n", i)
	fmt.Fprintln(w, s)
	if err := zipcontent.Fix(db, w, l, cfg, true); err != nil {
		return err
	}
	i++
	s = color.Info.Sprintf("%d. generate missing images\n", i)
	fmt.Fprintln(w, s)
	if err := images.Fix(db, w, l); err != nil {
		return err
	}
	i++
	s = color.Info.Sprintf("%d. generate missing text previews\n", i)
	fmt.Fprintln(w, s)
	if err := text.Fix(db, w, cfg); err != nil {
		return err
	}
	i++
	s = color.Info.Sprintf("%d. fix demozoo data conflicts\n", i)
	fmt.Fprintln(w, s)
	if err := demozoo.Fix(db, w); err != nil {
		return err
	}
	i++
	s = color.Info.Sprintf("%d. fix malformed database entries\n", i)
	fmt.Fprintln(w, s)
	if err := database.Fix(db, w, l); err != nil {
		return err
	}
	return groups.Fix(db, w)
}

func People(db *sql.DB, w io.Writer, directory string, pf arg.People) error {
	switch {
	case pf.Cronjob, pf.Forcejob:
		force := false
		if pf.Forcejob {
			force = true
		}
		if err := people.Cronjob(db, w, directory, force); err != nil {
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
		return people.DataList(db, w, "", directory, req)
	case htm, "h", "":
		return people.HTML(db, w, "", directory, req)
	case txt, "t":
		return people.Print(db, w, req)
	}
	return nil
}

func Rename(db *sql.DB, w io.Writer, args ...string) error {
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
	src, err := groups.Exact(db, oldArg)
	if err != nil {
		return err
	}
	if src < 1 {
		fmt.Fprintf(w, "no group matches found for %q\n", oldArg)
		return nil
	}
	newName := groups.Format(newArg)
	dest, err := groups.Exact(db, newName)
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
	i, err := groups.Update(db, newName, oldArg)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d records updated to use %q\n", i, newName)
	return nil
}

func TestSite(db *sql.DB, w io.Writer, base string) error { //nolint:funlen
	urls, err := sitemap.FileList(base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d various query-string options of the list of files\n", len(urls))
	sitemap.Success.Range(w, urls)

	const pingCount = 10
	total, ids, err := sitemap.RandIDs(db, pingCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	color.Primary.Printf("\nRequesting the <title> of %d random files from %d public records\n", pingCount, total)
	sitemap.LinkSuccess.Range(w, urls)

	total, ids, err = sitemap.RandIDs(db, pingCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	color.Primary.Printf("\nRequesting the content disposition of %d random file download from %d public records\n",
		pingCount, total)
	sitemap.Success.RangeFiles(w, urls)

	const hideCount = 2
	total, ids, err = sitemap.RandDeleted(db, hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	color.Primary.Printf("\nRequesting the <title> of %d "+
		"random files from %d disabled records\n", hideCount, total)
	sitemap.LinkNotFound.Range(w, urls)

	total, ids, err = sitemap.RandBlocked(db, hideCount)
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

	html3s, err := sitemap.AbsPathsH3(db, w, base)
	if err != nil {
		return err
	}
	color.Primary.Printf("\nRequesting %d static URLs used by the HTML3 text mode\n", len(html3s))
	sitemap.Success.Range(w, html3s)

	return nil
}
