package cmd

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/Defacto2/df2/lib/database"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

const resource string = "https://defacto2.net/f/"

// url composes the <url> tag in the sitemap
type url struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"` // optional
}

// Urlset is a sitemap XML template
type Urlset struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	Svs     []url    `xml:"url"`
}

// sitemapCmd represents the sitemap command
var sitemapCmd = &cobra.Command{
	Use:   "sitemap",
	Short: "An site map generator",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sitemap called")
		create()
	},
}

func init() {
	rootCmd.AddCommand(sitemapCmd)
}

// create generates and prints the sitemap.
func create() {
	// query
	var id string
	var createdat sql.NullString
	var updatedat sql.NullString
	db := database.Connect()
	rows, err := db.Query("SELECT `id`,`createdat`,`updatedat` FROM `files` WHERE `deletedat` IS NULL")
	logs.Check(err)
	defer db.Close()

	v := &Urlset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}

	// handle query results
	for rows.Next() {
		err = rows.Scan(&id, &createdat, &updatedat)
		logs.Check(err)
		// check for valid createdat and updatedat entries
		_, errU := updatedat.Value()
		_, errC := createdat.Value()
		if errU != nil || errC != nil {
			continue // skip record (log in future?)
		}
		// parse createdat and updatedat to use in the <lastmod> tag
		var lastmod string
		if udValid := updatedat.Valid; udValid {
			lastmod = updatedat.String
		} else if cdValid := createdat.Valid; cdValid {
			lastmod = createdat.String
		}
		lastmodFields := strings.Fields(lastmod)
		// append url
		var lastmodValue string // blank by default; <lastmod> tag has `omitempty` set, so it won't display if no value is given
		if len(lastmodFields) > 0 {
			lastmodValue = lastmodFields[0]
		}
		v.Svs = append(v.Svs, url{fmt.Sprintf("%v%v", resource, obfuscateParam(id)), lastmodValue})
	}
	output, err := xml.MarshalIndent(v, "", "")
	logs.Check(err)
	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
}

// obfuscateParam hides the param value using the method implemented in CFWheels obfuscateParam() helper.
func obfuscateParam(param string) string {
	rv := param // return value
	// check to make sure param doesn't begin with a 0 digit
	if rv0 := rv[0]; rv0 == '0' {
		return rv
	}
	paramInt, err := strconv.Atoi(param) // convert param to an int type
	if err != nil {
		return rv
	}
	iEnd := len(rv) // count the number of digits in param
	afloat64 := math.Pow10(iEnd) + float64(reverseInt(paramInt))
	// keep a and b as int type
	a := int(afloat64)
	b := 0
	for i := 1; i <= iEnd; i++ {
		// slice individual digits from param and sum them
		paramSlice, err := strconv.Atoi(string(param[iEnd-i]))
		if err != nil {
			return rv
		}
		b += paramSlice
	}
	// base 64 conversion
	a ^= 461
	b += 154
	return strconv.FormatInt(int64(b), 16) + strconv.FormatInt(int64(a), 16)
}

// reverseInt swaps the direction of the value, 12345 would return 54321.
func reverseInt(value int) int {
	int := strconv.Itoa(value)
	new := ""
	for x := len(int); x > 0; x-- {
		new += string(int[x-1])
	}
	newInt, err := strconv.Atoi(new)
	logs.Check(err)
	return newInt
}
