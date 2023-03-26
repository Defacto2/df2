package templ

// TableData is a container for the tableTmpl template.
type TableData struct {
	VER    string
	CREATE string
	TABLE  string
	INSERT string
	SQL    string
	UPDATE string
}

// TablesTmp is a container for the tablesTmpl template.
type TablesTmp struct {
	VER    string
	DB     string
	CREATE []TablesData
}

// TablesData is a data container for the tablesTmpl template.
type TablesData struct {
	Columns string
	Rows    string
	Table   string
}

const (
	CountFiles   = "SELECT COUNT(*) FROM `files`"
	CountWaiting = CountFiles + " WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

	SelKeys   = "SELECT `id` FROM `files`"
	SelNames  = "SELECT `filename` FROM `files`"
	SelUpdate = "SELECT `updatedat` FROM `files`" +
		" WHERE `createdat` <> `updatedat` AND `deletedby` IS NULL" +
		" ORDER BY `updatedat` DESC LIMIT 1"

	WhereDownloadBlock = "WHERE `file_security_alert_url` IS NOT NULL AND `file_security_alert_url` != ''"
	WhereAvailable     = "WHERE `deletedat` IS NULL"
	WhereHidden        = "WHERE `deletedat` IS NOT NULL"
)

const SelNewFiles = "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`filesize`," +
	"`web_id_demozoo`,`file_zip_content`,`updatedat`,`platform`,`file_integrity_strong`," +
	"`file_integrity_weak`,`web_id_pouet`,`group_brand_for`,`group_brand_by`,`section`\n" +
	"FROM `files`\n" +
	"WHERE `deletedby` IS NULL AND `deletedat` IS NOT NULL"

const Table = `
-- df2 v{{.VER}} Defacto2 MySQL {{.TABLE}} dump
-- source:        https://defacto2.net/sql
-- documentation: https://github.com/Defacto2/database

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
{{.CREATE}}
INSERT INTO {{.TABLE}} ({{.INSERT}}) VALUES
{{.SQL}}{{.UPDATE}};

-- {{now}}
`

const Tables = `
-- df2 v{{.VER}} Defacto2 MySQL complete dump
-- source:        https://defacto2.net/sql
-- documentation: https://github.com/Defacto2/database

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
{{.DB}}{{range .CREATE}}{{.Columns}}
{{.Rows}}
{{end}}
-- {{now}}
`

// as there is no escape feature for `raw literals`, these SQL CREATE statements append standard strings.

const NewDB = "\nDROP DATABASE IF EXISTS `defacto2-inno`;\n" +
	"CREATE DATABASE `defacto2-inno` /*!40100 DEFAULT CHARACTER SET utf8 */;\n" +
	"USE `defacto2-inno`;\n"

const NewFiles = "\nDROP TABLE IF EXISTS `files`;\n" +
	"CREATE TABLE `files` (\n" +
	"  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary key',\n" +
	"  `uuid` char(36) DEFAULT 'UUID()' COMMENT 'Global identifier',\n" +
	"  `list_relations` varchar(255) DEFAULT NULL COMMENT 'List of associated records',\n" +
	"  `web_id_github` varchar(1024) DEFAULT NULL COMMENT 'Id for a GitHub repository',\n" +
	"  `web_id_youtube` char(11) DEFAULT NULL COMMENT 'Id for a related YouTube video',\n" +
	"  `web_id_pouet` int(6) DEFAULT NULL COMMENT 'Id for a Pouët record',\n" +
	"  `web_id_demozoo` int(6) DEFAULT NULL COMMENT 'Id for a Demozoo record',\n" +
	"  `group_brand_for` varchar(100) DEFAULT NULL COMMENT 'Group or brand used to credit the file',\n" +
	"  `group_brand_by` varchar(100) DEFAULT NULL COMMENT" +
	" 'Optional alternative Group or brand used to credit the file',\n" +
	"  `record_title` varchar(100) DEFAULT NULL COMMENT 'Display title or magazine edition',\n" +
	"  `date_issued_year` smallint(4) DEFAULT NULL COMMENT 'Published date year',\n" +
	"  `date_issued_month` tinyint(2) DEFAULT NULL COMMENT 'Published date month',\n" +
	"  `date_issued_day` tinyint(2) DEFAULT NULL COMMENT 'Published date day',\n" +
	"  `credit_text` varchar(1024) DEFAULT NULL COMMENT 'Writing credits',\n" +
	"  `credit_program` varchar(100) DEFAULT NULL COMMENT 'Programming credits',\n" +
	"  `credit_illustration` varchar(1024) DEFAULT NULL COMMENT 'Artist credits',\n" +
	"  `credit_audio` varchar(100) DEFAULT NULL COMMENT 'Composer credits',\n" +
	"  `filename` varchar(255) DEFAULT NULL COMMENT 'File name',\n" +
	"  `filesize` int(11) DEFAULT NULL COMMENT 'Size of the file in bytes',\n" +
	"  `list_links` varchar(2048) DEFAULT NULL COMMENT 'List of URLs related to this file',\n" +
	"  `file_security_alert_url` varchar(256) DEFAULT NULL COMMENT 'URL showing results of a virus scan',\n" +
	"  `file_zip_content` longtext COMMENT 'Content of archive',\n" +
	"  `file_magic_type` varchar(255) DEFAULT NULL COMMENT 'File type meta data',\n" +
	"  `preview_image` varchar(1024) DEFAULT NULL COMMENT 'Internal file to use as a screenshot',\n" +
	"  `file_integrity_strong` char(96) DEFAULT NULL COMMENT 'SHA384 hash of file',\n" +
	"  `file_integrity_weak` char(32) DEFAULT NULL COMMENT 'MD5 hash of file',\n" +
	"  `file_last_modified` datetime DEFAULT NULL COMMENT 'Date last modified attribute saved to file',\n" +
	"  `platform` char(25) DEFAULT NULL COMMENT 'Computer platform',\n" +
	"  `section` char(25) DEFAULT NULL COMMENT 'Category',\n" +
	"  `comment` text COMMENT 'Description',\n" +
	"  `createdat` datetime DEFAULT NULL COMMENT 'Timestamp when record was created',\n" +
	"  `updatedat` datetime DEFAULT NULL COMMENT 'Timestamp when record was revised',\n" +
	"  `deletedat` datetime DEFAULT NULL COMMENT 'Timestamp used to ignore record',\n" +
	"  `updatedby` char(36) DEFAULT NULL COMMENT 'UUID of the user who last updated this record',\n" +
	"  `deletedby` char(36) DEFAULT NULL COMMENT 'UUID of the user who removed this record',\n" +
	"  `retrotxt_readme` varchar(255) DEFAULT NULL COMMENT 'Text file contained in archive to display',\n" +
	"  `retrotxt_no_readme` tinyint(2) DEFAULT NULL COMMENT 'Disable the use of RetroTxt',\n" +
	"  `dosee_run_program` varchar(255) DEFAULT NULL COMMENT 'Program contained in archive to run in DOSBox',\n" +
	"  `dosee_hardware_cpu` varchar(6) DEFAULT NULL COMMENT 'DOSee turn off expanded memory (EMS)',\n" +
	"  `dosee_hardware_graphic` varchar(8) DEFAULT NULL COMMENT 'DOSee graphics/machine override',\n" +
	"  `dosee_hardware_audio` varchar(9) DEFAULT NULL COMMENT 'DOSee audio override',\n" +
	"  `dosee_no_aspect_ratio_fix` tinyint(2) DEFAULT NULL COMMENT 'DOSee disable aspect ratio corrections',\n" +
	"  `dosee_incompatible` tinyint(2) DEFAULT NULL COMMENT 'Flag DOS program as incompatible for DOSBox',\n" +
	"  `dosee_no_ems` tinyint(2) DEFAULT NULL COMMENT 'DOSBox turn off EMS',\n" +
	"  `dosee_no_xms` tinyint(2) DEFAULT NULL COMMENT 'DOSee turn off extended memory (XMS)',\n" +
	"  `dosee_no_umb` tinyint(2) DEFAULT NULL COMMENT 'DOSee turn off upper memory block access (UMB)',\n" +
	"  `dosee_load_utilities` tinyint(2) DEFAULT NULL COMMENT 'DOSee load utilities',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `Browsing` (`date_issued_year`,`date_issued_month`,`date_issued_day`,`section`," +
	"`platform`,`filename`(191),`createdat`),\n" +
	"  FULLTEXT KEY `pubfor_pubby_pubedition_filename_comment` (`group_brand_for`,`group_brand_by`" +
	",`record_title`,`filename`,`comment`)\n" +
	") ENGINE=InnoDB AUTO_INCREMENT=31945 DEFAULT CHARSET=utf8 COMMENT='This database is the" +
	" complete collection of files for download';\n" +
	"\nTRUNCATE `files`;"

const NewGroups = "\nDROP TABLE IF EXISTS `groups`;\n" +
	"CREATE TABLE `groups` (\n" +
	"  `id` smallint(6) NOT NULL AUTO_INCREMENT COMMENT 'Primary key',\n" +
	"  `pubname` varchar(100) NOT NULL COMMENT 'Group or brand',\n" +
	"  `initialisms` varchar(255) NOT NULL COMMENT 'Initialisms or acronym',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `pubname` (`pubname`),\n" +
	"  KEY `id` (`id`)\n" +
	") ENGINE=InnoDB AUTO_INCREMENT=341 DEFAULT CHARSET=utf8 COMMENT='Initialism for groups';\n" +
	"\nTRUNCATE `groups`;"

const NewNetresources = "\nDROP TABLE IF EXISTS `netresources`;\n" +
	"CREATE TABLE `netresources` (\n" +
	"  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary key',\n" +
	"  `uuid` char(36) DEFAULT 'UUID()' COMMENT 'Global identifier',\n" +
	"  `legacyid` int(11) DEFAULT NULL COMMENT 'Former ids from defacto2net database',\n" +
	"  `httpstatuscode` int(3) DEFAULT NULL COMMENT 'Status code definition',\n" +
	"  `httpstatustext` varchar(255) DEFAULT NULL COMMENT 'Status code text',\n" +
	"  `httplocation` varchar(255) DEFAULT NULL COMMENT 'URI given by 301,302,303 codes',\n" +
	"  `httpetag` varchar(50) DEFAULT NULL COMMENT 'Hash key used for cache',\n" +
	"  `httplastmodified` varchar(100) DEFAULT NULL COMMENT 'Date used for cache',\n" +
	"  `metatitle` varchar(1000) DEFAULT NULL COMMENT 'Title metadata',\n" +
	"  `metadescription` varchar(1000) DEFAULT NULL COMMENT 'Description metadata',\n" +
	"  `metaauthors` varchar(1000) DEFAULT NULL COMMENT 'Authors metadata',\n" +
	"  `metakeywords` varchar(1000) DEFAULT NULL COMMENT 'Keywords metadata',\n" +
	"  `uriref` varchar(255) DEFAULT NULL COMMENT 'URL of the resource',\n" +
	"  `title` varchar(255) DEFAULT NULL COMMENT 'Title of resource',\n" +
	"  `date_issued_year` smallint(4) DEFAULT NULL,\n" +
	"  `date_issued_month` tinyint(2) DEFAULT NULL,\n" +
	"  `date_issued_day` tinyint(2) DEFAULT NULL,\n" +
	"  `comment` mediumtext COMMENT 'Default description when metadescription is empty',\n" +
	"  `categorykey` varchar(25) DEFAULT NULL COMMENT 'Category',\n" +
	"  `categorysort` varchar(25) DEFAULT NULL COMMENT 'Sorting category',\n" +
	"  `deletedat` datetime DEFAULT NULL COMMENT 'Timestamp used to disable record',\n" +
	"  `deletedatcomment` varchar(255) DEFAULT NULL COMMENT 'Reason for record to be disabled',\n" +
	"  `createdat` datetime DEFAULT NULL COMMENT 'Timestamp when record was created',\n" +
	"  `updatedat` datetime DEFAULT NULL COMMENT 'Timestamp when record was revised',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB AUTO_INCREMENT=696 DEFAULT CHARSET=utf8 COMMENT='Scene websites';\n" +
	"\nTRUNCATE `netresources`;"

const NewUsers = "\nDROP TABLE IF EXISTS `users`;\n" +
	"CREATE TABLE `users` (\n" +
	"  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Primary key',\n" +
	"  `uuid` char(36) DEFAULT 'UUID()' COMMENT 'Global identifier',\n" +
	"  `username` varchar(255) NOT NULL COMMENT 'Sign-in name',\n" +
	"  `password` varchar(128) DEFAULT NULL COMMENT 'Sign-in password',\n" +
	"  `displayname` varchar(255) DEFAULT NULL COMMENT 'Name or alias for display',\n" +
	"  `affiliation` varchar(255) DEFAULT NULL COMMENT 'User groups and affiliations',\n" +
	"  `email` varchar(255) DEFAULT NULL COMMENT 'Contact email address',\n" +
	"  `sysop` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'System operator toggle',\n" +
	"  `coop` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Co-operator toggle',\n" +
	"  `sessionfd` char(7) DEFAULT NULL COMMENT 'unused?',\n" +
	"  `createdat` datetime DEFAULT NULL COMMENT 'Timestamp when record was created',\n" +
	"  `updatedat` datetime DEFAULT NULL COMMENT 'Timestamp when record was revised',\n" +
	"  `deletedat` datetime DEFAULT NULL COMMENT 'Timestamp used to ignore record',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `username` (`username`)\n" +
	") ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 COMMENT='Operator user accounts';\n" +
	"\nTRUNCATE `users`;"
