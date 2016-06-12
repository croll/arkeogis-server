package model

import (
	"time"
	"database/sql"
)

type Charac struct {
	Id	int	`db:"id" json:"id"`
	Parent_id	int	`db:"parent_id" json:"parent_id"`
	Order	int	`db:"order" json:"order"`
	Author_user_id	int	`db:"author_user_id" json:"author_user_id"`	// User.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Charac_tr struct {
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Charac_id	int	`db:"charac_id" json:"charac_id"`	// Charac.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Chronology struct {
	Id	int	`db:"id" json:"id"`
	Parent_id	int	`db:"parent_id" json:"parent_id"`	// Chronology.Id
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
	Color	string	`db:"color" json:"color"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Chronology_tr struct {
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Chronology_id	int	`db:"chronology_id" json:"chronology_id"`	// Chronology.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type City struct {
	Geonameid	int	`db:"geonameid" json:"geonameid"`
	Country_geonameid	int	`db:"country_geonameid" json:"country_geonameid"`	// Country.Geonameid
	Geom	sql.NullString	`db:"geom" json:"geom"`
	Geom_centroid	string	`db:"geom_centroid" json:"geom_centroid"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type City_tr struct {
	City_geonameid	int	`db:"city_geonameid" json:"city_geonameid"`	// City.Geonameid
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Name_ascii	string	`db:"name_ascii" json:"name_ascii"`
}


type Company struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name"`
	City_geonameid	int	`db:"city_geonameid" json:"city_geonameid"`	// City.Geonameid
}


type Continent struct {
	Geonameid	int	`db:"geonameid" json:"geonameid"`
	Iso_code	string	`db:"iso_code" json:"iso_code"`
	Geom	sql.NullString	`db:"geom" json:"geom"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Continent_tr struct {
	Continent_geonameid	int	`db:"continent_geonameid" json:"continent_geonameid"`
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Name_ascii	string	`db:"name_ascii" json:"name_ascii"`
}


type Country struct {
	Geonameid	int	`db:"geonameid" json:"geonameid"`
	Iso_code	sql.NullString	`db:"iso_code" json:"iso_code"`
	Geom	sql.NullString	`db:"geom" json:"geom"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Country_tr struct {
	Country_geonameid	int	`db:"country_geonameid" json:"country_geonameid"`	// Country.Geonameid
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Name_ascii	string	`db:"name_ascii" json:"name_ascii"`
}


type Database struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name" min:"1" max:"255" error:"DATABASE.FIELD_NAME.T_CHECK_MANDATORY"`
	Scale_resolution	string	`db:"scale_resolution" json:"scale_resolution" enum:"undefined,object,site,watershed,micro-region,region,country,continent,world" error:"DATABASE.FIELD_SCALE_RESOLUTION.T_CHECK_INCORRECT"`
	Geographical_extent	string	`db:"geographical_extent" json:"geographical_extent" enum:"undefined,country,continent,international_waters,world" error:"DATABASE.FIELD_GEOGRAPHICAL_EXTENT.T_CHECK_INCORRECT"`
	Type	string	`db:"type" json:"type" enum:"undefined,inventory,research,literary-work" error:"DATABASE.FIELD_TYPE.T_CHECK_INCORRECT"`
	Owner	int	`db:"owner" json:"owner"`	// User.Id
	Editor	string	`db:"editor" json:"editor"`
	Contributor	string	`db:"contributor" json:"contributor"`
	Default_language	int	`db:"default_language" json:"default_language"`	// Lang.Id
	State	string	`db:"state" json:"state" enum:"undefined,in-progress,finished" error:"DATABASE.FIELD_STATE.T_CHECK_INCORRECT"`
	License_id	int	`db:"license_id" json:"license_id"`	// License.Id
	Published	bool	`db:"published" json:"published"`
	Soft_deleted	bool	`db:"soft_deleted" json:"soft_deleted"`
	Declared_creation_date	time.Time	`db:"declared_creation_date" json:"declared_creation_date"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Database__authors struct {
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
}


type Database__continent struct {
	Database_id	int	`db:"database_id" json:"database_id" xmltopsql:"ondelete:cascade"`	// Database.Id
	Continent_geonameid	int	`db:"continent_geonameid" json:"continent_geonameid"`	// Continent.Geonameid
}


type Database__country struct {
	Database_id	int	`db:"database_id" json:"database_id" xmltopsql:"ondelete:cascade"`	// Database.Id
	Country_geonameid	int	`db:"country_geonameid" json:"country_geonameid"`	// Country.Geonameid
}


type Database_context struct {
	Id	int	`db:"id" json:"id"`
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
	Context	string	`db:"context" json:"context"`
}


type Database_handle struct {
	Id	int	`db:"id" json:"id"`
	Database_id	int	`db:"database_id" json:"database_id" xmltopsql:"ondelete:cascade"`	// Database.Id
	Import_id	int	`db:"import_id" json:"import_id"`	// Import.Id
	Identifier	string	`db:"identifier" json:"identifier"`
	Url	string	`db:"url" json:"url"`
	Declared_creation_date	time.Time	`db:"declared_creation_date" json:"declared_creation_date"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
}


type Database_tr struct {
	Database_id	int	`db:"database_id" json:"database_id" xmltopsql:"ondelete:cascade"`	// Database.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Description	string	`db:"description" json:"description"`
	Geographical_limit	string	`db:"geographical_limit" json:"geographical_limit"`
	Bibliography	string	`db:"bibliography" json:"bibliography"`
	Context_description	string	`db:"context_description" json:"context_description"`
	Source_description	string	`db:"source_description" json:"source_description"`
	Source_relation	string	`db:"source_relation" json:"source_relation"`
	Copyright	string	`db:"copyright" json:"copyright"`
	Subject	string	`db:"subject" json:"subject"`
}


type Geographical_zone struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name"`
	Geom	string	`db:"geom" json:"geom"`
}


type Global_project struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
}


type Group struct {
	Id	int	`db:"id" json:"id"`
	Type	string	`db:"type" json:"type"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Group__permission struct {
	Group_id	int	`db:"group_id" json:"group_id"`	// Group.Id
	Permission_id	int	`db:"permission_id" json:"permission_id"`	// Permission.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Group_tr struct {
	Group_id	int	`db:"group_id" json:"group_id"`	// Group.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Import struct {
	Id	int	`db:"id" json:"id"`
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Filename	string	`db:"filename" json:"filename"`
	Number_of_lines	int	`db:"number_of_lines" json:"number_of_lines"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
}


type Lang struct {
	Id	int	`db:"id" json:"id"`
	Iso_code	string	`db:"iso_code" json:"iso_code"`
	Active	bool	`db:"active" json:"active"`
}


type Lang_tr struct {
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Lang_id_tr	int	`db:"lang_id_tr" json:"lang_id_tr"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
}


type License struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name"`
	Url	string	`db:"url" json:"url"`
}


type Permission struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Permission_tr struct {
	Permission_id	int	`db:"permission_id" json:"permission_id"`	// Permission.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Photo struct {
	Id	int	`db:"id" json:"id"`
	Photo	string	`db:"photo" json:"photo"`
}


type Project struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name" min:"1" error:"PROJECT.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"PROJECT.FIELD_NAME.T_CHECK_INCORRECT"`
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Geographical_zone_id	int	`db:"geographical_zone_id" json:"geographical_zone_id"`	// Geographical_zone.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Project__charac struct {
	Project__characs_set_id	int	`db:"project__characs_set_id" json:"project__characs_set_id"`	// Project__characs_set.Id
	Charac_id	int	`db:"charac_id" json:"charac_id"`	// Charac.Id
}


type Project__characs_set struct {
	Id	int	`db:"id" json:"id"`
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Copied_from_project__charac_set_id	int	`db:"copied_from_project__charac_set_id" json:"copied_from_project__charac_set_id"`	// Project__characs_set.Id
}


type Project__characs_set_tr struct {
	Project__characs_set_id	int	`db:"project__characs_set_id" json:"project__characs_set_id"`	// Project__characs_set.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Project__chronology struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Chronology_root_id	int	`db:"chronology_root_id" json:"chronology_root_id"`	// Chronology.Id
	Id_group	int	`db:"id_group" json:"id_group"`	// Group.Id
}


type Project__chronology_tr struct {
	Project__chronology_id	int	`db:"project__chronology_id" json:"project__chronology_id"`
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Project__databases struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
}


type Project__shapefile struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Shapefile_id	int	`db:"shapefile_id" json:"shapefile_id"`	// Shapefile.Id
}


type Project__wms_map struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Wms_map_id	int	`db:"wms_map_id" json:"wms_map_id"`	// Wms_map.Id
}


type Session struct {
	Token	string	`db:"token" json:"token"`
	Value	string	`db:"value" json:"value"`
}


type Shapefile struct {
	Id	int	`db:"id" json:"id"`
	Creator_user_id	int	`db:"creator_user_id" json:"creator_user_id"`	// User.Id
	Source_creation_date	time.Time	`db:"source_creation_date" json:"source_creation_date"`
	Filename	sql.NullString	`db:"filename" json:"filename"`
	Geom	string	`db:"geom" json:"geom"`
	Min_scale	int	`db:"min_scale" json:"min_scale"`
	Max_scale	int	`db:"max_scale" json:"max_scale"`
	Start_date1	int	`db:"start_date1" json:"start_date1"`
	Start_date2	int	`db:"start_date2" json:"start_date2"`
	End_date1	int	`db:"end_date1" json:"end_date1"`
	End_date2	int	`db:"end_date2" json:"end_date2"`
	Published	bool	`db:"published" json:"published" enum:"0,1" error:"SHAPEFILE.FIELD_PUBLISHED.T_CHECK_INCORRECT"`
	License_name	string	`db:"license_name" json:"license_name"`
	Odbl_license	bool	`db:"odbl_license" json:"odbl_license" enum:"0,1" error:"SHAPEFILE.FIELD_ODBL_LICENSE.T_CHECK_INCORRECT"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Shapefile_authors struct {
	Shapefile_id	int	`db:"shapefile_id" json:"shapefile_id"`	// Shapefile.Id
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
}


type Shapefile_tr struct {
	Shapefile_id	int	`db:"shapefile_id" json:"shapefile_id"`	// Shapefile.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name" min:"1" error:"SHAPEFILE.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"SHAPEFILE_TR.FIELD_NAME.T_CHECK_INCORRECT"`
	Description	string	`db:"description" json:"description"`
	Attribution	string	`db:"attribution" json:"attribution"`
	Bibiography	string	`db:"bibiography" json:"bibiography"`
}


type Site struct {
	Id	int	`db:"id" json:"id"`
	Code	string	`db:"code" json:"code" min:"1" error:"SITE.FIELD_CODE.T_CHECK_MANDATORY" max:"255" error:"SITE.FIELD_CODE.T_CHECK_INCORRECT"`
	Name	string	`db:"name" json:"name" min:"1" error:"SITE.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"SITE.FIELD_NAME.T_CHECK_INCORRECT"`
	City_name	string	`db:"city_name" json:"city_name"`
	City_geonameid	int	`db:"city_geonameid" json:"city_geonameid"`
	Geom	string	`db:"geom" json:"geom"`
	Geom_3d	string	`db:"geom_3d" json:"geom_3d"`
	Centroid	bool	`db:"centroid" json:"centroid" enum:"0,1" error:"SITE.FIELD_CENTROID.T_CHECK_MANDATORY"`
	Occupation	string	`db:"occupation" json:"occupation" enum:"not_documented,single,continuous,multiple" error:"SITE.FIELD_OCCUPATION.T_CHECK_INCORRECT"`
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Site_range struct {
	Id	int	`db:"id" json:"id"`
	Site_id	int	`db:"site_id" json:"site_id" xmltopsql:"ondelete:cascade"`	// Site.Id
	Start_date1	int	`db:"start_date1" json:"start_date1"`
	Start_date2	int	`db:"start_date2" json:"start_date2"`
	End_date1	int	`db:"end_date1" json:"end_date1"`
	End_date2	int	`db:"end_date2" json:"end_date2"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Site_range__charac struct {
	Id	int	`db:"id" json:"id"`
	Site_range_id	int	`db:"site_range_id" json:"site_range_id" xmltopsql:"ondelete:cascade"`	// Site_range.Id
	Charac_id	int	`db:"charac_id" json:"charac_id"`	// Charac.Id
	Exceptional	bool	`db:"exceptional" json:"exceptional"`
	Knowledge_type	string	`db:"knowledge_type" json:"knowledge_type" enum:"not_documented,literature,prospected_aerial,prospected_pedestrian,surveyed,dig" error:"DATABASE.FIELD_KNOWLEDGE_TYPE.T_CHECK_INCORRECT"`
}


type Site_range__charac_tr struct {
	Site_range__charac_id	int	`db:"site_range__charac_id" json:"site_range__charac_id" xmltopsql:"ondelete:cascade"`	// Site_range__charac.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Comment	string	`db:"comment" json:"comment"`
	Bibliography	string	`db:"bibliography" json:"bibliography"`
}


type Site_tr struct {
	Site_id	int	`db:"site_id" json:"site_id" xmltopsql:"ondelete:cascade"`	// Site.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Description	string	`db:"description" json:"description"`
}


type User struct {
	Id	int	`db:"id" json:"id"`
	Username	string	`db:"username" json:"username" min:"2" max:"32" error:"USER.FIELD_USERNAME.T_CHECK_INCORRECT"`
	Firstname	string	`db:"firstname" json:"firstname"`
	Lastname	string	`db:"lastname" json:"lastname" min:"1" error:"USER.FIELD_LASTNAME.T_CHECK_MANDATORY" max:"32" error:"USER.FIELD_LASTNAME.T_CHECK_INCORRECT"`
	Email	string	`db:"email" json:"email" email:"1" error:"USER.FIELD_EMAIL.T_CHECK_INCORRECT"`
	Password	string	`db:"password" json:"password" min:"6" max:"32" error:"USER.FIELD_PASSWORD.T_CHECK_INCORRECT"`
	Description	string	`db:"description" json:"description" min:"1" max:"2048" error:"USER.FIELD_DESCRIPTION.T_CHECK_MANDATORY"`
	Active	bool	`db:"active" json:"active"`
	First_lang_id	int	`db:"first_lang_id" json:"first_lang_id"`	// Lang.Id
	Second_lang_id	int	`db:"second_lang_id" json:"second_lang_id"`	// Lang.Id
	City_geonameid	int	`db:"city_geonameid" json:"city_geonameid"`	// City.Geonameid
	Photo_id	int	`db:"photo_id" json:"photo_id"`	// Photo.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type User__company struct {
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Company_id	int	`db:"company_id" json:"company_id"`	// Company.Id
}


type User__group struct {
	Group_id	int	`db:"group_id" json:"group_id"`	// Group.Id
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
}


type User_preferences struct {
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Key	string	`db:"key" json:"key"`
	Value	string	`db:"value" json:"value"`
}


type User_project struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
}


type Wms_map struct {
	Id	int	`db:"id" json:"id"`
	Creator_user_id	int	`db:"creator_user_id" json:"creator_user_id"`	// User.Id
	Url	sql.NullString	`db:"url" json:"url" min:"1" error:"WMS_MAP.FIELD_URL.T_CHECK_MANDATORY" max:"255" error:"WMS_MAP.FIELD_URL.T_CHECK_INCORRECT"`
	Source_creation_date	time.Time	`db:"source_creation_date" json:"source_creation_date"`
	Min_scale	int	`db:"min_scale" json:"min_scale"`
	Max_scale	int	`db:"max_scale" json:"max_scale"`
	Start_date1	int	`db:"start_date1" json:"start_date1"`
	Start_date2	int	`db:"start_date2" json:"start_date2"`
	End_date1	int	`db:"end_date1" json:"end_date1"`
	End_date2	int	`db:"end_date2" json:"end_date2"`
	Published	bool	`db:"published" json:"published"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Wms_map_tr struct {
	Wms_map_id	int	`db:"wms_map_id" json:"wms_map_id"`	// Wms_map.Id
	Lang_id	int	`db:"lang_id" json:"lang_id"`	// Lang.Id
	Name	string	`db:"name" json:"name" min:"1" error:"WMS_MAP.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"WMS_MAP_TR.FIELD_NAME.T_CHECK_INCORRECT"`
	Attribution	string	`db:"attribution" json:"attribution"`
	Copyright	string	`db:"copyright" json:"copyright"`
}


const User_InsertStr = "\"username\", \"firstname\", \"lastname\", \"email\", \"password\", \"description\", \"active\", \"first_lang_id\", \"second_lang_id\", \"city_geonameid\", \"photo_id\", \"created_at\", \"updated_at\""
const User_InsertValuesStr = ":username, :firstname, :lastname, :email, :password, :description, :active, :first_lang_id, :second_lang_id, :city_geonameid, :photo_id, now(), now()"
const User_UpdateStr = "\"username\" = :username, \"firstname\" = :firstname, \"lastname\" = :lastname, \"email\" = :email, \"password\" = :password, \"description\" = :description, \"active\" = :active, \"first_lang_id\" = :first_lang_id, \"second_lang_id\" = :second_lang_id, \"city_geonameid\" = :city_geonameid, \"photo_id\" = :photo_id, \"updated_at\" = now()"
const Lang_InsertStr = "\"iso_code\", \"active\""
const Lang_InsertValuesStr = ":iso_code, :active"
const Lang_UpdateStr = "\"iso_code\" = :iso_code, \"active\" = :active"
const User__company_InsertStr = ""
const User__company_InsertValuesStr = ""
const User__company_UpdateStr = ""
const Company_InsertStr = "\"name\", \"city_geonameid\""
const Company_InsertValuesStr = ":name, :city_geonameid"
const Company_UpdateStr = "\"name\" = :name, \"city_geonameid\" = :city_geonameid"
const Project_InsertStr = "\"name\", \"user_id\", \"geographical_zone_id\", \"created_at\", \"updated_at\""
const Project_InsertValuesStr = ":name, :user_id, :geographical_zone_id, now(), now()"
const Project_UpdateStr = "\"name\" = :name, \"user_id\" = :user_id, \"geographical_zone_id\" = :geographical_zone_id, \"updated_at\" = now()"
const Geographical_zone_InsertStr = "\"name\", \"geom\""
const Geographical_zone_InsertValuesStr = ":name, :geom"
const Geographical_zone_UpdateStr = "\"name\" = :name, \"geom\" = :geom"
const Chronology_InsertStr = "\"parent_id\", \"start_date\", \"end_date\", \"color\", \"created_at\", \"updated_at\""
const Chronology_InsertValuesStr = ":parent_id, :start_date, :end_date, :color, now(), now()"
const Chronology_UpdateStr = "\"parent_id\" = :parent_id, \"start_date\" = :start_date, \"end_date\" = :end_date, \"color\" = :color, \"updated_at\" = now()"
const Chronology_tr_InsertStr = "\"name\", \"description\""
const Chronology_tr_InsertValuesStr = ":name, :description"
const Chronology_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Project__chronology_InsertStr = "\"id_group\""
const Project__chronology_InsertValuesStr = ":id_group"
const Project__chronology_UpdateStr = "\"id_group\" = :id_group"
const Database_InsertStr = "\"name\", \"scale_resolution\", \"geographical_extent\", \"type\", \"owner\", \"editor\", \"contributor\", \"default_language\", \"state\", \"license_id\", \"published\", \"soft_deleted\", \"declared_creation_date\", \"created_at\", \"updated_at\""
const Database_InsertValuesStr = ":name, :scale_resolution, :geographical_extent, :type, :owner, :editor, :contributor, :default_language, :state, :license_id, :published, :soft_deleted, :declared_creation_date, now(), now()"
const Database_UpdateStr = "\"name\" = :name, \"scale_resolution\" = :scale_resolution, \"geographical_extent\" = :geographical_extent, \"type\" = :type, \"owner\" = :owner, \"editor\" = :editor, \"contributor\" = :contributor, \"default_language\" = :default_language, \"state\" = :state, \"license_id\" = :license_id, \"published\" = :published, \"soft_deleted\" = :soft_deleted, \"declared_creation_date\" = :declared_creation_date, \"updated_at\" = now()"
const Site_InsertStr = "\"code\", \"name\", \"city_name\", \"city_geonameid\", \"geom\", \"geom_3d\", \"centroid\", \"occupation\", \"database_id\", \"created_at\", \"updated_at\""
const Site_InsertValuesStr = ":code, :name, :city_name, :city_geonameid, :geom, :geom_3d, :centroid, :occupation, :database_id, now(), now()"
const Site_UpdateStr = "\"code\" = :code, \"name\" = :name, \"city_name\" = :city_name, \"city_geonameid\" = :city_geonameid, \"geom\" = :geom, \"geom_3d\" = :geom_3d, \"centroid\" = :centroid, \"occupation\" = :occupation, \"database_id\" = :database_id, \"updated_at\" = now()"
const Database_tr_InsertStr = "\"description\", \"geographical_limit\", \"bibliography\", \"context_description\", \"source_description\", \"source_relation\", \"copyright\", \"subject\""
const Database_tr_InsertValuesStr = ":description, :geographical_limit, :bibliography, :context_description, :source_description, :source_relation, :copyright, :subject"
const Database_tr_UpdateStr = "\"description\" = :description, \"geographical_limit\" = :geographical_limit, \"bibliography\" = :bibliography, \"context_description\" = :context_description, \"source_description\" = :source_description, \"source_relation\" = :source_relation, \"copyright\" = :copyright, \"subject\" = :subject"
const City_InsertStr = "\"country_geonameid\", \"geom\", \"geom_centroid\", \"created_at\", \"updated_at\""
const City_InsertValuesStr = ":country_geonameid, :geom, :geom_centroid, now(), now()"
const City_UpdateStr = "\"country_geonameid\" = :country_geonameid, \"geom\" = :geom, \"geom_centroid\" = :geom_centroid, \"updated_at\" = now()"
const Site_tr_InsertStr = "\"description\""
const Site_tr_InsertValuesStr = ":description"
const Site_tr_UpdateStr = "\"description\" = :description"
const Site_range_InsertStr = "\"site_id\", \"start_date1\", \"start_date2\", \"end_date1\", \"end_date2\", \"created_at\", \"updated_at\""
const Site_range_InsertValuesStr = ":site_id, :start_date1, :start_date2, :end_date1, :end_date2, now(), now()"
const Site_range_UpdateStr = "\"site_id\" = :site_id, \"start_date1\" = :start_date1, \"start_date2\" = :start_date2, \"end_date1\" = :end_date1, \"end_date2\" = :end_date2, \"updated_at\" = now()"
const Site_range__charac_tr_InsertStr = "\"comment\", \"bibliography\""
const Site_range__charac_tr_InsertValuesStr = ":comment, :bibliography"
const Site_range__charac_tr_UpdateStr = "\"comment\" = :comment, \"bibliography\" = :bibliography"
const City_tr_InsertStr = "\"name\", \"name_ascii\""
const City_tr_InsertValuesStr = ":name, :name_ascii"
const City_tr_UpdateStr = "\"name\" = :name, \"name_ascii\" = :name_ascii"
const Charac_tr_InsertStr = "\"name\", \"description\""
const Charac_tr_InsertValuesStr = ":name, :description"
const Charac_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Charac_InsertStr = "\"parent_id\", \"order\", \"author_user_id\", \"created_at\", \"updated_at\""
const Charac_InsertValuesStr = ":parent_id, :order, :author_user_id, now(), now()"
const Charac_UpdateStr = "\"parent_id\" = :parent_id, \"order\" = :order, \"author_user_id\" = :author_user_id, \"updated_at\" = now()"
const Wms_map_InsertStr = "\"creator_user_id\", \"url\", \"source_creation_date\", \"min_scale\", \"max_scale\", \"start_date1\", \"start_date2\", \"end_date1\", \"end_date2\", \"published\", \"created_at\", \"updated_at\""
const Wms_map_InsertValuesStr = ":creator_user_id, :url, :source_creation_date, :min_scale, :max_scale, :start_date1, :start_date2, :end_date1, :end_date2, :published, now(), now()"
const Wms_map_UpdateStr = "\"creator_user_id\" = :creator_user_id, \"url\" = :url, \"source_creation_date\" = :source_creation_date, \"min_scale\" = :min_scale, \"max_scale\" = :max_scale, \"start_date1\" = :start_date1, \"start_date2\" = :start_date2, \"end_date1\" = :end_date1, \"end_date2\" = :end_date2, \"published\" = :published, \"updated_at\" = now()"
const Wms_map_tr_InsertStr = "\"name\", \"attribution\", \"copyright\""
const Wms_map_tr_InsertValuesStr = ":name, :attribution, :copyright"
const Wms_map_tr_UpdateStr = "\"name\" = :name, \"attribution\" = :attribution, \"copyright\" = :copyright"
const Shapefile_InsertStr = "\"creator_user_id\", \"source_creation_date\", \"filename\", \"geom\", \"min_scale\", \"max_scale\", \"start_date1\", \"start_date2\", \"end_date1\", \"end_date2\", \"published\", \"license_name\", \"odbl_license\", \"created_at\", \"updated_at\""
const Shapefile_InsertValuesStr = ":creator_user_id, :source_creation_date, :filename, :geom, :min_scale, :max_scale, :start_date1, :start_date2, :end_date1, :end_date2, :published, :license_name, :odbl_license, now(), now()"
const Shapefile_UpdateStr = "\"creator_user_id\" = :creator_user_id, \"source_creation_date\" = :source_creation_date, \"filename\" = :filename, \"geom\" = :geom, \"min_scale\" = :min_scale, \"max_scale\" = :max_scale, \"start_date1\" = :start_date1, \"start_date2\" = :start_date2, \"end_date1\" = :end_date1, \"end_date2\" = :end_date2, \"published\" = :published, \"license_name\" = :license_name, \"odbl_license\" = :odbl_license, \"updated_at\" = now()"
const Shapefile_tr_InsertStr = "\"name\", \"description\", \"attribution\", \"bibiography\""
const Shapefile_tr_InsertValuesStr = ":name, :description, :attribution, :bibiography"
const Shapefile_tr_UpdateStr = "\"name\" = :name, \"description\" = :description, \"attribution\" = :attribution, \"bibiography\" = :bibiography"
const Project__databases_InsertStr = ""
const Project__databases_InsertValuesStr = ""
const Project__databases_UpdateStr = ""
const Project__wms_map_InsertStr = ""
const Project__wms_map_InsertValuesStr = ""
const Project__wms_map_UpdateStr = ""
const Project__shapefile_InsertStr = ""
const Project__shapefile_InsertValuesStr = ""
const Project__shapefile_UpdateStr = ""
const Site_range__charac_InsertStr = "\"site_range_id\", \"charac_id\", \"exceptional\", \"knowledge_type\""
const Site_range__charac_InsertValuesStr = ":site_range_id, :charac_id, :exceptional, :knowledge_type"
const Site_range__charac_UpdateStr = "\"site_range_id\" = :site_range_id, \"charac_id\" = :charac_id, \"exceptional\" = :exceptional, \"knowledge_type\" = :knowledge_type"
const Project__charac_InsertStr = ""
const Project__charac_InsertValuesStr = ""
const Project__charac_UpdateStr = ""
const Project__characs_set_InsertStr = "\"project_id\", \"copied_from_project__charac_set_id\""
const Project__characs_set_InsertValuesStr = ":project_id, :copied_from_project__charac_set_id"
const Project__characs_set_UpdateStr = "\"project_id\" = :project_id, \"copied_from_project__charac_set_id\" = :copied_from_project__charac_set_id"
const Project__characs_set_tr_InsertStr = "\"name\", \"description\""
const Project__characs_set_tr_InsertValuesStr = ":name, :description"
const Project__characs_set_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const User_project_InsertStr = "\"start_date\", \"end_date\""
const User_project_InsertValuesStr = ":start_date, :end_date"
const User_project_UpdateStr = "\"start_date\" = :start_date, \"end_date\" = :end_date"
const Global_project_InsertStr = ""
const Global_project_InsertValuesStr = ""
const Global_project_UpdateStr = ""
const Project__chronology_tr_InsertStr = "\"name\", \"description\""
const Project__chronology_tr_InsertValuesStr = ":name, :description"
const Project__chronology_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Group_InsertStr = "\"type\", \"created_at\", \"updated_at\""
const Group_InsertValuesStr = ":type, now(), now()"
const Group_UpdateStr = "\"type\" = :type, \"updated_at\" = now()"
const Permission_InsertStr = "\"name\", \"created_at\", \"updated_at\""
const Permission_InsertValuesStr = ":name, now(), now()"
const Permission_UpdateStr = "\"name\" = :name, \"updated_at\" = now()"
const Group_tr_InsertStr = "\"name\", \"description\""
const Group_tr_InsertValuesStr = ":name, :description"
const Group_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Permission_tr_InsertStr = "\"name\", \"description\""
const Permission_tr_InsertValuesStr = ":name, :description"
const Permission_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const User__group_InsertStr = ""
const User__group_InsertValuesStr = ""
const User__group_UpdateStr = ""
const Group__permission_InsertStr = "\"created_at\", \"updated_at\""
const Group__permission_InsertValuesStr = "now(), now()"
const Group__permission_UpdateStr = "\"updated_at\" = now()"
const Country_InsertStr = "\"iso_code\", \"geom\", \"created_at\", \"updated_at\""
const Country_InsertValuesStr = ":iso_code, :geom, now(), now()"
const Country_UpdateStr = "\"iso_code\" = :iso_code, \"geom\" = :geom, \"updated_at\" = now()"
const User_preferences_InsertStr = "\"value\""
const User_preferences_InsertValuesStr = ":value"
const User_preferences_UpdateStr = "\"value\" = :value"
const Country_tr_InsertStr = "\"name\", \"name_ascii\""
const Country_tr_InsertValuesStr = ":name, :name_ascii"
const Country_tr_UpdateStr = "\"name\" = :name, \"name_ascii\" = :name_ascii"
const Database__authors_InsertStr = ""
const Database__authors_InsertValuesStr = ""
const Database__authors_UpdateStr = ""
const Import_InsertStr = "\"database_id\", \"user_id\", \"filename\", \"number_of_lines\", \"created_at\""
const Import_InsertValuesStr = ":database_id, :user_id, :filename, :number_of_lines, now()"
const Import_UpdateStr = "\"database_id\" = :database_id, \"user_id\" = :user_id, \"filename\" = :filename, \"number_of_lines\" = :number_of_lines"
const License_InsertStr = "\"name\", \"url\""
const License_InsertValuesStr = ":name, :url"
const License_UpdateStr = "\"name\" = :name, \"url\" = :url"
const Database__country_InsertStr = ""
const Database__country_InsertValuesStr = ""
const Database__country_UpdateStr = ""
const Database_handle_InsertStr = "\"database_id\", \"import_id\", \"identifier\", \"url\", \"declared_creation_date\", \"created_at\""
const Database_handle_InsertValuesStr = ":database_id, :import_id, :identifier, :url, :declared_creation_date, now()"
const Database_handle_UpdateStr = "\"database_id\" = :database_id, \"import_id\" = :import_id, \"identifier\" = :identifier, \"url\" = :url, \"declared_creation_date\" = :declared_creation_date"
const Shapefile_authors_InsertStr = ""
const Shapefile_authors_InsertValuesStr = ""
const Shapefile_authors_UpdateStr = ""
const Continent_tr_InsertStr = "\"name\", \"name_ascii\""
const Continent_tr_InsertValuesStr = ":name, :name_ascii"
const Continent_tr_UpdateStr = "\"name\" = :name, \"name_ascii\" = :name_ascii"
const Continent_InsertStr = "\"iso_code\", \"geom\", \"created_at\", \"updated_at\""
const Continent_InsertValuesStr = ":iso_code, :geom, now(), now()"
const Continent_UpdateStr = "\"iso_code\" = :iso_code, \"geom\" = :geom, \"updated_at\" = now()"
const Database__continent_InsertStr = ""
const Database__continent_InsertValuesStr = ""
const Database__continent_UpdateStr = ""
const Lang_tr_InsertStr = "\"name\""
const Lang_tr_InsertValuesStr = ":name"
const Lang_tr_UpdateStr = "\"name\" = :name"
const Session_InsertStr = "\"value\""
const Session_InsertValuesStr = ":value"
const Session_UpdateStr = "\"value\" = :value"
const Photo_InsertStr = "\"photo\""
const Photo_InsertValuesStr = ":photo"
const Photo_UpdateStr = "\"photo\" = :photo"
const Database_context_InsertStr = "\"database_id\", \"context\""
const Database_context_InsertValuesStr = ":database_id, :context"
const Database_context_UpdateStr = "\"database_id\" = :database_id, \"context\" = :context"
