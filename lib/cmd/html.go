package cmd

import (
	"fmt"

	"github.com/Defacto2/uuid/v2/lib/database"
	"github.com/spf13/cobra"
)

const enforced = "-,art,bbs,demo,ftp,group,intro,magazine"

// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "A HTML page and template generator",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("html called")
		database.GroupsToHTML("all", false)
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)
}

// switch(arguments.key) {
// 	case "magazine": sql = "section = 'magazine' AND"; break;
// 	case "bbs": sql = "RIGHT(group_brand_for,4) = ' BBS' AND"; break;
// 	case "ftp": sql = "RIGHT(group_brand_for,4) = ' FTP' AND"; break;
// 	// This will only display groups who are listed under group_brand_for. group_brand_by only groups will be ignored
// 	case "group": sql = "RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND"; break;
// }
// function getOrganisations(string key, boolean includeSoftDeletes) {
// 	local.q = local.sql = ""
// 	savecontent variable="sql" {
// 		// disable group_brand_by listings for BBS, FTP, group, magazine filters
// 		if(arguments.key != 'bbs' && arguments.key != 'ftp' && arguments.key != 'group' && arguments.key != 'magazine') {
// 			writeOutput("(SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 #_extraWHERE(arguments.key,arguments.includeSoftDeletes)#)")
// 			writeOutput(" UNION")
// 			writeOutput(" (SELECT DISTINCT group_brand_by AS pubCombined	FROM files WHERE Length(group_brand_by) <> 0 #_extraWHERE(arguments.key,arguments.includeSoftDeletes)#)")
// 		} else {
// 			writeOutput(" SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 #_extraWHERE(arguments.key,arguments.includeSoftDeletes)#")
// 		}
// 		writeOutput(" ORDER BY pubCombined")
// 	}
// 	q = queryExecute(sql, [], { datasource=get('dataSourceName') } );
// 	return q;
// }

// /**
// * Custom SQL to get a list of groups with initialisms formatted as 'Advanced Bitchin Cats (ABC)'
// */
// function getOrganisationsInit() {
// 	local.q = local.sql = ""
// 	savecontent variable="sql" {
// 		writeOutput("SELECT pubValue, (SELECT CONCAT(pubCombined, ' ', '(', initialisms, ')') FROM groups WHERE pubName = pubCombined AND Length(initialisms) <> 0) AS pubCombined")
// 		writeOutput(" FROM (SELECT TRIM(group_brand_for) AS pubValue, group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0")
// 		writeOutput(" UNION SELECT TRIM(group_brand_by) AS pubValue, group_brand_by AS pubCombined FROM files WHERE Length(group_brand_by) <> 0) AS pubTbl")
// 		writeOutput(" ORDER BY pubTbl.pubValue")
// 	}
// 	q = queryExecute(sql, [], { datasource=get('dataSourceName') } );
// 	return q;
// }

// switch(params.key) {
// 	case 'bbs': header = title = crumb = 'Bulletin Board Sites'; break;
// 	case 'ftp': header = title = crumb = 'Internet FTP sites'; break;
// 	case 'group': header = title = 'Scene groups'; break;
// 	case 'magazine': header = title = 'Magazines'; break;
// 	default: title = 'Groups ' & get('myapp')['menuorganisation#params.key#'].name; break;
// }
// 	if(len(crumb) == 0) crumb = LCase(header)
// 	breadcrumbs &= RDFaCrumb(3, crumb, urlFor(route='organisationFilter', key=params.key))
// 	description = 'A list of #LCase(title)#'
// 	if(params.key eq 'bbs') description &= ' (BBS)'
// 	if(params.key eq 'ftp') description &= ' (FTP)'
// } else {
// 	description = 'A list of all groups and organisations'
// }
// canonical = "/organisation/list/-"
// local.organisationQuery = model("Organisation").getOrganisations(key=params.key,returnAs="query",includeSoftDeletes=includeSoftDeletes);
// listOfResults = groupsListed(organisationQuery)
// listOfResults = listSort(listOfResults, "text", "asc")
