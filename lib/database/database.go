package database

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/uuid/v2/lib/archive"
	"github.com/Defacto2/uuid/v2/lib/directories"
	"github.com/Defacto2/uuid/v2/lib/logs"

	_ "github.com/go-sql-driver/mysql" // MySQL database driver
	"github.com/google/uuid"
)

// UpdateID is a user id to use with the updatedby column.
const UpdateID string = "b66dc282-a029-4e99-85db-2cf2892fffcc"

// Connection information for a MySQL database.
type Connection struct {
	Name   string // database name
	User   string // access username
	Pass   string // access password
	Server string // host server protocol, address and port
}

// Empty is used as a blank value for search maps.
// See: https://dave.cheney.net/2014/03/25/the-empty-struct
type Empty struct{}

// IDs are unique UUID values used by the database and filenames.
type IDs map[string]struct{}

// Record of a file item.
type Record struct {
	ID   string // mysql auto increment id
	UUID string // record unique id
	File string // absolute path to file
	Name string // filename
}

var (
	// TODO move to configuration
	d       = Connection{Name: "defacto2-inno", User: "root", Pass: "password", Server: "tcp(localhost:3306)"}
	proofID string
	pwPath  string // The path to a secured text file containing the d.User login password
)

func recordNew(values []sql.RawBytes) bool {
	if values[2] == nil || string(values[2]) != string(values[3]) {
		return false
	}
	return true
}

// Connect to the database.
func Connect() *sql.DB {
	pw := readPassword()
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@%v/%v", d.User, pw, d.Server, d.Name))
	logs.Check(err)
	// ping the server to make sure the connection works
	err = db.Ping()
	logs.Check(err)
	return db
}

// CreateProof ...
func CreateProof(id string, ow bool, all bool) error {
	if !validUUID(id) && !validID(id) {
		return fmt.Errorf("invalid id given %q it needs to be an auto-generated MySQL id or an uuid", id)
	}
	proofID = id
	return CreateProofs(ow, all)
}

// CreateProofs is a placeholder to scan archives.
func CreateProofs(ow bool, all bool) error {
	db := Connect()
	defer db.Close()
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := "WHERE `section` = 'releaseproof'"
	if proofID != "" {
		switch {
		case validUUID(proofID):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, proofID)
		case validID(proofID):
			w = fmt.Sprintf("%v AND `id`=%q", w, proofID)
		}
	}
	rows, err := db.Query(s + "FROM `files`" + w)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	// more information: https://github.com/go-sql-driver/mysql/wiki/Examples
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	dir := directories.Init(false)
	// fetch the rows
	cnt := 0
	missing := 0
	// todo move to sep func to allow individual record parsing
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		logs.Check(err)
		if new := recordNew(values); !new && !all {
			continue
		}
		cnt++
		r := Record{ID: string(values[0]), UUID: string(values[1]), Name: string(values[4])}
		r.File = filepath.Join(dir.UUID, r.UUID)
		// ping file
		if _, err := os.Stat(r.File); os.IsNotExist(err) {
			fmt.Printf("✗ item %v (%v) missing %v\n", cnt, r.ID, r.File)
			missing++
			continue
		}
		// iterate through each value
		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "id":
				fmt.Printf("✓ item %v (%v) ", cnt, value)
			case "uuid":
				fmt.Printf("%v, ", value)
			case "createdat":
				fmt.Printf("%v, ", value)
			case "filename":
				fmt.Printf("%v\n", value)
			case "file_zip_content":
				if col == nil || ow {
					if u := fileZipContent(r); !u {
						continue
					}
					// todo: tag platform based on found files
					err := archive.Extract(r.File, r.UUID)
					logs.Log(err)
				}
			case "deletedat":
			case "updatedat": // ignore
			default:
				fmt.Printf("   %v: %v\n", columns[i], value)
			}
		}
		fmt.Println("---------------")
	}
	logs.Check(rows.Err())
	fmt.Println("Total proofs handled: ", cnt)
	if missing > 0 {
		fmt.Println("UUID files not found: ", missing)
	}
	return nil
}

// CreateUUIDMap builds a map of all the unique UUID values stored in the Defacto2 database.
func CreateUUIDMap() (int, IDs) {
	db := Connect()
	defer db.Close()
	// query database
	var id, uuid string
	rows, err := db.Query("SELECT `id`,`uuid` FROM `files`")
	logs.Check(err)
	m := IDs{} // this map is to store all the UUID values used in the database
	// handle query results
	rc := 0 // row count
	for rows.Next() {
		err = rows.Scan(&id, &uuid)
		logs.Check(err)
		m[uuid] = Empty{} // store record `uuid` value as a key name in the map `m` with an empty value
		rc++
	}
	return rc, m
}

// GroupsX returns a list of organisations or groups.
func GroupsX(where string, count bool) {
	db := Connect()

	//<h2><a href="/g/13-omens">13 OMENS</a></h2><hr>

	// switch(arguments.key) {
	// 	case "magazine": sql = "section = 'magazine' AND"; break;
	// 	case "bbs": sql = "RIGHT(group_brand_for,4) = ' BBS' AND"; break;
	// 	case "ftp": sql = "RIGHT(group_brand_for,4) = ' FTP' AND"; break;
	// 	// This will only display groups who are listed under group_brand_for. group_brand_by only groups will be ignored
	// 	case "group": sql = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"; break;
	// }

	s := "SELECT pubValue, (SELECT CONCAT(pubCombined, ' ', '(', initialisms, ')') FROM groups WHERE pubName = pubCombined AND Length(initialisms) <> 0) AS pubCombined"
	s += " FROM (SELECT TRIM(group_brand_for) AS pubValue, group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0"
	s += " UNION SELECT TRIM(group_brand_by) AS pubValue, group_brand_by AS pubCombined FROM files WHERE Length(group_brand_by) <> 0) AS pubTbl"
	rows, err := db.Query(s)
	logs.Check(err)
	defer db.Close()
	var (
		grp         sql.NullString
		grpAndShort sql.NullString
	)
	i := 0
	g := ""
	grps := []string{}
	for rows.Next() {
		err = rows.Scan(&grp, &grpAndShort)
		logs.Check(err)
		_, errU := grp.Value()
		_, errC := grpAndShort.Value()
		if errU != nil || errC != nil {
			continue
		}
		i++
		switch grpAndShort.String {
		case "":
			g = fmt.Sprintf("%v", grp.String)
		default:
			g = fmt.Sprintf("%v", grpAndShort.String)
		}
		if count {
			g += fmt.Sprint(GroupCount(grp.String), " files ")
		}
		grps = append(grps, g)
	}
	fmt.Println(strings.Join(grps, ", "))
	fmt.Println("Total groups", i)
}

// Groups returns a list of organizations or groups filtered by name.
func Groups(name string, count bool) ([]string, int) {
	db := Connect()
	s := sqlGroups(name, false)
	// count records
	rows, err := db.Query(s)
	logs.Check(err)
	defer db.Close()
	total := 0
	for rows.Next() {
		total++
	}
	// interate through records
	rows, err = db.Query(s)
	logs.Check(err)
	defer db.Close()
	var grp sql.NullString
	i := 0
	g := ""
	grps := []string{}
	for rows.Next() {
		err = rows.Scan(&grp)
		logs.Check(err)
		if _, err = grp.Value(); err != nil {
			continue
		}
		i++
		progressSum(i, total)
		g = fmt.Sprintf("%v", grp.String)
		if count {
			g += fmt.Sprint(GroupCount(grp.String), " files ")
		}
		grps = append(grps, g)
	}
	return grps, total
}

// GroupsToHTML ...
func GroupsToHTML(name string, count bool) {
	// TODO create a cronjob flag that retrieves the most recent updateat value and compares it to now()
	// <h2><a href="/g/13-omens">13 OMENS</a></h2><hr>
	tpl := `{{range .}}{{if .Hr}}<hr>{{end}}<h2><a href="/g/{{.ID}}">{{.Name}}</a></h2>{{end}}`
	type Group struct {
		ID   string
		Name string
		Hr   bool
	}
	grp, _ := Groups(name, false)
	data := make([]Group, len(grp))
	cap := ""
	hr := false
	for i := range grp {
		n := grp[i]
		switch c := n[:1]; {
		case cap == "":
			cap = c
		case c != cap:
			cap = c
			hr = true
		default:
			hr = false
		}
		data[i] = Group{
			ID:   MakeSlug(grp[i]),
			Name: grp[i],
			Hr:   hr,
		}
	}
	t, err := template.New("h2").Parse(tpl)
	logs.Check(err)
	err = t.Execute(os.Stdout, data)
	logs.Check(err)
}

// MakeSlug takes a name and makes it into a URL friendly slug.
func MakeSlug(name string) string {
	n := name
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.ReplaceAll(n, ", ", "*")
	n = strings.ReplaceAll(n, " & ", " ampersand ")
	re := regexp.MustCompile(` ([0-9])`)
	n = re.ReplaceAllString(n, `-$1`)
	re = regexp.MustCompile(`[^A-Za-z0-9 \-\+\.\_\*]`) // remove all chars except these
	n = re.ReplaceAllString(n, ``)
	n = strings.ToLower(n)
	re = regexp.MustCompile(` ([a-z])`)
	n = re.ReplaceAllString(n, `-$1`)
	return n
}

// GroupsPrint ...
func GroupsPrint(name string, count bool) {
	g, i := Groups(name, count)
	fmt.Println(strings.Join(g, ", "))
	fmt.Println("Total groups", i)
}

func progressSum(count int, total int) {
	fmt.Printf("\rBuilding %d/%d", count, total)
}

func progressPct(count int, total int) {
	fmt.Printf("\rBuilding %.2f %%", float64(count)/float64(total)*100)
}

// GroupCount returns the number of file entries associated with a group.
func GroupCount(name string) int {
	db := Connect()
	n := name
	var count int
	s := "SELECT COUNT(*) FROM files WHERE "
	s += fmt.Sprintf("group_brand_for='%v' OR group_brand_for LIKE '%v,%%' OR group_brand_for LIKE '%%, %v,%%' OR group_brand_for LIKE '%%, %v'", n, n, n, n)
	s += fmt.Sprintf(" OR group_brand_by='%v' OR group_brand_by LIKE '%v,%%' OR group_brand_by LIKE '%%, %v,%%' OR group_brand_by LIKE '%%, %v'", n, n, n, n)
	row := db.QueryRow(s)
	err := row.Scan(&count)
	logs.Check(err)
	defer db.Close()
	return count
}

// Update is a temp SQL update func.
func Update(id string, content string) {
	db := Connect()
	defer db.Close()
	update, err := db.Prepare("UPDATE files SET file_zip_content=?,updatedat=NOW(),updatedby=?,platform=?,deletedat=NULL,deletedby=NULL WHERE id=?")
	logs.Check(err)
	r, err := update.Exec(content, UpdateID, "image", id)
	logs.Check(err)
	fmt.Println("Updated file_zip_content", r)
}

func fileZipContent(r Record) bool {
	a, err := archive.Read(r.File)
	if err != nil {
		logs.Log(err)
		return false
	}
	Update(r.ID, strings.Join(a, "\n"))
	return true
}

// readPassword attempts to read and return the Defacto2 database user password when stored in a local text file.
func readPassword() string {
	// fetch database password
	pwFile, err := os.Open(pwPath)
	// return an empty password if path fails
	if err != nil {
		//log.Print("WARNING: ", err)
		return d.Pass
	}
	defer pwFile.Close()
	pw, err := ioutil.ReadAll(pwFile)
	logs.Check(err)
	return strings.TrimSpace(fmt.Sprintf("%s", pw))
}

func validUUID(id string) bool {
	if _, err := uuid.Parse(id); err != nil {
		return false
	}
	return true
}

func validID(id string) bool {
	if _, err := strconv.Atoi(id); err != nil {
		return false
	}
	return true
}

// sqlGroups returns a complete SQL WHERE statement where the groups are filtered by name.
func sqlGroups(name string, includeSoftDeletes bool) string {
	inc := includeSoftDeletes
	c := [4]string{"bbs", "ftp", "group", "magazine"}
	skip := false
	for _, a := range c {
		if a == name {
			skip = true
		}
	}
	var sql string
	switch skip {
	case true: // disable group_brand_by listings for BBS, FTP, group, magazine filters
		sql = "SELECT DISTINCT group_brand_for AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_for) <> 0 " + sqlGroupsWhere(name, inc) + ")"
	default:
		sql = "(SELECT DISTINCT group_brand_for AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_for) <> 0 " + sqlGroupsWhere(name, inc) + ")"
		sql += " UNION "
		sql += "(SELECT DISTINCT group_brand_by AS pubCombined "
		sql += "FROM files WHERE Length(group_brand_by) <> 0 " + sqlGroupsWhere(name, inc) + ")"
	}
	return sql + " ORDER BY pubCombined"
}

// sqlGroupsFilter returns a partial SQL WHERE statement to filer groups by name.
func sqlGroupsFilter(name string) string {
	var sql string
	switch name {
	case "magazine":
		sql = "section = 'magazine' AND"
	case "bbs":
		sql = "RIGHT(group_brand_for,4) = ' BBS' AND"
	case "ftp":
		sql = "RIGHT(group_brand_for,4) = ' FTP' AND"
	case "group": // only display groups who are listed under group_brand_for, group_brand_by only groups will be ignored
		sql = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"
	}
	return sql
}

// sqlGroupsWhere returns a partial SQL WHERE statement where groups are filtered by name.
func sqlGroupsWhere(name string, includeSoftDeletes bool) string {
	sql := sqlGroupsFilter(name)
	switch {
	case sql != "" && includeSoftDeletes:
		sql = "AND " + sql
	case sql == "" && includeSoftDeletes: // do nothing
	case sql != "" && !includeSoftDeletes:
		sql = "AND " + sql + " `deletedat` IS NULL"
	default:
		sql = "AND `deletedat` IS NULL"
	}
	l := len(sql)
	if l > 4 && sql[l-4:] == " AND" {
		fmt.Printf("%q|", sql[l-4:])
		return sql[:l-4]
	}
	return sql
}
