package model

import (
	"time"
	"database/sql"
)

type Charac struct {
	Id	int	`db:"id" json:"id" xmltopsql:"ondelete:cascade"`
	Parent_id	int	`db:"parent_id" json:"parent_id" xmltopsql:"ondelete:cascade"`	// Charac.Id
	Order	int	`db:"order" json:"order"`
	Author_user_id	int	`db:"author_user_id" json:"author_user_id"`	// User.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Charac_root struct {
	Root_charac_id	int	`db:"root_charac_id" json:"root_charac_id" xmltopsql:"ondelete:cascade"`	// Charac.Id
	Admin_group_id	int	`db:"admin_group_id" json:"admin_group_id"`	// Group.Id
	Cached_langs	string	`db:"cached_langs" json:"cached_langs"`
}


type Charac_tr struct {
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Charac_id	int	`db:"charac_id" json:"charac_id" xmltopsql:"ondelete:cascade"`	// Charac.Id
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Chronology struct {
	Id	int	`db:"id" json:"id" xmltopsql:"ondelete:cascade"`
	Parent_id	int	`db:"parent_id" json:"parent_id" xmltopsql:"ondelete:cascade"`	// Chronology.Id
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
	Color	string	`db:"color" json:"color"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Chronology_root struct {
	Root_chronology_id	int	`db:"root_chronology_id" json:"root_chronology_id" xmltopsql:"ondelete:cascade"`	// Chronology.Id
	Admin_group_id	int	`db:"admin_group_id" json:"admin_group_id"`	// Group.Id
	Author_user_id	int	`db:"author_user_id" json:"author_user_id"`	// User.Id
	Credits	string	`db:"credits" json:"credits"`
	Active	bool	`db:"active" json:"active"`
	Geom	string	`db:"geom" json:"geom"`
	Cached_langs	string	`db:"cached_langs" json:"cached_langs"`
}


type Chronology_tr struct {
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Chronology_id	int	`db:"chronology_id" json:"chronology_id" xmltopsql:"ondelete:cascade"`	// Chronology.Id
	Name	string	`db:"name" json:"name" xmltopsql:"ondelete:cascade"`
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
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
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
	Lang_isocode	sql.NullString	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
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
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
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
	Default_language	string	`db:"default_language" json:"default_language"`	// Lang.Isocode
	State	string	`db:"state" json:"state" enum:"undefined,in-progress,finished" error:"DATABASE.FIELD_STATE.T_CHECK_INCORRECT"`
	License_id	int	`db:"license_id" json:"license_id"`	// License.Id
	Published	bool	`db:"published" json:"published"`
	Soft_deleted	bool	`db:"soft_deleted" json:"soft_deleted"`
	Geographical_extent_geom	string	`db:"geographical_extent_geom" json:"geographical_extent_geom"`
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
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
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Description	string	`db:"description" json:"description"`
	Geographical_limit	string	`db:"geographical_limit" json:"geographical_limit"`
	Bibliography	string	`db:"bibliography" json:"bibliography"`
	Context_description	string	`db:"context_description" json:"context_description"`
	Source_description	string	`db:"source_description" json:"source_description"`
	Source_relation	string	`db:"source_relation" json:"source_relation"`
	Copyright	string	`db:"copyright" json:"copyright"`
	Subject	string	`db:"subject" json:"subject"`
}


type Group struct {
	Id	int	`db:"id" json:"id"`
	Type	string	`db:"type" json:"type"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Group__permission struct {
	Group_id	int	`db:"group_id" json:"group_id" xmltopsql:"ondelete:cascade"`	// Group.Id
	Permission_id	int	`db:"permission_id" json:"permission_id"`	// Permission.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Group_tr struct {
	Group_id	int	`db:"group_id" json:"group_id" xmltopsql:"ondelete:cascade"`	// Group.Id
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Import struct {
	Id	int	`db:"id" json:"id"`
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Md5sum	string	`db:"md5sum" json:"md5sum"`
	Filename	string	`db:"filename" json:"filename"`
	Number_of_lines	int	`db:"number_of_lines" json:"number_of_lines"`
	Number_of_sites	int	`db:"number_of_sites" json:"number_of_sites"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
}


type Lang struct {
	Isocode	string	`db:"isocode" json:"isocode"`
	Active	bool	`db:"active" json:"active"`
}


type Lang_tr struct {
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Lang_isocode_tr	string	`db:"lang_isocode_tr" json:"lang_isocode_tr"`	// Lang.Isocode
	Name	string	`db:"name" json:"name"`
}


type License struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name"`
	Url	string	`db:"url" json:"url"`
}


type Map_layer struct {
	Id	int	`db:"id" json:"id"`	// Map_layer__authors.Map_layer_id
	Creator_user_id	int	`db:"creator_user_id" json:"creator_user_id"`	// User.Id
	Type	string	`db:"type" json:"type"`
	Url	string	`db:"url" json:"url" min:"1" error:"WMS_MAP.FIELD_URL.T_CHECK_MANDATORY" max:"255" error:"WMS_MAP.FIELD_URL.T_CHECK_INCORRECT"`
	Identifier	string	`db:"identifier" json:"identifier"`
	Min_scale	int	`db:"min_scale" json:"min_scale"`
	Max_scale	int	`db:"max_scale" json:"max_scale"`
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
	Image_format	string	`db:"image_format" json:"image_format"`
	Geographical_extent_geom	string	`db:"geographical_extent_geom" json:"geographical_extent_geom"`
	Published	bool	`db:"published" json:"published"`
	License	string	`db:"license" json:"license"`
	License_id	int	`db:"license_id" json:"license_id"`	// License.Id
	Max_usage_date	time.Time	`db:"max_usage_date" json:"max_usage_date"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Map_layer__authors struct {
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Map_layer_id	int	`db:"map_layer_id" json:"map_layer_id"`
}


type Map_layer_tr struct {
	Map_layer_id	int	`db:"map_layer_id" json:"map_layer_id"`	// Map_layer.Id
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Name	string	`db:"name" json:"name" min:"1" error:"WMS_MAP.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"WMS_MAP_TR.FIELD_NAME.T_CHECK_INCORRECT"`
	Attribution	string	`db:"attribution" json:"attribution"`
	Copyright	string	`db:"copyright" json:"copyright"`
	Description	string	`db:"description" json:"description"`
}


type Permission struct {
	Id	int	`db:"id" json:"id"`
	Name	string	`db:"name" json:"name"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Permission_tr struct {
	Permission_id	int	`db:"permission_id" json:"permission_id"`	// Permission.Id
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Name	string	`db:"name" json:"name"`
	Description	string	`db:"description" json:"description"`
}


type Photo struct {
	Id	int	`db:"id" json:"id"`
	Photo	string	`db:"photo" json:"photo"`
}


type Project struct {
	Id	int	`db:"id" json:"id"`	// Project__charac.Project_id
	Name	string	`db:"name" json:"name" min:"1" error:"PROJECT.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"PROJECT.FIELD_NAME.T_CHECK_INCORRECT"`
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
	Geom	string	`db:"geom" json:"geom"`
}


type Project__charac struct {
	Project_id	int	`db:"project_id" json:"project_id" xmltopsql:"ondelete:cascade"`
	Root_charac_id	int	`db:"root_charac_id" json:"root_charac_id"`	// Charac.Id
}


type Project__chronology struct {
	Project_id	int	`db:"project_id" json:"project_id" xmltopsql:"ondelete:cascade"`	// Project.Id
	Root_chronology_id	int	`db:"root_chronology_id" json:"root_chronology_id"`	// Chronology.Id
}


type Project__database struct {
	Project_id	int	`db:"project_id" json:"project_id" xmltopsql:"ondelete:cascade"`	// Project.Id
	Database_id	int	`db:"database_id" json:"database_id"`	// Database.Id
}


type Project__map_layer struct {
	Project_id	int	`db:"project_id" json:"project_id" xmltopsql:"ondelete:cascade"`	// Project.Id
	Map_layer_id	int	`db:"map_layer_id" json:"map_layer_id"`	// Map_layer.Id
}


type Project__shapefile struct {
	Project_id	int	`db:"project_id" json:"project_id"`	// Project.Id
	Shapefile_id	int	`db:"shapefile_id" json:"shapefile_id"`	// Shapefile.Id
}


type Project_hidden_characs struct {
	Project_id	int	`db:"project_id" json:"project_id" xmltopsql:"ondelete:cascade"`	// Project.Id
	Charac_id	int	`db:"charac_id" json:"charac_id" xmltopsql:"ondelete:cascade"`	// Charac.Id
}


type Saved_query struct {
	Project_id	int	`db:"project_id" json:"project_id" xmltopsql:"ondelete:cascade"`	// Project.Id
	Name	string	`db:"name" json:"name" min:"1"`
	Params	string	`db:"params" json:"params"`
}


type Session struct {
	Token	string	`db:"token" json:"token"`
	Value	string	`db:"value" json:"value"`
}


type Shapefile struct {
	Id	int	`db:"id" json:"id"`
	Creator_user_id	int	`db:"creator_user_id" json:"creator_user_id"`	// User.Id
	Filename	string	`db:"filename" json:"filename"`
	Md5sum	string	`db:"md5sum" json:"md5sum"`
	Geojson	string	`db:"geojson" json:"geojson"`
	Geojson_with_data	string	`db:"geojson_with_data" json:"geojson_with_data"`
	Start_date	int	`db:"start_date" json:"start_date"`
	End_date	int	`db:"end_date" json:"end_date"`
	Geographical_extent_geom	string	`db:"geographical_extent_geom" json:"geographical_extent_geom"`
	Published	bool	`db:"published" json:"published" enum:"0,1" error:"SHAPEFILE.FIELD_PUBLISHED.T_CHECK_INCORRECT"`
	License	string	`db:"license" json:"license"`
	License_id	int	`db:"license_id" json:"license_id"`	// License.Id
	Declared_creation_date	time.Time	`db:"declared_creation_date" json:"declared_creation_date"`
	Created_at	time.Time	`db:"created_at" json:"created_at"`
	Updated_at	time.Time	`db:"updated_at" json:"updated_at"`
}


type Shapefile__authors struct {
	User_id	int	`db:"user_id" json:"user_id"`	// User.Id
	Shapefile_id	int	`db:"shapefile_id" json:"shapefile_id"`	// Shapefile.Id
}


type Shapefile_tr struct {
	Shapefile_id	int	`db:"shapefile_id" json:"shapefile_id"`	// Shapefile.Id
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Name	string	`db:"name" json:"name" min:"1" error:"SHAPEFILE.FIELD_NAME.T_CHECK_MANDATORY" max:"255" error:"SHAPEFILE_TR.FIELD_NAME.T_CHECK_INCORRECT"`
	Attribution	string	`db:"attribution" json:"attribution"`
	Copyright	string	`db:"copyright" json:"copyright"`
	Description	string	`db:"description" json:"description"`
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
	Altitude	int	`db:"altitude" json:"altitude"`
	Start_date1	int	`db:"start_date1" json:"start_date1"`
	Start_date2	int	`db:"start_date2" json:"start_date2"`
	End_date1	int	`db:"end_date1" json:"end_date1"`
	End_date2	int	`db:"end_date2" json:"end_date2"`
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
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
	Comment	string	`db:"comment" json:"comment"`
	Bibliography	string	`db:"bibliography" json:"bibliography"`
}


type Site_tr struct {
	Site_id	int	`db:"site_id" json:"site_id" xmltopsql:"ondelete:cascade"`	// Site.Id
	Lang_isocode	string	`db:"lang_isocode" json:"lang_isocode"`	// Lang.Isocode
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
	First_lang_isocode	string	`db:"first_lang_isocode" json:"first_lang_isocode"`	// Lang.Isocode
	Second_lang_isocode	string	`db:"second_lang_isocode" json:"second_lang_isocode"`	// Lang.Isocode
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


const User_InsertStr = "\"username\", \"firstname\", \"lastname\", \"email\", \"password\", \"description\", \"active\", \"first_lang_isocode\", \"second_lang_isocode\", \"city_geonameid\", \"photo_id\", \"created_at\", \"updated_at\""
const User_InsertValuesStr = ":username, :firstname, :lastname, :email, :password, :description, :active, :first_lang_isocode, :second_lang_isocode, :city_geonameid, :photo_id, now(), now()"
const User_UpdateStr = "\"username\" = :username, \"firstname\" = :firstname, \"lastname\" = :lastname, \"email\" = :email, \"password\" = :password, \"description\" = :description, \"active\" = :active, \"first_lang_isocode\" = :first_lang_isocode, \"second_lang_isocode\" = :second_lang_isocode, \"city_geonameid\" = :city_geonameid, \"photo_id\" = :photo_id, \"updated_at\" = now()"
const Lang_InsertStr = "\"active\""
const Lang_InsertValuesStr = ":active"
const Lang_UpdateStr = "\"active\" = :active"
const User__company_InsertStr = ""
const User__company_InsertValuesStr = ""
const User__company_UpdateStr = ""
const Company_InsertStr = "\"name\", \"city_geonameid\""
const Company_InsertValuesStr = ":name, :city_geonameid"
const Company_UpdateStr = "\"name\" = :name, \"city_geonameid\" = :city_geonameid"
const Project_InsertStr = "\"name\", \"user_id\", \"created_at\", \"updated_at\", \"start_date\", \"end_date\", \"geom\""
const Project_InsertValuesStr = ":name, :user_id, now(), now(), :start_date, :end_date, :geom"
const Project_UpdateStr = "\"name\" = :name, \"user_id\" = :user_id, \"updated_at\" = now(), \"start_date\" = :start_date, \"end_date\" = :end_date, \"geom\" = :geom"
const Chronology_InsertStr = "\"parent_id\", \"start_date\", \"end_date\", \"color\", \"created_at\", \"updated_at\""
const Chronology_InsertValuesStr = ":parent_id, :start_date, :end_date, :color, now(), now()"
const Chronology_UpdateStr = "\"parent_id\" = :parent_id, \"start_date\" = :start_date, \"end_date\" = :end_date, \"color\" = :color, \"updated_at\" = now()"
const Chronology_tr_InsertStr = "\"name\", \"description\""
const Chronology_tr_InsertValuesStr = ":name, :description"
const Chronology_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Project__chronology_InsertStr = ""
const Project__chronology_InsertValuesStr = ""
const Project__chronology_UpdateStr = ""
const Database_InsertStr = "\"name\", \"scale_resolution\", \"geographical_extent\", \"type\", \"owner\", \"editor\", \"contributor\", \"default_language\", \"state\", \"license_id\", \"published\", \"soft_deleted\", \"geographical_extent_geom\", \"start_date\", \"end_date\", \"declared_creation_date\", \"created_at\", \"updated_at\""
const Database_InsertValuesStr = ":name, :scale_resolution, :geographical_extent, :type, :owner, :editor, :contributor, :default_language, :state, :license_id, :published, :soft_deleted, :geographical_extent_geom, :start_date, :end_date, :declared_creation_date, now(), now()"
const Database_UpdateStr = "\"name\" = :name, \"scale_resolution\" = :scale_resolution, \"geographical_extent\" = :geographical_extent, \"type\" = :type, \"owner\" = :owner, \"editor\" = :editor, \"contributor\" = :contributor, \"default_language\" = :default_language, \"state\" = :state, \"license_id\" = :license_id, \"published\" = :published, \"soft_deleted\" = :soft_deleted, \"geographical_extent_geom\" = :geographical_extent_geom, \"start_date\" = :start_date, \"end_date\" = :end_date, \"declared_creation_date\" = :declared_creation_date, \"updated_at\" = now()"
const Site_InsertStr = "\"code\", \"name\", \"city_name\", \"city_geonameid\", \"geom\", \"geom_3d\", \"centroid\", \"occupation\", \"database_id\", \"created_at\", \"updated_at\", \"altitude\", \"start_date1\", \"start_date2\", \"end_date1\", \"end_date2\""
const Site_InsertValuesStr = ":code, :name, :city_name, :city_geonameid, :geom, :geom_3d, :centroid, :occupation, :database_id, now(), now(), :altitude, :start_date1, :start_date2, :end_date1, :end_date2"
const Site_UpdateStr = "\"code\" = :code, \"name\" = :name, \"city_name\" = :city_name, \"city_geonameid\" = :city_geonameid, \"geom\" = :geom, \"geom_3d\" = :geom_3d, \"centroid\" = :centroid, \"occupation\" = :occupation, \"database_id\" = :database_id, \"updated_at\" = now(), \"altitude\" = :altitude, \"start_date1\" = :start_date1, \"start_date2\" = :start_date2, \"end_date1\" = :end_date1, \"end_date2\" = :end_date2"
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
const Map_layer_InsertStr = "\"creator_user_id\", \"type\", \"url\", \"identifier\", \"min_scale\", \"max_scale\", \"start_date\", \"end_date\", \"image_format\", \"geographical_extent_geom\", \"published\", \"license\", \"license_id\", \"max_usage_date\", \"created_at\", \"updated_at\""
const Map_layer_InsertValuesStr = ":creator_user_id, :type, :url, :identifier, :min_scale, :max_scale, :start_date, :end_date, :image_format, :geographical_extent_geom, :published, :license, :license_id, :max_usage_date, now(), now()"
const Map_layer_UpdateStr = "\"creator_user_id\" = :creator_user_id, \"type\" = :type, \"url\" = :url, \"identifier\" = :identifier, \"min_scale\" = :min_scale, \"max_scale\" = :max_scale, \"start_date\" = :start_date, \"end_date\" = :end_date, \"image_format\" = :image_format, \"geographical_extent_geom\" = :geographical_extent_geom, \"published\" = :published, \"license\" = :license, \"license_id\" = :license_id, \"max_usage_date\" = :max_usage_date, \"updated_at\" = now()"
const Map_layer_tr_InsertStr = "\"name\", \"attribution\", \"copyright\", \"description\""
const Map_layer_tr_InsertValuesStr = ":name, :attribution, :copyright, :description"
const Map_layer_tr_UpdateStr = "\"name\" = :name, \"attribution\" = :attribution, \"copyright\" = :copyright, \"description\" = :description"
const Shapefile_InsertStr = "\"creator_user_id\", \"filename\", \"md5sum\", \"geojson\", \"geojson_with_data\", \"start_date\", \"end_date\", \"geographical_extent_geom\", \"published\", \"license\", \"license_id\", \"declared_creation_date\", \"created_at\", \"updated_at\""
const Shapefile_InsertValuesStr = ":creator_user_id, :filename, :md5sum, :geojson, :geojson_with_data, :start_date, :end_date, :geographical_extent_geom, :published, :license, :license_id, :declared_creation_date, now(), now()"
const Shapefile_UpdateStr = "\"creator_user_id\" = :creator_user_id, \"filename\" = :filename, \"md5sum\" = :md5sum, \"geojson\" = :geojson, \"geojson_with_data\" = :geojson_with_data, \"start_date\" = :start_date, \"end_date\" = :end_date, \"geographical_extent_geom\" = :geographical_extent_geom, \"published\" = :published, \"license\" = :license, \"license_id\" = :license_id, \"declared_creation_date\" = :declared_creation_date, \"updated_at\" = now()"
const Shapefile_tr_InsertStr = "\"name\", \"attribution\", \"copyright\", \"description\""
const Shapefile_tr_InsertValuesStr = ":name, :attribution, :copyright, :description"
const Shapefile_tr_UpdateStr = "\"name\" = :name, \"attribution\" = :attribution, \"copyright\" = :copyright, \"description\" = :description"
const Project__database_InsertStr = ""
const Project__database_InsertValuesStr = ""
const Project__database_UpdateStr = ""
const Project__map_layer_InsertStr = ""
const Project__map_layer_InsertValuesStr = ""
const Project__map_layer_UpdateStr = ""
const Project__shapefile_InsertStr = ""
const Project__shapefile_InsertValuesStr = ""
const Project__shapefile_UpdateStr = ""
const Site_range__charac_InsertStr = "\"site_range_id\", \"charac_id\", \"exceptional\", \"knowledge_type\""
const Site_range__charac_InsertValuesStr = ":site_range_id, :charac_id, :exceptional, :knowledge_type"
const Site_range__charac_UpdateStr = "\"site_range_id\" = :site_range_id, \"charac_id\" = :charac_id, \"exceptional\" = :exceptional, \"knowledge_type\" = :knowledge_type"
const Project_hidden_characs_InsertStr = ""
const Project_hidden_characs_InsertValuesStr = ""
const Project_hidden_characs_UpdateStr = ""
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
const Import_InsertStr = "\"database_id\", \"user_id\", \"md5sum\", \"filename\", \"number_of_lines\", \"number_of_sites\", \"created_at\""
const Import_InsertValuesStr = ":database_id, :user_id, :md5sum, :filename, :number_of_lines, :number_of_sites, now()"
const Import_UpdateStr = "\"database_id\" = :database_id, \"user_id\" = :user_id, \"md5sum\" = :md5sum, \"filename\" = :filename, \"number_of_lines\" = :number_of_lines, \"number_of_sites\" = :number_of_sites"
const License_InsertStr = "\"name\", \"url\""
const License_InsertValuesStr = ":name, :url"
const License_UpdateStr = "\"name\" = :name, \"url\" = :url"
const Database__country_InsertStr = ""
const Database__country_InsertValuesStr = ""
const Database__country_UpdateStr = ""
const Database_handle_InsertStr = "\"database_id\", \"import_id\", \"identifier\", \"url\", \"declared_creation_date\", \"created_at\""
const Database_handle_InsertValuesStr = ":database_id, :import_id, :identifier, :url, :declared_creation_date, now()"
const Database_handle_UpdateStr = "\"database_id\" = :database_id, \"import_id\" = :import_id, \"identifier\" = :identifier, \"url\" = :url, \"declared_creation_date\" = :declared_creation_date"
const Shapefile__authors_InsertStr = ""
const Shapefile__authors_InsertValuesStr = ""
const Shapefile__authors_UpdateStr = ""
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
const Charac_root_InsertStr = "\"admin_group_id\", \"cached_langs\""
const Charac_root_InsertValuesStr = ":admin_group_id, :cached_langs"
const Charac_root_UpdateStr = "\"admin_group_id\" = :admin_group_id, \"cached_langs\" = :cached_langs"
const Chronology_root_InsertStr = "\"admin_group_id\", \"author_user_id\", \"credits\", \"active\", \"geom\", \"cached_langs\""
const Chronology_root_InsertValuesStr = ":admin_group_id, :author_user_id, :credits, :active, :geom, :cached_langs"
const Chronology_root_UpdateStr = "\"admin_group_id\" = :admin_group_id, \"author_user_id\" = :author_user_id, \"credits\" = :credits, \"active\" = :active, \"geom\" = :geom, \"cached_langs\" = :cached_langs"
const Map_layer__authors_InsertStr = ""
const Map_layer__authors_InsertValuesStr = ""
const Map_layer__authors_UpdateStr = ""
const Project__charac_InsertStr = ""
const Project__charac_InsertValuesStr = ""
const Project__charac_UpdateStr = ""
const Saved_query_InsertStr = "\"params\""
const Saved_query_InsertValuesStr = ":params"
const Saved_query_UpdateStr = "\"params\" = :params"
