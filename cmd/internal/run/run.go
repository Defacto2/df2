// Package run is a collection of work functions used by the cobra.Command
// methods.
package run

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/df2/cmd/internal/arg"
	"github.com/Defacto2/df2/pkg/archive"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/demozoo"
	"github.com/Defacto2/df2/pkg/directories"
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
	ErrArg     = errors.New("argument is unsupported")
	ErrToFew   = errors.New("too few arguments given")
	ErrNothing = errors.New(str.NothingToDo)
	ErrZap     = errors.New("zap logger cannot be nil")
)

// Data is the work function for the data command.
func Data(db *sql.DB, w io.Writer, d database.Flags) error {
	if db == nil {
		return database.ErrDB
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

// API is the work function for the api command.
func APIs(db *sql.DB, w io.Writer, a arg.APIs) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	switch {
	case a.Refresh:
		return demozoo.RefreshMeta(db, w)
	case a.Pouet:
		return demozoo.RefreshPouet(db, w)
	case a.SyncDos:
		return syncdos(db, w)
	case a.SyncWin:
		return syncwin(db, w)
	default:
		return fmt.Errorf("%v %w", a, ErrArg)
	}
}

// Env is the work function for the env command.
func Env(w io.Writer, l *zap.SugaredLogger, cfg conf.Config) error {
	if l == nil {
		return ErrZap
	}
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintln(w, "Quietly creating directories.")
	dirs := []string{cfg.Downloads, cfg.Images, cfg.Thumbs, cfg.IncomingFiles, cfg.IncomingImgs, cfg.SQLDumps}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			l.Errorln(err)
		}
	}
	return nil
}

// Demozoos is the work function for the demozoo command.
func Demozoo(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config, dz arg.Demozoo) error {
	if db == nil {
		return database.ErrDB
	}
	if l == nil {
		return ErrZap
	}
	if w == nil {
		w = io.Discard
	}
	empty := []string{}
	r := demozoo.Request{
		All:       dz.All,
		Overwrite: dz.Overwrite,
		Config:    cfg,
		Logger:    l,
	}
	switch {
	case dz.New, dz.All:
		return r.Queries(db, w)
	case dz.ID != "":
		return r.Query(db, w, dz.ID)
	case dz.Releaser != 0:
		return releaser(db, w, dz.Releaser)
	case dz.Ping != 0:
		return ping(w, dz.Ping)
	case dz.Download != 0:
		return download(w, dz.Download)
	case len(dz.Extract) == 1:
		return extract(db, w, cfg, dz.Extract[0])
	case len(dz.Extract) > 1: // limit to the first 2 flags
		zoo := archive.Demozoo{
			Source:   dz.Extract[0],
			UUID:     dz.Extract[1],
			VarNames: &empty,
			Config:   cfg,
		}
		d, err := zoo.Decompress(db, w)
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
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	p := demozoo.MsDosProducts{}
	if err := p.Get(db, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "There were %d new productions found\n", p.Finds)
	return nil
}

func syncwin(db *sql.DB, w io.Writer) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	p := demozoo.WindowsProducts{}
	if err := p.Get(db, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "There were %d new productions found\n", p.Finds)
	return nil
}

func releaser(db *sql.DB, w io.Writer, id uint) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	r := demozoo.Releaser{}
	if err := r.Get(id); err != nil {
		return err
	}
	fmt.Fprintf(w, "Demozoo ID %v, HTTP status %v\n", id, r.Status)
	if r.Code != statusOk {
		return nil
	}
	p := demozoo.ReleaserProducts{}
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
	b, err := prompt.YN(w, s, true)
	if err != nil {
		return err
	}
	if !b {
		return nil
	}
	return demozoo.InsertProds(db, w, &p.API)
}

func ping(w io.Writer, id uint) error {
	if w == nil {
		w = io.Discard
	}
	f := demozoo.Product{}
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
	if w == nil {
		w = io.Discard
	}
	f := demozoo.Product{}
	if err := f.Get(id); err != nil {
		return err
	}
	fmt.Fprintf(w, "Demozoo ID %v, HTTP status %v\n", id, f.Status)
	f.API.Downloads(w)
	fmt.Fprintln(w)
	return nil
}

func extract(db *sql.DB, w io.Writer, cfg conf.Config, src string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	empty := []string{}
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	dz := archive.Demozoo{
		Source:   src,
		UUID:     id.String(),
		VarNames: &empty,
		Config:   cfg,
	}
	d, err := dz.Decompress(db, w)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, d)
	return nil
}

// Group is the work function for the group command.
func Groups(db *sql.DB, w, dest io.Writer, gro arg.Group) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	req := groups.Request{
		Filter:      gro.Filter,
		Counts:      gro.Counts,
		Initialisms: gro.Init,
		Progress:    gro.Progress,
	}
	switch gro.Format {
	case datal, dl, "d":
		return req.DataList(db, w, dest)
	case htm, "h", "":
		return req.HTML(db, w, dest)
	case txt, "t":
		if _, err := req.Print(db, w); err != nil {
			return err
		}
		return nil
	}
	return ErrNothing
}

// GroupCron is the work function for the group cronjob command.
func GroupCron(db *sql.DB, w io.Writer, cfg conf.Config, gro arg.Group) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	ow := false
	if gro.Forcejob {
		ow = true
	}
	dir := cfg.HTMLExports
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("html export directory does not exist: %w: %s", err, dir)
	}
	for _, tag := range groups.Tags() {
		name := fmt.Sprintf("%s.htm", tag)
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			if err1 := directories.Touch(path); err1 != nil {
				return fmt.Errorf("cronjob: %w: %s", err1, path)
			}
		}
		last, err := database.LastUpdate(db)
		if err != nil {
			return fmt.Errorf("cronjob last update: %w", err)
		}
		update := true
		if !ow {
			update, err = database.FileUpdate(path, last)
		}
		switch {
		case err != nil:
			return fmt.Errorf("cronjob file update: %w", err)
		case !update:
			fmt.Fprintf(w, "%s has nothing to update (%s)\n", tag, path)
			continue
		default:
		}
		f, err := os.Create(path) // file is always truncated
		if err != nil {
			return err
		}
		defer f.Close()
		dest := bufio.NewWriter(f)
		if err := groups.Cronjob(db, w, dest, tag, ow); err != nil {
			return err
		}
		fmt.Fprintln(dest)
		dest.Flush()
	}
	return nil
}

func newDemozoo(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config) error {
	newDZ := demozoo.Request{
		All:       false,
		Overwrite: false,
		Refresh:   false,
		Config:    cfg,
		Logger:    l,
	}
	return newDZ.Queries(db, w)
}

func newProof(db *sql.DB, w io.Writer, cfg conf.Config) error {
	newProof := proof.Request{
		Overwrite:   false,
		All:         false,
		HideMissing: false,
	}
	return newProof.Queries(db, w, cfg)
}

func genZIPList(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config) error {
	return zipcontent.Fix(db, w, l, cfg, true)
}

func genImage(db *sql.DB, w io.Writer, cfg conf.Config) error {
	return images.Fix(db, w, cfg)
}

func genText(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config) error {
	return text.Fix(db, w, l, cfg)
}

func fixDZ(db *sql.DB, w io.Writer) error {
	return demozoo.Fix(db, w)
}

func fixDB(db *sql.DB, w io.Writer) error {
	return database.Fix(db, w)
}

func fixGroup(db *sql.DB, w io.Writer) error {
	return groups.Fix(db, w)
}

// New is the work function for the new command.
func New(db *sql.DB, w io.Writer, l *zap.SugaredLogger, cfg conf.Config) error {
	if db == nil {
		return database.ErrDB
	}
	if l == nil {
		return ErrZap
	}
	if w == nil {
		w = io.Discard
	}
	l.Info("> SCAN for new general submissions and cleanup")
	if err := newDemozoo(db, w, l, cfg); err != nil {
		l.Errorln(err)
	}
	l.Info("> SCAN for new proof submissions and cleanup")
	if err := newProof(db, w, cfg); err != nil {
		l.Errorln(err)
	}
	l.Info("> SCAN file archives")
	if err := genZIPList(db, w, l, cfg); err != nil {
		l.Errorln(err)
	}
	l.Info("> GENERATE any missing screenshots or thumbnails")
	if err := genImage(db, w, cfg); err != nil {
		l.Errorln(err)
	}
	l.Info("> GENERATE any previews for the text files")
	if err := genText(db, w, l, cfg); err != nil {
		l.Errorln(err)
	}
	l.Info("> FIX Demozoo conflicts")
	if err := fixDZ(db, w); err != nil {
		l.Errorln(err)
	}
	l.Info("> FIX malformed database records")
	if err := fixDB(db, w); err != nil {
		l.Errorln(err)
	}
	l.Info("> FIX malformed database groups")
	if err := fixGroup(db, w); err != nil {
		l.Errorln(err)
	}
	return nil
}

// People is the work function for the people command.
func People(db *sql.DB, w io.Writer, directory string, f arg.People) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	switch {
	case f.Cronjob, f.Forcejob:
		force := false
		if f.Forcejob {
			force = true
		}
		return people.Cronjob(db, w, directory, force)
	}
	fmtflags := [7]string{datal, htm, txt, dl, "d", "h", "t"}
	if err := arg.FilterFlag(w, people.Filters(), "filter", f.Filter); err != nil {
		return err
	}
	req := people.Flags{}
	if err := arg.FilterFlag(w, fmtflags, "format", f.Format); err != nil {
		return err
	}
	if f.Format != "" {
		req = people.Flags{Filter: f.Filter, Progress: f.Progress}
	}
	switch f.Format {
	case datal, dl, "d":
		return people.DataList(db, w, "", req)
	case htm, "h", "":
		return people.HTML(db, w, "", req)
	case txt, "t":
		return people.Print(db, w, req)
	}
	return ErrNothing
}

// Rename is the work function for the rename command.
func Rename(db *sql.DB, w io.Writer, args ...string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
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
	b, err := prompt.YN(w, "Rename the group", false)
	if err != nil {
		return err
	}
	if !b {
		return nil
	}
	i, err := groups.Update(db, newName, oldArg)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d records updated to use %q\n", i, newName)
	return nil
}

// TestSite is the work function for the test command.
func TestSite(db *sql.DB, w io.Writer, base string) error { //nolint:funlen
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	urls, err := sitemap.FileList(base)
	if err != nil {
		return err
	}
	s := color.Primary.Sprintf("\nRequesting %d various query-string options of the list of files\n", len(urls))
	fmt.Fprint(w, s)
	sitemap.Success.Range(w, urls)

	const pings = 10
	total, ids, err := sitemap.RandIDs(db, pings)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	s = color.Primary.Sprintf("\nRequesting the <title> of %d random files from %d public records\n", pings, total)
	fmt.Fprint(w, s)
	sitemap.LinkSuccess.Range(w, urls)

	total, ids, err = sitemap.RandIDs(db, pings)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	s = color.Primary.Sprintf("\nRequesting the content disposition of %d random file download from %d public records\n",
		pings, total)
	fmt.Fprint(w, s)
	sitemap.Success.RangeFiles(w, urls)

	const hideCount = 2
	total, ids, err = sitemap.RandDeleted(db, hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.File)
	s = color.Primary.Sprintf("\nRequesting the <title> of %d "+
		"random files from %d disabled records\n", hideCount, total)
	fmt.Fprint(w, s)
	sitemap.LinkNotFound.Range(w, urls)

	total, ids, err = sitemap.RandBlocked(db, hideCount)
	if err != nil {
		return err
	}
	urls = ids.JoinPaths(base, sitemap.Download)
	s = color.Primary.Sprintf("\nRequesting the content disposition of %d "+
		"random file download from %d disabled records\n", hideCount, total)
	fmt.Fprint(w, s)
	sitemap.NotFound.RangeFiles(w, urls)

	badIDs := []int{-99999999, -1, 0, 99999999}
	urls = sitemap.IDs(badIDs).JoinPaths(base, sitemap.File)
	badElms := []string{"-", "womble-bomble", "<script>", "1+%48*1"}
	for _, elm := range badElms {
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
	s = color.Primary.Sprintf("\nRequesting the <title> of %d invalid file URLs\n", len(urls))
	fmt.Fprint(w, s)
	sitemap.NotFound.Range(w, urls)

	paths, err := sitemap.AbsPaths(base)
	if err != nil {
		return err
	}
	s = color.Primary.Sprintf("\nRequesting %d static URLs used in the sitemap.xml\n", len(paths))
	fmt.Fprint(w, s)
	sitemap.Success.Range(w, paths[:])

	html3s, err := sitemap.AbsPathsH3(db, base)
	if err != nil {
		return err
	}
	s = color.Primary.Sprintf("\nRequesting %d static URLs used by the HTML3 text mode\n", len(html3s))
	fmt.Fprint(w, s)
	sitemap.Success.Range(w, html3s)

	return nil
}
