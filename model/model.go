package model

import (
	"database/sql"
	"time"
)

type Charac struct {
	Id             int       `db:"id" json:"id"`
	Parent_id      int       `db:"parent_id" json:"parent_id"`
	Order          int       `db:"order" json:"order"`
	Author_user_id int       `db:"author_user_id" json:"author_user_id"` // User.Id
	Created_at     time.Time `db:"created_at" json:"created_at"`
	Updated_at     time.Time `db:"updated_at" json:"updated_at"`
}

type Charac_tr struct {
	Lang_id     int            `db:"lang_id" json:"lang_id"`     // Lang.Id
	Charac_id   int            `db:"charac_id" json:"charac_id"` // Charac.Id
	Name        string         `db:"name" json:"name"`
	Description sql.NullString `db:"description" json:"description"`
}

type Chronology struct {
	Id         int       `db:"id" json:"id"`
	Owner_id   int       `db:"owner_id" json:"owner_id"`   // User.Id
	Parent_id  int       `db:"parent_id" json:"parent_id"` // Chronology.Id
	Start_date int       `db:"start_date" json:"start_date"`
	End_date   int       `db:"end_date" json:"end_date"`
	Color      string    `db:"color" json:"color"`
	Created_at time.Time `db:"created_at" json:"created_at"`
	Updated_at time.Time `db:"updated_at" json:"updated_at"`
}

type Chronology_tr struct {
	Lang_id       int            `db:"lang_id" json:"lang_id"`             // Lang.Id
	Chronology_id int            `db:"chronology_id" json:"chronology_id"` // Chronology.Id
	Name          string         `db:"name" json:"name"`
	Description   sql.NullString `db:"description" json:"description"`
}

type City struct {
	Geonameid         int            `db:"geonameid" json:"geonameid"`
	Country_geonameid int            `db:"country_geonameid" json:"country_geonameid"` // Country.Geonameid
	Geom              sql.NullString `db:"geom" json:"geom"`
	Geom_centroid     sql.NullString `db:"geom_centroid" json:"geom_centroid"`
}

type City_tr struct {
	City_geonameid int    `db:"city_geonameid" json:"city_geonameid"` // City.Geonameid
	Lang_id        int    `db:"lang_id" json:"lang_id"`               // Lang.Id
	Name           string `db:"name" json:"name"`
	Name_ascii     string `db:"name_ascii" json:"name_ascii"`
}

type Company struct {
	Id             int    `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	City_geonameid int    `db:"city_geonameid" json:"city_geonameid"` // City.Geonameid
}

type Continent struct {
	Geonameid  int            `db:"geonameid" json:"geonameid"`
	Iso_code   string         `db:"iso_code" json:"iso_code"`
	Geom       sql.NullString `db:"geom" json:"geom"`
	Created_at time.Time      `db:"created_at" json:"created_at"`
	Updated_at time.Time      `db:"updated_at" json:"updated_at"`
}

type Continent_tr struct {
	Continent_geonameid int    `db:"continent_geonameid" json:"continent_geonameid"`
	Lang_id             int    `db:"lang_id" json:"lang_id"` // Lang.Id
	Name                string `db:"name" json:"name"`
	Name_ascii          string `db:"name_ascii" json:"name_ascii"`
}

type Country struct {
	Geonameid  int            `db:"geonameid" json:"geonameid"`
	Iso_code   sql.NullString `db:"iso_code" json:"iso_code"`
	Geom       sql.NullString `db:"geom" json:"geom"`
	Created_at time.Time      `db:"created_at" json:"created_at"`
	Updated_at time.Time      `db:"updated_at" json:"updated_at"`
}

type Country_tr struct {
	Country_geonameid int            `db:"country_geonameid" json:"country_geonameid"` // Country.Geonameid
	Lang_id           int            `db:"lang_id" json:"lang_id"`                     // Lang.Id
	Name              string         `db:"name" json:"name"`
	Name_ascii        sql.NullString `db:"name_ascii" json:"name_ascii"`
}

type Database struct {
	Id                   int       `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	Scale_resolution     string    `db:"scale_resolution" json:"scale_resolution"`
	Geographical_extent  string    `db:"geographical_extent" json:"geographical_extent"`
	Type                 string    `db:"type" json:"type"`
	Owner                int       `db:"owner" json:"owner"` // User.Id
	Source_creation_date time.Time `db:"source_creation_date" json:"source_creation_date"`
	Data_set             string    `db:"data_set" json:"data_set"`
	Identifier           string    `db:"identifier" json:"identifier"`
	Source               string    `db:"source" json:"source"`
	Source_url           string    `db:"source_url" json:"source_url"`
	Publisher            string    `db:"publisher" json:"publisher"`
	Contributor          string    `db:"contributor" json:"contributor"`
	Default_language     int       `db:"default_language" json:"default_language"` // Lang.Id
	Relation             string    `db:"relation" json:"relation"`
	Coverage             string    `db:"coverage" json:"coverage"`
	Copyright            string    `db:"copyright" json:"copyright"`
	State                string    `db:"state" json:"state"`
	Published            bool      `db:"published" json:"published"`
	License_id           int       `db:"license_id" json:"license_id"` // License.Id
	Context              string    `db:"context" json:"context"`
	Context_description  string    `db:"context_description" json:"context_description"`
	Subject              string    `db:"subject" json:"subject"`
	Created_at           time.Time `db:"created_at" json:"created_at"`
	Updated_at           time.Time `db:"updated_at" json:"updated_at"`
}

type Database__authors struct {
	Database_id int `db:"database_id" json:"database_id"` // Database.Id
	User_id     int `db:"user_id" json:"user_id"`         // User.Id
}

type Database__continent struct {
	Id_database         int `db:"id_database" json:"id_database"`                 // Database.Id
	Continent_geonameid int `db:"continent_geonameid" json:"continent_geonameid"` // Continent.Geonameid
}

type Database__country struct {
	Database_id       int `db:"database_id" json:"database_id"`             // Database.Id
	Country_geonameid int `db:"country_geonameid" json:"country_geonameid"` // Country.Geonameid
}

type Database__handle struct {
	Id          int       `db:"id" json:"id"`
	Handle_id   int       `db:"handle_id" json:"handle_id"`     // Handle.Id
	Database_id int       `db:"database_id" json:"database_id"` // Database.Id
	Value       string    `db:"value" json:"value"`
	Created_at  time.Time `db:"created_at" json:"created_at"`
}

type Database_tr struct {
	Database_id        int            `db:"database_id" json:"database_id"` // Database.Id
	Lang_id            int            `db:"lang_id" json:"lang_id"`         // Lang.Id
	Description        sql.NullString `db:"description" json:"description"`
	Geographical_limit sql.NullString `db:"geographical_limit" json:"geographical_limit"`
	Bibliography       sql.NullString `db:"bibliography" json:"bibliography"`
}

type Geographical_zone struct {
	Id   int            `db:"id" json:"id"`
	Name sql.NullString `db:"name" json:"name"`
	Geom string         `db:"geom" json:"geom"`
}

type Global_project struct {
	Project_id int `db:"project_id" json:"project_id"` // Project.Id
}

type Group struct {
	Id         int       `db:"id" json:"id"`
	Created_at time.Time `db:"created_at" json:"created_at"`
	Udpated_at time.Time `db:"udpated_at" json:"udpated_at"`
}

type Group__permission struct {
	Group_id      int       `db:"group_id" json:"group_id"`           // Group.Id
	Permission_id int       `db:"permission_id" json:"permission_id"` // Permission.Id
	Created_at    time.Time `db:"created_at" json:"created_at"`
	Updated_at    time.Time `db:"updated_at" json:"updated_at"`
}

type Group_tr struct {
	Group_id    int            `db:"group_id" json:"group_id"` // Group.Id
	Lang_id     int            `db:"lang_id" json:"lang_id"`   // Lang.Id
	Name        string         `db:"name" json:"name"`
	Description sql.NullString `db:"description" json:"description"`
}

type Handle struct {
	Id   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Url  string `db:"url" json:"url"`
}

type Import struct {
	Id          int       `db:"id" json:"id"`
	Database_id int       `db:"database_id" json:"database_id"` // Database.Id
	User_id     int       `db:"user_id" json:"user_id"`         // User.Id
	Filename    string    `db:"filename" json:"filename"`
	Created_at  time.Time `db:"created_at" json:"created_at"`
}

type Lang struct {
	Id       int    `db:"id" json:"id"`
	Iso_code string `db:"iso_code" json:"iso_code"`
	Active   bool   `db:"active" json:"active"`
}

type License struct {
	Id   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Url  string `db:"url" json:"url"`
}

type Permission struct {
	Id         int       `db:"id" json:"id"`
	Created_at time.Time `db:"created_at" json:"created_at"`
	Updated_at time.Time `db:"updated_at" json:"updated_at"`
	Name       string    `db:"name" json:"name"`
}

type Permission_tr struct {
	Permission_id int            `db:"permission_id" json:"permission_id"` // Permission.Id
	Lang_id       int            `db:"lang_id" json:"lang_id"`             // Lang.Id
	Name          string         `db:"name" json:"name"`
	Description   sql.NullString `db:"description" json:"description"`
}

type Project struct {
	Id                   int       `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	User_id              int       `db:"user_id" json:"user_id"`                           // User.Id
	Geographical_zone_id int       `db:"geographical_zone_id" json:"geographical_zone_id"` // Geographical_zone.Id
	Created_at           time.Time `db:"created_at" json:"created_at"`
	Updated_at           time.Time `db:"updated_at" json:"updated_at"`
}

type Project__charac struct {
	Project__characs_set_id int `db:"project__characs_set_id" json:"project__characs_set_id"` // Project__characs_set.Id
	Charac_id               int `db:"charac_id" json:"charac_id"`                             // Charac.Id
}

type Project__characs_set struct {
	Id                                 int `db:"id" json:"id"`
	Project_id                         int `db:"project_id" json:"project_id"`                                                 // Project.Id
	Copied_from_project__charac_set_id int `db:"copied_from_project__charac_set_id" json:"copied_from_project__charac_set_id"` // Project__characs_set.Id
}

type Project__characs_set_tr struct {
	Project__characs_set_id int            `db:"project__characs_set_id" json:"project__characs_set_id"` // Project__characs_set.Id
	Lang_id                 int            `db:"lang_id" json:"lang_id"`                                 // Lang.Id
	Name                    string         `db:"name" json:"name"`
	Description             sql.NullString `db:"description" json:"description"`
}

type Project__chronology struct {
	Project_id         int `db:"project_id" json:"project_id"`                 // Project.Id
	Chronology_root_id int `db:"chronology_root_id" json:"chronology_root_id"` // Chronology.Id
	Id_group           int `db:"id_group" json:"id_group"`                     // Group.Id
}

type Project__chronology_tr struct {
	Project__chronology_id int            `db:"project__chronology_id" json:"project__chronology_id"`
	Lang_id                int            `db:"lang_id" json:"lang_id"` // Lang.Id
	Name                   string         `db:"name" json:"name"`
	Description            sql.NullString `db:"description" json:"description"`
}

type Project__databases struct {
	Project_id  int `db:"project_id" json:"project_id"`   // Project.Id
	Database_id int `db:"database_id" json:"database_id"` // Database.Id
}

type Project__shapefile struct {
	Project_id   int `db:"project_id" json:"project_id"`     // Project.Id
	Shapefile_id int `db:"shapefile_id" json:"shapefile_id"` // Shapefile.Id
}

type Project__wms_map struct {
	Project_id int `db:"project_id" json:"project_id"` // Project.Id
	Wms_map_id int `db:"wms_map_id" json:"wms_map_id"` // Wms_map.Id
}

type Shapefile struct {
	Id                   int            `db:"id" json:"id"`
	Creator_user_id      int            `db:"creator_user_id" json:"creator_user_id"` // User.Id
	Source_creation_date time.Time      `db:"source_creation_date" json:"source_creation_date"`
	Filename             sql.NullString `db:"filename" json:"filename"`
	Geom                 sql.NullString `db:"geom" json:"geom"`
	Min_scale            int            `db:"min_scale" json:"min_scale"`
	Max_scale            int            `db:"max_scale" json:"max_scale"`
	Start_date           int            `db:"start_date" json:"start_date"`
	Start_date_qualifier string         `db:"start_date_qualifier" json:"start_date_qualifier"`
	End_date             int            `db:"end_date" json:"end_date"`
	End_date_qualifier   string         `db:"end_date_qualifier" json:"end_date_qualifier"`
	Published            bool           `db:"published" json:"published"`
	Created_at           time.Time      `db:"created_at" json:"created_at"`
	Updated_at           time.Time      `db:"updated_at" json:"updated_at"`
	License_name         string         `db:"license_name" json:"license_name"`
	Odbl_license         bool           `db:"odbl_license" json:"odbl_license"`
}

type Shapefile_authors struct {
	Shapefile_id int `db:"shapefile_id" json:"shapefile_id"` // Shapefile.Id
	User_id      int `db:"user_id" json:"user_id"`           // User.Id
}

type Shapefile_tr struct {
	Shapefile_id int            `db:"shapefile_id" json:"shapefile_id"` // Shapefile.Id
	Lang_id      int            `db:"lang_id" json:"lang_id"`           // Lang.Id
	Name         string         `db:"name" json:"name"`
	Description  sql.NullString `db:"description" json:"description"`
	Attribution  sql.NullString `db:"attribution" json:"attribution"`
	Bibiography  string         `db:"bibiography" json:"bibiography"`
}

type Site struct {
	Id             int       `db:"id" json:"id"`
	Code           string    `db:"code" json:"code"`
	Name           string    `db:"name" json:"name"`
	City_name      string    `db:"city_name" json:"city_name"`
	City_geonameid int       `db:"city_geonameid" json:"city_geonameid"`
	Geom           string    `db:"geom" json:"geom"`
	Centroid       bool      `db:"centroid" json:"centroid"`
	Occupation     string    `db:"occupation" json:"occupation"`
	Database_id    int       `db:"database_id" json:"database_id"` // Database.Id
	Created_at     time.Time `db:"created_at" json:"created_at"`
	Updated_at     time.Time `db:"updated_at" json:"updated_at"`
}

type Site_range struct {
	Id                   int       `db:"id" json:"id"`
	Site_id              int       `db:"site_id" json:"site_id"` // Site.Id
	Start_date           int       `db:"start_date" json:"start_date"`
	Start_date_qualifier string    `db:"start_date_qualifier" json:"start_date_qualifier"`
	End_date             int       `db:"end_date" json:"end_date"`
	End_date_qualifier   string    `db:"end_date_qualifier" json:"end_date_qualifier"`
	Knowledge_type       string    `db:"knowledge_type" json:"knowledge_type"`
	Depth                int       `db:"depth" json:"depth"`
	Created_at           time.Time `db:"created_at" json:"created_at"`
	Updated_at           time.Time `db:"updated_at" json:"updated_at"`
}

type Site_range__charac struct {
	Site_range_id int  `db:"site_range_id" json:"site_range_id"` // Site_range.Id
	Charac_id     int  `db:"charac_id" json:"charac_id"`         // Charac.Id
	Exceptional   bool `db:"exceptional" json:"exceptional"`
}

type Site_range_tr struct {
	Site_range_id int    `db:"site_range_id" json:"site_range_id"` // Site_range.Id
	Lang_id       int    `db:"lang_id" json:"lang_id"`             // Lang.Id
	Comment       string `db:"comment" json:"comment"`
	Bibliography  string `db:"bibliography" json:"bibliography"`
}

type Site_tr struct {
	Site_id     int    `db:"site_id" json:"site_id"` // Site.Id
	Lang_id     int    `db:"lang_id" json:"lang_id"` // Lang.Id
	Description string `db:"description" json:"description"`
}

type User struct {
	Id             int       `db:"id" json:"id"`
	Username       string    `db:"username" json:"username" min:"2" max:"32" error:"USER.FIELD_USERNAME.T_CHECK_INCORRECT"`
	Firstname      string    `db:"firstname" json:"firstname" min:"1" max:"32" error:"USER.FIELD_FIRSTNAME.T_CHECK_MANDATORY"`
	Lastname       string    `db:"lastname" json:"lastname" min:"1" max:"32" error:"USER.FIELD_LASTNAME.T_CHECK_MANDATORY"`
	Email          string    `db:"email" json:"email" email:"1" error:"USER.FIELD_EMAIL.T_CHECK_INCORRECT"`
	Password       string    `db:"password" json:"password" min:"6" max:"32" error:"USER.FIELD_PASSWORD.T_CHECK_INCORRECT"`
	Description    string    `db:"description" json:"description" min:"1" max:"2048" error:"USER.FIELD_DESCRIPTION.T_CHECK_MANDATORY"`
	Active         bool      `db:"active" json:"active" enum:"0,1" error:"USER.FIELD_ACTIVE.T_CHECK_MANDATORY"`
	First_lang_id  int       `db:"first_lang_id" json:"first_lang_id"`   // Lang.Id
	Second_lang_id int       `db:"second_lang_id" json:"second_lang_id"` // Lang.Id
	City_geonameid int       `db:"city_geonameid" json:"city_geonameid"` // City.Geonameid
	Created_at     time.Time `db:"created_at" json:"created_at"`
	Updated_at     time.Time `db:"updated_at" json:"updated_at"`
}

type User__company struct {
	User_id    int `db:"user_id" json:"user_id"`       // User.Id
	Company_id int `db:"company_id" json:"company_id"` // Company.Id
}

type User__group struct {
	Group_id int `db:"group_id" json:"group_id"` // Group.Id
	User_id  int `db:"user_id" json:"user_id"`   // User.Id
}

type User_preferences struct {
	User_id int            `db:"user_id" json:"user_id"` // User.Id
	Key     string         `db:"key" json:"key"`
	Value   sql.NullString `db:"value" json:"value"`
}

type User_project struct {
	Project_id int `db:"project_id" json:"project_id"` // Project.Id
	Start_date int `db:"start_date" json:"start_date"`
	End_date   int `db:"end_date" json:"end_date"`
}

type Wms_map struct {
	Id                   int            `db:"id" json:"id"`
	Creator_user_id      int            `db:"creator_user_id" json:"creator_user_id"` // User.Id
	Url                  sql.NullString `db:"url" json:"url"`
	Source_creation_date time.Time      `db:"source_creation_date" json:"source_creation_date"`
	Min_scale            int            `db:"min_scale" json:"min_scale"`
	Max_scale            int            `db:"max_scale" json:"max_scale"`
	Start_date           int            `db:"start_date" json:"start_date"`
	Start_date_qualifier string         `db:"start_date_qualifier" json:"start_date_qualifier"`
	End_date             int            `db:"end_date" json:"end_date"`
	End_date_qualifier   string         `db:"end_date_qualifier" json:"end_date_qualifier"`
	Published            bool           `db:"published" json:"published"`
	Created_at           time.Time      `db:"created_at" json:"created_at"`
	Updated_at           time.Time      `db:"updated_at" json:"updated_at"`
}

type Wms_map_tr struct {
	Wms_map_id  int            `db:"wms_map_id" json:"wms_map_id"` // Wms_map.Id
	Lang_id     int            `db:"lang_id" json:"lang_id"`       // Lang.Id
	Name        string         `db:"name" json:"name"`
	Attribution sql.NullString `db:"attribution" json:"attribution"`
	Copyright   sql.NullString `db:"copyright" json:"copyright"`
}

const User_InsertStr = "\"username\", \"firstname\", \"lastname\", \"email\", \"password\", \"description\", \"active\", \"first_lang_id\", \"second_lang_id\", \"city_geonameid\", \"created_at\", \"updated_at\""
const User_InsertValuesStr = ":username, :firstname, :lastname, :email, :password, :description, :active, :first_lang_id, :second_lang_id, :city_geonameid, now(), now()"
const User_UpdateStr = "\"username\" = :username, \"firstname\" = :firstname, \"lastname\" = :lastname, \"email\" = :email, \"password\" = :password, \"description\" = :description, \"active\" = :active, \"first_lang_id\" = :first_lang_id, \"second_lang_id\" = :second_lang_id, \"city_geonameid\" = :city_geonameid, \"updated_at\" = now()"
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
const Project_UpdateStr = "\"name\" = :name, \"user_id\" = :user_id, \"geographical_zone_id\" = :geographical_zone_id, , \"updated_at\" = now()"
const Geographical_zone_InsertStr = "\"name\", \"geom\""
const Geographical_zone_InsertValuesStr = ":name, :geom"
const Geographical_zone_UpdateStr = "\"name\" = :name, \"geom\" = :geom"
const Chronology_InsertStr = "\"owner_id\", \"parent_id\", \"start_date\", \"end_date\", \"color\", \"created_at\", \"updated_at\""
const Chronology_InsertValuesStr = ":owner_id, :parent_id, :start_date, :end_date, :color, now(), now()"
const Chronology_UpdateStr = "\"owner_id\" = :owner_id, \"parent_id\" = :parent_id, \"start_date\" = :start_date, \"end_date\" = :end_date, \"color\" = :color, , \"updated_at\" = now()"
const Chronology_tr_InsertStr = "\"name\", \"description\""
const Chronology_tr_InsertValuesStr = ":name, :description"
const Chronology_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Project__chronology_InsertStr = "\"id_group\""
const Project__chronology_InsertValuesStr = ":id_group"
const Project__chronology_UpdateStr = "\"id_group\" = :id_group"
const Database_InsertStr = "\"name\", \"scale_resolution\", \"geographical_extent\", \"type\", \"owner\", \"source_creation_date\", \"data_set\", \"identifier\", \"source\", \"source_url\", \"publisher\", \"contributor\", \"default_language\", \"relation\", \"coverage\", \"copyright\", \"state\", \"published\", \"license_id\", \"context\", \"context_description\", \"subject\", \"created_at\", \"updated_at\""
const Database_InsertValuesStr = ":name, :scale_resolution, :geographical_extent, :type, :owner, :source_creation_date, :data_set, :identifier, :source, :source_url, :publisher, :contributor, :default_language, :relation, :coverage, :copyright, :state, :published, :license_id, :context, :context_description, :subject, now(), now()"
const Database_UpdateStr = "\"name\" = :name, \"scale_resolution\" = :scale_resolution, \"geographical_extent\" = :geographical_extent, \"type\" = :type, \"owner\" = :owner, \"source_creation_date\" = :source_creation_date, \"data_set\" = :data_set, \"identifier\" = :identifier, \"source\" = :source, \"source_url\" = :source_url, \"publisher\" = :publisher, \"contributor\" = :contributor, \"default_language\" = :default_language, \"relation\" = :relation, \"coverage\" = :coverage, \"copyright\" = :copyright, \"state\" = :state, \"published\" = :published, \"license_id\" = :license_id, \"context\" = :context, \"context_description\" = :context_description, \"subject\" = :subject, , \"updated_at\" = now()"
const Site_InsertStr = "\"code\", \"name\", \"city_name\", \"city_geonameid\", \"geom\", \"centroid\", \"occupation\", \"database_id\", \"created_at\", \"updated_at\""
const Site_InsertValuesStr = ":code, :name, :city_name, :city_geonameid, :geom, :centroid, :occupation, :database_id, now(), now()"
const Site_UpdateStr = "\"code\" = :code, \"name\" = :name, \"city_name\" = :city_name, \"city_geonameid\" = :city_geonameid, \"geom\" = :geom, \"centroid\" = :centroid, \"occupation\" = :occupation, \"database_id\" = :database_id, , \"updated_at\" = now()"
const Database_tr_InsertStr = "\"description\", \"geographical_limit\", \"bibliography\""
const Database_tr_InsertValuesStr = ":description, :geographical_limit, :bibliography"
const Database_tr_UpdateStr = "\"description\" = :description, \"geographical_limit\" = :geographical_limit, \"bibliography\" = :bibliography"
const City_InsertStr = "\"country_geonameid\", \"geom\", \"geom_centroid\""
const City_InsertValuesStr = ":country_geonameid, :geom, :geom_centroid"
const City_UpdateStr = "\"country_geonameid\" = :country_geonameid, \"geom\" = :geom, \"geom_centroid\" = :geom_centroid"
const Site_tr_InsertStr = "\"description\""
const Site_tr_InsertValuesStr = ":description"
const Site_tr_UpdateStr = "\"description\" = :description"
const Site_range_InsertStr = "\"site_id\", \"start_date\", \"start_date_qualifier\", \"end_date\", \"end_date_qualifier\", \"knowledge_type\", \"depth\", \"created_at\", \"updated_at\""
const Site_range_InsertValuesStr = ":site_id, :start_date, :start_date_qualifier, :end_date, :end_date_qualifier, :knowledge_type, :depth, now(), now()"
const Site_range_UpdateStr = "\"site_id\" = :site_id, \"start_date\" = :start_date, \"start_date_qualifier\" = :start_date_qualifier, \"end_date\" = :end_date, \"end_date_qualifier\" = :end_date_qualifier, \"knowledge_type\" = :knowledge_type, \"depth\" = :depth, , \"updated_at\" = now()"
const Site_range_tr_InsertStr = "\"comment\", \"bibliography\""
const Site_range_tr_InsertValuesStr = ":comment, :bibliography"
const Site_range_tr_UpdateStr = "\"comment\" = :comment, \"bibliography\" = :bibliography"
const City_tr_InsertStr = "\"name\", \"name_ascii\""
const City_tr_InsertValuesStr = ":name, :name_ascii"
const City_tr_UpdateStr = "\"name\" = :name, \"name_ascii\" = :name_ascii"
const Charac_tr_InsertStr = "\"name\", \"description\""
const Charac_tr_InsertValuesStr = ":name, :description"
const Charac_tr_UpdateStr = "\"name\" = :name, \"description\" = :description"
const Charac_InsertStr = "\"parent_id\", \"order\", \"author_user_id\", \"created_at\", \"updated_at\""
const Charac_InsertValuesStr = ":parent_id, :order, :author_user_id, now(), now()"
const Charac_UpdateStr = "\"parent_id\" = :parent_id, \"order\" = :order, \"author_user_id\" = :author_user_id, , \"updated_at\" = now()"
const Wms_map_InsertStr = "\"creator_user_id\", \"url\", \"source_creation_date\", \"min_scale\", \"max_scale\", \"start_date\", \"start_date_qualifier\", \"end_date\", \"end_date_qualifier\", \"published\", \"created_at\", \"updated_at\""
const Wms_map_InsertValuesStr = ":creator_user_id, :url, :source_creation_date, :min_scale, :max_scale, :start_date, :start_date_qualifier, :end_date, :end_date_qualifier, :published, now(), now()"
const Wms_map_UpdateStr = "\"creator_user_id\" = :creator_user_id, \"url\" = :url, \"source_creation_date\" = :source_creation_date, \"min_scale\" = :min_scale, \"max_scale\" = :max_scale, \"start_date\" = :start_date, \"start_date_qualifier\" = :start_date_qualifier, \"end_date\" = :end_date, \"end_date_qualifier\" = :end_date_qualifier, \"published\" = :published, , \"updated_at\" = now()"
const Wms_map_tr_InsertStr = "\"name\", \"attribution\", \"copyright\""
const Wms_map_tr_InsertValuesStr = ":name, :attribution, :copyright"
const Wms_map_tr_UpdateStr = "\"name\" = :name, \"attribution\" = :attribution, \"copyright\" = :copyright"
const Shapefile_InsertStr = "\"creator_user_id\", \"source_creation_date\", \"filename\", \"geom\", \"min_scale\", \"max_scale\", \"start_date\", \"start_date_qualifier\", \"end_date\", \"end_date_qualifier\", \"published\", \"created_at\", \"updated_at\", \"license_name\", \"odbl_license\""
const Shapefile_InsertValuesStr = ":creator_user_id, :source_creation_date, :filename, :geom, :min_scale, :max_scale, :start_date, :start_date_qualifier, :end_date, :end_date_qualifier, :published, now(), now(), :license_name, :odbl_license"
const Shapefile_UpdateStr = "\"creator_user_id\" = :creator_user_id, \"source_creation_date\" = :source_creation_date, \"filename\" = :filename, \"geom\" = :geom, \"min_scale\" = :min_scale, \"max_scale\" = :max_scale, \"start_date\" = :start_date, \"start_date_qualifier\" = :start_date_qualifier, \"end_date\" = :end_date, \"end_date_qualifier\" = :end_date_qualifier, \"published\" = :published, , \"updated_at\" = now(), \"license_name\" = :license_name, \"odbl_license\" = :odbl_license"
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
const Site_range__charac_InsertStr = "\"exceptional\""
const Site_range__charac_InsertValuesStr = ":exceptional"
const Site_range__charac_UpdateStr = "\"exceptional\" = :exceptional"
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
const Group_InsertStr = "\"created_at\", \"udpated_at\""
const Group_InsertValuesStr = "now(), :udpated_at"
const Group_UpdateStr = ", \"udpated_at\" = :udpated_at"
const Permission_InsertStr = "\"created_at\", \"updated_at\", \"name\""
const Permission_InsertValuesStr = "now(), now(), :name"
const Permission_UpdateStr = ", \"updated_at\" = now(), \"name\" = :name"
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
const Group__permission_UpdateStr = ", \"updated_at\" = now()"
const Country_InsertStr = "\"iso_code\", \"geom\", \"created_at\", \"updated_at\""
const Country_InsertValuesStr = ":iso_code, :geom, now(), now()"
const Country_UpdateStr = "\"iso_code\" = :iso_code, \"geom\" = :geom, , \"updated_at\" = now()"
const User_preferences_InsertStr = "\"value\""
const User_preferences_InsertValuesStr = ":value"
const User_preferences_UpdateStr = "\"value\" = :value"
const Country_tr_InsertStr = "\"name\", \"name_ascii\""
const Country_tr_InsertValuesStr = ":name, :name_ascii"
const Country_tr_UpdateStr = "\"name\" = :name, \"name_ascii\" = :name_ascii"
const Database__authors_InsertStr = ""
const Database__authors_InsertValuesStr = ""
const Database__authors_UpdateStr = ""
const Import_InsertStr = "\"database_id\", \"user_id\", \"filename\", \"created_at\""
const Import_InsertValuesStr = ":database_id, :user_id, :filename, now()"
const Import_UpdateStr = "\"database_id\" = :database_id, \"user_id\" = :user_id, \"filename\" = :filename, "
const License_InsertStr = "\"name\", \"url\""
const License_InsertValuesStr = ":name, :url"
const License_UpdateStr = "\"name\" = :name, \"url\" = :url"
const Handle_InsertStr = "\"name\", \"url\""
const Handle_InsertValuesStr = ":name, :url"
const Handle_UpdateStr = "\"name\" = :name, \"url\" = :url"
const Database__country_InsertStr = ""
const Database__country_InsertValuesStr = ""
const Database__country_UpdateStr = ""
const Database__handle_InsertStr = "\"handle_id\", \"database_id\", \"value\", \"created_at\""
const Database__handle_InsertValuesStr = ":handle_id, :database_id, :value, now()"
const Database__handle_UpdateStr = "\"handle_id\" = :handle_id, \"database_id\" = :database_id, \"value\" = :value, "
const Shapefile_authors_InsertStr = ""
const Shapefile_authors_InsertValuesStr = ""
const Shapefile_authors_UpdateStr = ""
const Continent_tr_InsertStr = "\"name\", \"name_ascii\""
const Continent_tr_InsertValuesStr = ":name, :name_ascii"
const Continent_tr_UpdateStr = "\"name\" = :name, \"name_ascii\" = :name_ascii"
const Continent_InsertStr = "\"iso_code\", \"geom\", \"created_at\", \"updated_at\""
const Continent_InsertValuesStr = ":iso_code, :geom, now(), now()"
const Continent_UpdateStr = "\"iso_code\" = :iso_code, \"geom\" = :geom, , \"updated_at\" = now()"
const Database__continent_InsertStr = ""
const Database__continent_InsertValuesStr = ""
const Database__continent_UpdateStr = ""
