package rename

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const space = " "

// Clean a malformed group name and save the fix to the database.
func Clean(name string) bool {
	fix := CleanStr(name)
	if fix == name {
		return false
	}
	count, status := int64(0), str.Y()
	ok := true
	count, err := Update(fix, name)
	if err != nil {
		status = str.X()
		ok = false
	}
	logs.Printf("%s %q %s %s (%d)\n",
		status, name, color.Question.Sprint("⟫"), color.Info.Sprint(fix), count)
	return ok
}

// CleanStr fixes the malformed string.
func CleanStr(s string) string {
	f := database.TrimSP(s)
	f = database.StripChars(f)
	f = database.StripStart(f)
	f = strings.TrimSpace(f)
	f = TrimThe(f)
	f = Format(f)
	return f
}

// Format returns a copy of s with custom formatting.
func Format(s string) string {
	const (
		acronym   = 3
		separator = ", "
	)
	if len(s) <= acronym {
		return strings.ToUpper(s)
	}
	groups := strings.Split(s, separator)
	for j, group := range groups {
		g := strings.ToLower(strings.TrimSpace(group))
		g = FmtSyntax(g)
		if fix := FmtExact(g); fix != "" {
			groups[j] = fix
			continue
		}

		words := strings.Split(g, space)
		last := len(words) - 1
		for i, word := range words {
			word = TrimDot(word)
			if fix := fixHyphens(word); fix != "" {
				words[i] = fix
				continue
			}
			words[i] = fixWord(word, i, last)
		}
		groups[j] = strings.Join(words, space)
	}
	return strings.Join(groups, separator)
}

// FmtSyntax formats the special ampersand (&) character
// to be usable with the URL in use by the group.
func FmtSyntax(w string) string {
	if !strings.Contains(w, "&") {
		return w
	}
	s := w
	trimDupes := regexp.MustCompile(`\&+`)
	s = trimDupes.ReplaceAllString(s, "&")

	trimPrefix := regexp.MustCompile(`^\&+`)
	s = trimPrefix.ReplaceAllString(s, "")

	trimSuffix := regexp.MustCompile(`\&+$`)
	s = trimSuffix.ReplaceAllString(s, "")

	addWhitespace := regexp.MustCompile(`(\S)\&(\S)`) // \S matches any character that's not whitespace
	s = addWhitespace.ReplaceAllString(s, "$1 & $2")
	return s
}

// TrimDot removes any trailing dots from s.
func TrimDot(s string) string {
	const short = 2
	if len(s) < short {
		return s
	}
	if l := s[len(s)-1:]; l == "." {
		return s[:len(s)-1]
	}
	return s
}

// TrimThe removes 'the' prefix from s.
func TrimThe(s string) string {
	const short = 2
	a := strings.Split(s, space)
	if len(a) < short {
		return s
	}
	l := a[len(a)-1]
	if strings.EqualFold(a[0], "the") && (l == "BBS" || l == "FTP") {
		return strings.Join(a[1:], space) // drop "the" prefix
	}
	return s
}

// Update replaces all instances of the group name with a new group name.
func Update(newName, group string) (int64, error) {
	db := database.Connect()
	defer db.Close()
	stmt, err := db.Prepare("UPDATE `files` SET group_brand_for=?," +
		" group_brand_by=? WHERE (group_brand_for=? OR group_brand_by=?)")
	if err != nil {
		return 0, fmt.Errorf("rename statement: %w", err)
	}
	defer stmt.Close()
	res, err := stmt.Exec(newName, newName, group, group)
	if err != nil {
		return 0, fmt.Errorf("rename exec: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rename rows affected: %w", err)
	}
	return count, db.Close()
}

// FmtExact matches the exact group name to apply a format.
func FmtExact(g string) string {
	switch g {
	// all uppercase full groups
	case "anz ftp", "mor ftp", "msv ftp", "nos ftp", "pox ftp", "scf ftp", "scsi ftp",
		"tbb ftp", "tog ftp", "top ftp", "tph-qqt", "tpw ftp", "u4ea ftp", "zoo ftp",
		"3wa bbs", "acb bbs", "bcp bbs", "cwl bbs", "es bbs", "dv8 bbs", "fic bbs",
		"lms bbs", "lta bbs", "ls bbs", "lpc bbs", "og bbs", "okc bbs", "uct bbs", "tsi bbs",
		"tsc bbs", "trt 2001 bbs", "tiw bbs", "tfz 2 bbs", "ppps bbs", "pp bbs", "pmc bbs",
		"crsiso", "tus fx", "lsdiso", "cnx ftp", "tph-qqt ftp", "swat", "psxdox", "nsdap",
		"new dtl", "lkcc", "core", "qed bbs", "psi bbs", "tcsm bbs",
		"2nd2none bbs", "ckc bbs", "beer":
		return strings.ToUpper(g)
	case "scenet":
		// all lowercase full groups
		return strings.ToLower(g)
	}
	return FmtByName(g)
}

// FmtByName formats the group name using stylized casing.
func FmtByName(s string) string {
	name := strings.ToLower(s)
	fmtGroups := map[string]string{
		"drm ftp":               "dRM FTP",
		"dst ftp":               "dst FTP",
		"nofx bbs":              "NoFX BBS",
		"noclass":               "NoClass",
		"pjs tower":             "PJs Tower BBS",
		"tsg ftp":               "tSG FTP",
		"xquizit ftp":           "XquiziT FTP",
		"vdr lake ftp":          "VDR Lake FTP",
		"ptl club":              "PTL Club",
		"dvtiso":                "DVTiSO",
		"rhvid":                 "RHViD",
		"trsi":                  "TRSi",
		"htbzine":               "HTBZine",
		"mci escapes":           "mci escapes",
		"79th trac":             "79th TRAC",
		"unreal magazine":       "UnReal Magazine",
		"ice weekly newsletter": "iCE Weekly Newsletter",
		"biased":                "bIASED",
		"dreadloc":              "DREADLoC",
		"cybermail":             "CyberMail",
		"excretion anarchy":     "eXCReTION Anarchy",
		"pocketheaven":          "PocketHeaven",
		"rzsoft ftp":            "RZSoft FTP",
	}
	for group, replace := range fmtGroups {
		if name == group {
			return replace
		}
	}
	renGroups := map[string]string{
		"2000 ad":    "2000AD",
		"hashx":      "Hash X",
		"phoenixbbs": "Phoenix BBS",
	}
	for group, replace := range renGroups {
		if name == group {
			return replace
		}
	}
	return ""
}

func fmtWord(w string) string {
	switch w {
	case "3d", "abc", "acdc", "ad", "amf", "ansi", "asm", "au", "bbc", "bbs", "bc", "cd",
		"cgi", "diz", "dox", "eu", "faq", "fbi", "ftp", "fr", "fx", "fxp", "hq", "id", "ii",
		"iii", "iso", "kgb", "pc", "pcb", "pcp", "pda", "psx", "pwa", "ssd", "st", "tnt",
		"tsr", "ufo", "uk", "us", "usa", "uss", "ussr", "vcd", "whq", "mp3", "rom", "fm",
		"am", "pm", "gbc", "gif", "xxx", "rpm":
		return strings.ToUpper(w)
	case "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th",
		"10th", "11th", "12th", "13th":
		return strings.ToLower(w)
	case "7of9":
		return strings.ToLower(w)
	default:
		return ""
	}
}

func fmtSuffix(w string, title cases.Caser) string {
	uppers := []string{"ad", "bc", "am", "pm"}
	for _, x := range uppers {
		if val := upperSuffix(w, x); val != "" {
			return val
		}
	}
	switch {
	case strings.HasSuffix(w, "dox"):
		val := strings.TrimSuffix(w, "dox")
		return fmt.Sprintf("%sDox", title.String(val))
	case strings.HasSuffix(w, "fxp"):
		val := strings.TrimSuffix(w, "fxp")
		return fmt.Sprintf("%sFXP", title.String(val))
	case strings.HasSuffix(w, "iso"):
		val := strings.TrimSuffix(w, "iso")
		return fmt.Sprintf("%sISO", title.String(val))
	case strings.HasSuffix(w, "nfo"):
		val := strings.TrimSuffix(w, "nfo")
		return fmt.Sprintf("%sNFO", title.String(val))
	case strings.HasPrefix(w, "pc-"):
		val := strings.TrimPrefix(w, "pc-")
		return fmt.Sprintf("PC-%s", title.String(val))
	case strings.HasPrefix(w, "lsd"):
		val := strings.TrimPrefix(w, "lsd")
		return fmt.Sprintf("LSD%s", title.String(val))
	}
	return ""
}

func upperSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		return ""
	}
	x := strings.TrimSuffix(s, suffix)
	if val, err := strconv.Atoi(x); err == nil {
		return fmt.Sprintf("%d%s", val, strings.ToUpper(suffix))
	}
	return ""
}

func fmtSequence(w string, i int) string {
	if i != 0 {
		return ""
	}
	switch w { //nolint:gocritic
	case "inc":
		// note: Format() applies UPPER to all 3 letter or smaller words
		return strings.ToUpper(w)
	}
	return ""
}

func fmtConnect(w string, position, last int) string {
	const first = 0
	if position == first || position == last {
		return ""
	}
	switch w {
	case "a", "as", "and", "at", "by", "el", "of", "for", "from", "in", "is", "or", "tha",
		"the", "to", "with":
		return strings.ToLower(w)
	}
	return ""
}

func fixWord(w string, position, last int) string {
	if fix := fmtConnect(w, position, last); fix != "" {
		return fix
	}
	if fix := fmtWord(w); fix != "" {
		return fix
	}
	title := cases.Title(language.English, cases.NoLower)
	if fix := fmtSuffix(w, title); fix != "" {
		return fix
	}
	if fix := fmtSequence(w, position); fix != "" {
		return fix
	}
	return title.String(w)
}

func fixHyphens(w string) string {
	const hyphen = "-"
	if !strings.Contains(w, hyphen) {
		return ""
	}
	compounds := strings.Split(w, hyphen)
	last := len(compounds) - 1
	for i, word := range compounds {
		compounds[i] = fixWord(word, i, last)
	}
	return strings.Join(compounds, hyphen)
}
