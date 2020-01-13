package cmd

import (
	"github.com/Defacto2/df2/lib/demozoo"
	"github.com/Defacto2/df2/lib/logs"
	"github.com/spf13/cobra"
)

// demozooCmd represents the demozoo command
var demozooCmd = &cobra.Command{
	Use:   "demozoo",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Println("demozoo called")
		// data := demozoo.Fetch(179611)
		// //		data.Print()
		// p, s := data.PouetID()
		// logs.Printf("Pouet ID %v and HTTP status %v\n", p, s)
		// data.Downloads()
		var err error
		r := demozoo.Request{
			Overwrite: false,
			All:       false,
			HideMiss:  false}
		switch {
		// case proo.id != "":
		// 	err = r.Query(proo.id)
		default:
			err = r.Query("26732")
			//err = r.Queries()
		}
		logs.Check(err)
	},
}

func init() {
	rootCmd.AddCommand(demozooCmd)
}

// /**
// * Lookup and fetch the title of a Demozoo ID in JSON format.
// * @id Demozoo ID to lookup.
// * @method Either `extenalUpload`, `json` or `title`
// *
// * NOTE: append a m=dump to the queryString to dump() the data
// */
// public any function lookupDemozoo(string id="#params.key#",string method="title") {
// 	onlyProvides("json")
// 	local.json = ""
// 	local.e = []
// 	if(structKeyExists(params,"m")) arguments.method = params.m
// 	local.htmlData = downloadHTML('https://demozoo.org/api/v1/productions/#arguments.id#?format=json','GET')
// 	if(arguments.method == "dump") {
// 		dump(htmlData)
// 		abort;
// 	}
// 	if(IsSimpleValue(htmlData) && htmlData contains "404") json = "Demozoo ID #arguments.id# is invalid"
// 	else if(!IsStruct(htmlData)) json = "Demozoo API did not work"
// 	else if(!StructKeyExists(htmlData, "title")) json = "Demozoo API did not return a title"
// 	else {
// 		local.import = htmlData
// 		switch(arguments.method) {
// 			case "extenalUpload":
// 				local.fileExist = model("Upload").Count(where="web_id_demozoo=#params.key#",includeSoftDeletes=true)
// 				if(fileExist gt 0) json = "Demozoo ID #arguments.id# already exists on Defacto2";
// 				else json = _parseAPIDemozoo(import)
// 			break;
// 			case "json": json = local.import
// 			break;
// 			case "title":
// 				json = htmlData.title
// 				if(StructKeyExists(htmlData, "author_nicks")) {
// 					for (e in htmlData.author_nicks) {
// 						if(structKeyExists(e.releaser, "is_group") && e.releaser.is_group eq true) {
// 							if(json == htmlData.title) json = "#json# by #e.name#"
// 							else json = "#json#, #e.name#"
// 						}
// 					}
// 				}
// 			break;
// 		}
// 	}
// 	renderText(SerializeJSON(json))
// }

// ControlCnts.introlesspouet = model("Files").count(where=dynamicSqlForFile(params={key='introlesspouet',section='',platform=''}).WHERESTATE,includeSoftDeletes=true)
// web_id_demozoo, web_id_pouet
//

/**
 * Parse the Demozoo JSON response
 *
 * @import Demozoo JSON data as a structure
 */
//  private any function _parseAPIDemozoo(required struct import) {
// 	local.json = []
// 	json[1] = {id="record_title",v=""}
// 	json[2] = {id="group_brand_for",v=""}
// 	json[3] = {id="group_brand_by",v=""}
// 	json[4] = {id="date_issued_day",v=""}
// 	json[5] = {id="date_issued_month",v=""}
// 	json[6] = {id="date_issued_year",v=""}
// 	json[7] = {id="credit_program",v=""}
// 	json[8] = {id="credit_illustration",v=""}
// 	json[9] = {id="credit_audio",v=""}
// 	json[10] = {id="platform",v=""}
// 	json[11] = {id="section",v=""}
// 	json[12] = {id="list_links",v=""}
// 	json[13] = {id="web_id_pouet",v=""}
// 	json[14] = {id="credit_text",v=""}
// 	json[15] = {id="thumbnail1",v=""}
// 	json[16] = {id="thumbnail2",v=""}
// 	json[17] = {id="thumbnail3",v=""}
// 	json[18] = {id="web_id_youtube",v=""}
// 	json[19] = {id="comment",v=""}
// 	// add youtube
// 	local.regx
// 	local.regy
// 	local.item
// 	local.find = 0
// 	local.slugs = []
// 	// platform (also checks validity)
// 	if(arrayLen(import.platforms) == 0) return "The production '#import.title#' has no platform so it does not look suitable for Defacto2";
// 	slugs = ["Browser","Java","Linux","MS-Dos","Windows"]
// 	find = 0
// 	for(item in import.platforms) {
// 		find = arrayContainsNoCase(slugs, item.name)
// 		if(find > 0) slug = "true";
// 		else slug = item.name
// 	}
// 	if(slug != "true") return "The #slug# platform used by the production '#import.title#' does not look suitable for Defacto2";
// 	slugs = ["html","java","linux","dOS","windows"]
// 	json[10].v = slugs[find]
// 	// check types (also checks validity)
// 	slug = ""
// 	slugs = ["Diskmag","Textmag","Game","Intro","Demo","BBStro","Cracktro","Tool"]
// 	for (item in import.types) {
// 		find = arrayContainsNoCase(slugs, item.name)
// 		if(find eq 0 and Right(item.name,5) eq "Intro") {
// 			slug = "true"
// 			find = 4
// 			break
// 		}
// 		if(find > 0) { slug = "true"; break; }
// 		else slug = item.name
// 	}
// 	if(slug != "true") return "The #slug# type used with the production '#import.title#' does not look suitable for Defacto2";
// 	slugs = ["magazine","magazine","dEMO","dEMO","dEMO","bbs","releaseadvert","programmingtool"]
// 	json[11].v = slugs[find]
// 	// sanitise data
// 	json[1].v = import.title
// 	// groups
// 	find = 0
// 	for (item in import.author_nicks) {
// 		find += 1
// 		if(find == 1) json[2].v = item.name
// 		if(find == 2) json[3].v = item.name
// 		if(find > 2) break;
// 	}
// 	// handle Demozoo's quirks
// 	// if title contains 'x FTP' or 'x BBS' then we swap data
// 	if (right(import.title,4) == ' BBS' || right(import.title,4) == ' FTP') {
// 		if (find == 0) {
// 			json[1].v = ''
// 			json[2].v = import.title
// 		} else if (Len(json[3].v) == 0) {
// 			// blank record_title
// 			json[1].v = ''
// 			// copy group_brand_for to group_brand_by
// 			json[3].v = json[2].v
// 			// set group_brand_for to eq title
// 			json[2].v = import.title
// 		}
// 	}
// 	else if (import.title contains 'application generator') {
// 		json[11].v = "groupapplication"
// 	}
// 	// else if title contains 'x FTP (?)' or 'x BBS (?)' then swap data
// 	else {
// 		regx = reFindNoCase(' BBS \([^\d]*(\d+)[^\d]*\)$', import.title)
// 		regy = reFindNoCase(' FTP \([^\d]*(\d+)[^\d]*\)$', import.title)
// 		if (regx > 0) {
// 			// blank record_title
// 			json[1].v = ''
// 			// copy group_brand_for to group_brand_by
// 			json[3].v = json[2].v
// 			// set group_brand_for to eq title with ' BBS (?)' dropped
// 			json[2].v = left(import.title, regx) & ' BBS'
// 		}
// 		else if (regy > 0) {
// 			// blank record_title
// 			json[1].v = ''
// 			// copy group_brand_for to group_brand_by
// 			json[3].v = json[2].v
// 			// set group_brand_for to eq title with ' FTP (?)' dropped
// 			json[2].v = left(import.title, regx) & ' FTP'
// 		}
// 	}
// 	// dates
// 	if(isNull(import.release_date) == false) {
// 		slugs = listToArray(import.release_date,"-")
// 		if(arrayLen(slugs) >= 3) json[4].v = slugs[3]
// 		if(arrayLen(slugs) >= 2) json[5].v = NumberFormat(slugs[2])
// 		if(arrayLen(slugs) >= 1) json[6].v = slugs[1]
// 	}
// 	// credits
// 	for(item in import.credits) {
// 		switch(item.category) {
// 			case "Code": json[7].v = listAppend(json[7].v,item.nick.name);
// 			break;
// 			case "Graphics": json[8].v = listAppend(json[8].v,item.nick.name);
// 			break;
// 			case "Music": json[9].v = listAppend(json[9].v,item.nick.name);
// 			break;
// 			case "Text": json[14].v = listAppend(json[14].v,item.nick.name);
// 			break;
// 		}
// 	}
// 	// download link
// 	if(arrayLen(import.download_links)) json[12].v = import.download_links[1].url
// 	// screenshots
// 	find = 0
// 	for(item in import.screenshots) {
// 		find += 1
// 		if(find EQ 1) json[15].v = item.thumbnail_url
// 		if(find EQ 2) json[16].v = item.thumbnail_url
// 		if(find EQ 3) json[17].v = item.thumbnail_url
// 		if(find GT 3) break;
// 	}
// 	return json
// }
