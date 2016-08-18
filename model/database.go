/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2016 CROLL SAS
 *
 * Authors :
 *  Christophe Beveraggi <beve@croll.fr>
 *  Nicolas Dimitrijevic <nicolas@croll.fr>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package model

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
)

// DatabaseAuthor stores essential informations about an author
type DatabaseAuthor struct {
	Id        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Fullname  string `json:"fullname"`
}

// CountryInfos store country info and translations
type CountryInfos struct {
	Country
	Country_tr
	Country_geonameid int            `json:"-"`
	Lang_isocode      string         `json:"-"`
	Created_at        time.Time      `json:"-"`
	Updated_at        time.Time      `json:"-"`
	Geom              sql.NullString `json:"-"`
}

// ContinentInfos store country info and translations
type ContinentInfos struct {
	Continent
	Continent_tr
	Continent_geonameid int            `json:"-"`
	Lang_isocode        string         `json:"-"`
	Created_at          time.Time      `json:"-"`
	Updated_at          time.Time      `json:"-"`
	Geom                sql.NullString `json:"-"`
}

// ImportFullInfos stores all informations about an import
type ImportFullInfos struct {
	Import
	Fullname string `json:"fullname"`
}

// DatabaseFullInfos stores all informations about a database
type DatabaseFullInfos struct {
	Database
	Imports             []ImportFullInfos  `json:"imports"`
	Countries           []CountryInfos     `json:"countries"`
	Continents          []ContinentInfos   `json:"continents"`
	Handles             []Database_handle  `json:"handles"`
	Authors             []DatabaseAuthor   `json:"authors"`
	Contexts            []Database_context `json:"contexts"`
	Owner_name          string             `json:"owner_name"`
	License             string             `json:"license"`
	Description         map[string]string  `json:"description"`
	Geographical_limit  map[string]string  `json:"geographical_limit"`
	Bibliography        map[string]string  `json:"bibliography"`
	Context_description map[string]string  `json:"context_description"`
	Source_description  map[string]string  `json:"source_description"`
	Source_relation     map[string]string  `json:"source_relation"`
	Copyright           map[string]string  `json:"copyright"`
	Subject             map[string]string  `json:"subject"`
}

// DoesExist check if database exist with a name and an owner
func (d *Database) DoesExist(tx *sqlx.Tx) (exists bool, err error) {
	exists = false
	err = tx.QueryRowx("SELECT id FROM \"database\" WHERE name = $1 AND owner = $2", d.Name, d.Owner).Scan(&d.Id)
	switch {
	case err == sql.ErrNoRows:
		return exists, nil
	case err != nil:
		return exists, errors.New("database::DoesExist: " + err.Error())
	}
	return true, nil
}

// AnotherExistsWithSameName checks if database already exists with same name and owned by another user
func (d *Database) AnotherExistsWithSameName(tx *sqlx.Tx) (exists bool, err error) {
	exists = false
	err = tx.QueryRowx("SELECT id FROM \"database\" WHERE name = $1 AND owner != $2", d.Name, d.Owner).Scan(&d.Id)
	switch {
	case err == sql.ErrNoRows:
		return exists, nil
	case err != nil:
		return exists, errors.New("database::AnotherExistsWithSameName: " + err.Error())
	}
	return true, nil
}

// Get retrieves informations about a database stored in the main table
func (d *Database) Get(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT * from \"database\" WHERE id=:id")
	defer stmt.Close()
	if err != nil {
		return errors.New("database::Get: " + err.Error())
	}
	return stmt.Get(d, d)
}

// GetFullInfos returns all informations about a database
func (d *Database) GetFullInfos(tx *sqlx.Tx, langIsocode string) (db DatabaseFullInfos, err error) {
	db = DatabaseFullInfos{}

	if d.Id == 0 {
		db.Imports = make([]ImportFullInfos, 0)
		db.Countries = make([]CountryInfos, 0)
		db.Continents = make([]ContinentInfos, 0)
		db.Handles = make([]Database_handle, 0)
		db.Authors = make([]DatabaseAuthor, 0)
		db.Contexts = make([]Database_context, 0)
		db.Handles = make([]Database_handle, 0)
		db.License = "-"
		return
	}

	// err = tx.Get(&db, "SELECT name, scale_resolution, geographical_extent, type, declared_creation_date, owner, editor, contributor, default_language, state, license_id, published, soft_deleted, d.created_at, d.updated_at, firstname || ' ' || lastname as owner_name FROM \"database\" d LEFT JOIN \"user\" u ON d.owner = u.id WHERE d.id = $1", d.Id)

	err = tx.Get(&db, "SELECT d.*, ST_AsGeoJSON(d.geographical_extent_geom) as geographical_extent_geom, firstname || ' ' || lastname as owner_name, l.name AS license FROM \"database\" d LEFT JOIN \"user\" u ON d.owner = u.id LEFT JOIN \"license\" l ON d.license_id = l.id WHERE d.id = $1", d.Id)
	if err != nil {
		return
	}

	db.Authors, err = d.GetAuthorsList(tx)
	if err != nil {
		return
	}
	db.Countries, err = d.GetCountryList(tx, langIsocode)
	if err != nil {
		return
	}
	db.Continents, err = d.GetContinentList(tx, langIsocode)
	if err != nil {
		return
	}
	db.Handles, err = d.GetHandles(tx)
	if err != nil {
		return
	}
	db.Imports, err = d.GetImportList(tx)
	if err != nil {
		return
	}
	db.Contexts, err = d.GetContextsList(tx)
	if err != nil {
		return
	}
	err = db.GetTranslations(tx)
	return
}

// GetFullInfosAsJSON returns all infos about a database
func (d *Database) GetFullInfosAsJSON(tx *sqlx.Tx, langIsocode string) (jsonString string, err error) {

	var q = make([]string, 7)

	q[0] = db.AsJSON("SELECT db.*, ST_AsGeoJSON(db.geographical_extent_geom) as geographical_extent_geom, firstname || ' ' || lastname as owner_name, l.name AS license FROM \"database\" db LEFT JOIN \"user\" u ON db.owner = u.id LEFT JOIN \"license\" l ON d.license_id = l.id WHERE db.id = d.id", false, "infos", true)

	q[1] = db.AsJSON("SELECT u.id, u.firstname, u.lastname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = d.id", true, "authors", true)

	q[2] = db.AsJSON("SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM country c LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid LEFT JOIN country_tr ct ON c.geonameid = ct.country_geonameid WHERE dc.database_id = d.id and ct.lang_isocode = '"+langIsocode+"'", true, "countries", true)

	q[3] = db.AsJSON("SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM continent c LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid LEFT JOIN continent_tr ct ON c.geonameid = ct.continent_geonameid WHERE dc.database_id = d.id AND ct.lang_isocode = '"+langIsocode+"'", true, "continents", true)

	q[4] = db.AsJSON("SELECT i.*, u.firstname, u.lastname FROM import i LEFT JOIN \"user\" u ON i.user_id = u.id WHERE database_id = d.id ORDER BY i.id DESC", true, "imports", true)

	q[5] = db.AsJSON("SELECT context FROM database_context WHERE database_id = d.id", true, "contexts", true)

	q[6] = GetQueryTranslationsAsJSONObject("database_tr", "database_id = d.id", "translations", true, "description", "bibliography", "geographical_limit", "context_description")

	// fmt.Println(q[0])
	// fmt.Println(q[1])
	// fmt.Println(q[2])
	// fmt.Println(q[3])
	// fmt.Println(q[4])
	// fmt.Println(q[5])

	// fmt.Println(db.JSONQueryBuilder(q, "database d", "d.id = "+strconv.Itoa(d.Id)))

	err = tx.Get(&jsonString, db.JSONQueryBuilder(q, "database d", "d.id = "+strconv.Itoa(d.Id)))

	if err != nil {
		err = errors.New("database::GetFullInfosAsJSON: " + err.Error())
	}

	if jsonString == "" {
		jsonString = "null"
	}

	return

}

// Create insert the database into arkeogis db
func (d *Database) Create(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("INSERT INTO \"database\" (" + Database_InsertStr + ") VALUES (" + Database_InsertValuesStr + ") RETURNING id")
	defer stmt.Close()
	if err != nil {
		return errors.New("database::Create: " + err.Error())
	}
	err = stmt.Get(&d.Id, d)
	if err != nil {
		err = errors.New("database::Create: " + err.Error())
	}
	return
}

// Update database informations
func (d *Database) Update(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("UPDATE \"database\" SET "+Database_UpdateStr+" WHERE id=:id", d)
	if err != nil {
		err = errors.New("database::Update: " + err.Error())
	}
	return
}

// Delete database and all infos associated
func (d *Database) Delete(tx *sqlx.Tx) (err error) {

	err = d.DeleteContexts(tx)
	if err != nil {
		return
	}

	err = d.DeleteContinents(tx)
	if err != nil {
		return
	}

	err = d.DeleteCountries(tx)
	if err != nil {
		return
	}

	err = d.DeleteAuthors(tx)
	if err != nil {
		return
	}

	err = d.DeleteHandles(tx)
	if err != nil {
		return
	}

	err = d.DeleteImports(tx)
	if err != nil {
		return
	}

	err = d.DeleteSites(tx)
	if err != nil {
		return
	}

	_, err = tx.NamedExec("DELETE FROM \"database_tr\" WHERE database_id=:id", d)
	if err != nil {
		return
	}

	_, err = tx.NamedExec("DELETE FROM \"project__databases\" WHERE database_id=:id", d)
	if err != nil {
		return
	}

	_, err = tx.NamedExec("DELETE FROM \"database\" WHERE id=:id", d)

	return
}

// DeleteSites deletes all sites linked to a database
func (d *Database) DeleteSites(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"site\" WHERE database_id=:id", d)
	return
}

// GetCountryList lists all countries linked to a database
func (d *Database) GetCountryList(tx *sqlx.Tx, langIsocode string) ([]CountryInfos, error) {
	countries := []CountryInfos{}
	err := tx.Select(&countries, "SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM country c LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid LEFT JOIN country_tr ct ON c.geonameid = ct.country_geonameid WHERE dc.database_id = $1 and ct.lang_isocode = $2", d.Id, langIsocode)
	if err != nil {
		err = errors.New("database::GetCountryList: " + err.Error())
	}
	return countries, err
}

// AddCountries links countries to a database
func (d *Database) AddCountries(tx *sqlx.Tx, countryIds []int) (err error) {
	for _, id := range countryIds {
		_, err := tx.Exec("INSERT INTO database__country (database_id, country_geonameid) VALUES ($1, $2)", d.Id, id)
		if err != nil {
			return errors.New("database::AddCountries: " + err.Error())
		}
	}
	return
}

// DeleteCountries unlinks countries to a database
func (d *Database) DeleteCountries(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"database__country\" WHERE database_id=:id", d)
	if err != nil {
		err = errors.New("database::DeleteCountries: " + err.Error())
	}
	return
}

// GetContinentList lists all continents linked to a database
func (d *Database) GetContinentList(tx *sqlx.Tx, langIsocode string) (continents []ContinentInfos, err error) {
	continents = []ContinentInfos{}
	err = tx.Select(&continents, "SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM continent c LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid LEFT JOIN continent_tr ct ON c.geonameid = ct.continent_geonameid WHERE dc.database_id = $1 AND ct.lang_isocode = $2", d.Id, langIsocode)
	if err != nil {
		err = errors.New("database::GetContinentList: " + err.Error())
	}
	return continents, err
}

// AddContinents links continents to a database
func (d *Database) AddContinents(tx *sqlx.Tx, continentIds []int) (err error) {
	for _, id := range continentIds {
		_, err := tx.Exec("INSERT INTO database__continent (database_id, continent_geonameid) VALUES ($1, $2)", d.Id, id)
		if err != nil {
			return errors.New("database::AddContinents: " + err.Error())
		}
	}
	return
}

// DeleteContinents unlinks countries to a database
func (d *Database) DeleteContinents(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"database__continent\" WHERE database_id=:id", d)
	if err != nil {
		err = errors.New("database::DeleteContinents: " + err.Error())
	}
	return
}

// GetHandles get last handle linked to a database
func (d *Database) GetLastHandle(tx *sqlx.Tx) (handle *Database_handle, err error) {
	handle = &Database_handle{}
	err = tx.Get(handle, "SELECT * FROM database_handle WHERE database_id = $1 ORDER BY id DESC LIMIT 1", d.Id)
	switch {
	case err == sql.ErrNoRows:
		return handle, nil
	case err != nil:
		return
	}
	return
}

// GetHandles lists all handles linked to a database
func (d *Database) GetHandles(tx *sqlx.Tx) (handles []Database_handle, err error) {
	handles = []Database_handle{}
	err = tx.Select(&handles, "SELECT import_id, identifier, url, declared_creation_date, created_at FROM database_handle WHERE database_id = $1", d.Id)
	if err != nil {
		err = errors.New("database::GetHandles: " + err.Error())
	}
	return
}

// AddHandle links a handle  to a database
func (d *Database) AddHandle(tx *sqlx.Tx, handle *Database_handle) (id int, err error) {
	stmt, err := tx.PrepareNamed("INSERT INTO \"database_handle\" (" + Database_handle_InsertStr + ") VALUES (" + Database_handle_InsertValuesStr + ") RETURNING id")
	defer stmt.Close()
	if err != nil {
		return id, errors.New("database::AddHandle: " + err.Error())
	}
	err = stmt.Get(&id, handle)
	if err != nil {
		err = errors.New("database::AddHandle: " + err.Error())
	}
	return
}

// UpdateHandle links continents to a database
func (d *Database) UpdateHandle(tx *sqlx.Tx, handle *Database_handle) (err error) {
	_, err = tx.NamedExec("UPDATE database_handle SET "+Database_handle_UpdateStr+" WHERE id = :id", handle)
	if err != nil {
		err = errors.New("database::UpdateHandle: " + err.Error())
	}
	return
}

// DeleteHandles unlinks handles
func (d *Database) DeleteSpecificHandle(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec("DELETE FROM \"database_handle\" WHERE identifier = $1", id)
	if err != nil {
		err = errors.New("database::DeleteHandle: " + err.Error())
	}
	return err
}

func (d *Database) DeleteHandles(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"database_handle\" WHERE database_id = :id", d)
	if err != nil {
		err = errors.New("database::DeleteHandles: " + err.Error())
	}
	return err
}

// DeleteImports unlinks database import logs
func (d *Database) DeleteImports(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"import\" WHERE database_id= :id", d)
	if err != nil {
		err = errors.New("database::DeleteImports: " + err.Error())
	}
	return err
}

// GetAuthorsList lists all user designed as author of a database
func (d *Database) GetAuthorsList(tx *sqlx.Tx) (authors []DatabaseAuthor, err error) {
	err = tx.Select(&authors, "SELECT u.id, u.firstname, u.lastname, u.firstname || ' ' || u.lastname AS fullname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = $1", d.Id)
	if err != nil {
		err = errors.New("database::GetAuthorsList: " + err.Error())
	}
	return
}

// SetAuthors links users as authors to a database
func (d *Database) SetAuthors(tx *sqlx.Tx, authors []int) (err error) {
	for _, uid := range authors {
		_, err = tx.Exec("INSERT INTO \"database__authors\" (database_id, user_id) VALUES ($1, $2)", d.Id, uid)
		if err != nil {
			return errors.New("database::SetAuthors: " + err.Error())
		}
	}
	return
}

// DeleteAuthors deletes the author linked to a database
func (d *Database) DeleteAuthors(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"database__authors\" WHERE database_id=:id", d)
	if err != nil {
		err = errors.New("database::DeleteAuthors: " + err.Error())
	}
	return
}

// GetContextsList lists all user designed as context of a database
func (d *Database) GetContextsList(tx *sqlx.Tx) (contexts []Database_context, err error) {
	err = tx.Select(&contexts, "SELECT id, context FROM database_context WHERE database_id = $1", d.Id)
	if err != nil {
		err = errors.New("database::GetContextsList: " + err.Error())
	}
	return
}

// SetContexts links users as contexts to a database
func (d *Database) SetContexts(tx *sqlx.Tx, contexts []string) error {
	for _, cname := range contexts {
		_, err := tx.Exec("INSERT INTO \"database_context\" (database_id, context) VALUES ($1, $2)", d.Id, cname)
		if err != nil {
			return errors.New("database::SetContexts: " + err.Error())
		}
	}
	return nil
}

// DeleteContexts deletes the context linked to a database
func (d *Database) DeleteContexts(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"database_context\" WHERE database_id=:id", d)
	if err != nil {
		err = errors.New("database::DeleteContexts: " + err.Error())
	}
	return err
}

func (d *Database) SetTranslations(tx *sqlx.Tx, field string, translations []struct {
	Lang_Isocode string
	Text         string
}) (err error) {

	// Check if translation entry exists for this database and this lang

	var transID int

	for _, tr := range translations {
		err = tx.QueryRow("SELECT count(database_id) FROM database_tr WHERE database_id = $1 AND lang_isocode = $2", d.Id, tr.Lang_Isocode).Scan(&transID)
		if transID == 0 {
			_, err = tx.Exec("INSERT INTO database_tr (database_id, lang_isocode, description, geographical_limit, bibliography, context_description, source_description, source_relation, copyright, subject) VALUES ($1, $2, '', '', '', '', '', '', '', '')", d.Id, tr.Lang_Isocode)
			if err != nil {
				err = errors.New("database::SetTranslations: " + err.Error())
			}
		}
		if tr.Text != "" {
			_, err = tx.Exec("UPDATE database_tr SET "+field+" = $1 WHERE database_id = $2 and lang_isocode = $3", tr.Text, d.Id, tr.Lang_Isocode)
		}
	}

	if err != nil {
		err = errors.New("database::SetTranslations: " + err.Error())
	}
	return
}

// GetOwnerInfos get all informations about the owner of the database
func (d *Database) GetOwnerInfos(tx *sqlx.Tx) (owner DatabaseAuthor, err error) {
	err = tx.Get(owner, "SELECT * FROM \"user\" u LEFT JOIN \"database\" d ON u.id = d.owner WHERE d.id = $1", d.Id)
	if err != nil {
		err = errors.New("database::GetOwnerInfos: " + err.Error())
	}
	return
}

// GetImportList lists all informations about an import (date, filename, etc)
func (d *Database) GetImportList(tx *sqlx.Tx) (imports []ImportFullInfos, err error) {
	imports = []ImportFullInfos{}
	err = tx.Select(&imports, "SELECT i.*, u.firstname || ' ' ||  u.lastname AS fullname FROM import i LEFT JOIN \"user\" u ON i.user_id = u.id WHERE i.database_id = $1 ORDER BY id DESC", d.Id)
	if err != nil {
		err = errors.New("database::GetImportList: " + err.Error())
	}
	return
}

// GetLastImport lists last import informations
func (d *Database) GetLastImport(tx *sqlx.Tx) (imp Import, err error) {
	imp = Import{}
	err = tx.Get(&imp, "SELECT * FROM import WHERE i.jdatabase_id = $1 ORDER by id DESC LIMIT 1", d.Id)
	if err != nil {
		err = errors.New("database::GetLastImport: " + err.Error())
	}
	return
}

// UpdateFields updates "database" fields (crazy isn't it ?)
func (d *Database) UpdateFields(tx *sqlx.Tx, params interface{}, fields ...string) (err error) {
	var upd string
	for i, f := range fields {
		upd += "\"" + f + "\" = :" + f
		if i+1 < len(fields) {
			upd += ", "
		}
	}
	query := "UPDATE \"database\" SET " + upd + " WHERE id = :id"

	if err != nil {
		err = errors.New("database::UpdateFields: " + err.Error())
	}

	_, err = tx.NamedExec(query, params)

	return

}

// CacheGeom get database sites extend and cache enveloppe
func (d *Database) CacheGeom(tx *sqlx.Tx) (err error) {
	// Extent
	//_, err = tx.NamedExec("SELECT ST_Envelope(sites.geom::::geometry) FROM (SELECT geom FROM site WHERE database_id = :id) as sites", d)
	// Envelope
	_, err = tx.NamedExec("UPDATE database SET geographical_extent_geom = (SELECT (ST_Envelope((SELECT ST_Multi(ST_Collect(f.geom)) as singlegeom FROM (SELECT (ST_Dump(geom::::geometry)).geom As geom FROM site WHERE database_id = :id) As f)))) WHERE id = :id", d)
	if err != nil {
		err = errors.New("database::CacheGeom: " + err.Error())
	}
	return
}

// CacheDates get database sites extend and cache enveloppe
func (d *Database) CacheDates(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("UPDATE database SET start_date = (SELECT min(start_date1) FROM site_range WHERE site_id IN (SELECT id FROM site where database_id = :id) AND start_date1 != -2147483648), end_date = (SELECT max(end_date2) FROM site_range WHERE site_id IN (SELECT id FROM site where database_id = :id) AND end_date2 != 2147483647) WHERE id = :id", d)
	if err != nil {
		err = errors.New("database::CheckDates: " + err.Error())
	}
	return
}

// IsLinkedToProject returns true or false if database is linked or not to user project
func (d *Database) IsLinkedToProject(tx *sqlx.Tx, project_ID int) (linked bool, err error) {
	linked = false
	c := 0
	err = tx.Get(&c, "SELECT count(*) FROM project__databases WHERE project_id = $1 AND database_id = $2", project_ID, d.Id)
	if c > 0 {
		linked = true
	}
	return
}

// LinkToUserProject links database to project
func (d *Database) LinkToUserProject(tx *sqlx.Tx, project_ID int) (err error) {
	_, err = tx.Exec("INSERT INTO project__databases (project_id, database_id) VALUES ($1, $2)", project_ID, d.Id)
	return
}

// ExportCSV exports database and sites as as csv file
func (d *Database) ExportCSV(tx *sqlx.Tx) (outp string, err error) {

	var buff bytes.Buffer

	w := csv.NewWriter(&buff)
	w.Comma = ';'
	w.UseCRLF = true

	err = w.Write([]string{"SITE_SOURCE_ID", "SITE_NAME", "MAIN_CITY_NAME", "GEONAME_ID", "PROJECTION_SYSTEM", "LONGITUDE", "LATITUDE", "ALTITUDE", "CITY_CENTROID", "STATE_OF_KNOWLEDGE", "OCCUPATION", "STARTING_PERIOD", "ENDING_PERIOD", "CARAC_NAME", "CARAC_LVL1", "CARAC_LVL2", "CARAC_LVL3", "CARAC_LVL4", "CARAC_EXP", "BIBLIOGRAPHY", "COMMENTS"})
	if err != nil {
		log.Println("database::ExportCSV : ", err.Error())
	}
	w.Flush()

	// Datatabase isocode

	err = tx.Get(d, "SELECT name, default_language FROM \"database\" WHERE id = $1", d.Id)
	if err != nil {
		return
	}

	// Cache characs
	characs := make(map[int]string)

	q := "WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || ';' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC\n"

	rows, err := tx.Query(q, d.Default_language)
	switch {
	case err == sql.ErrNoRows:
		rows.Close()
		return outp, nil
	case err != nil:
		rows.Close()
		return
	}
	for rows.Next() {
		var id int
		var path string
		if err = rows.Scan(&id, &path); err != nil {
			return
		}
		characs[id] = path
	}

	q = "SELECT s.code, s.name, s.city_name, s.city_geonameid, ST_X(s.geom::geometry) as longitude, ST_Y(s.geom::geometry) as latitude, ST_X(s.geom_3d::geometry) as longitude_3d, ST_Y(s.geom_3d::geometry) as latitude3d, ST_Z(s.geom_3d::geometry) as altitude, s.centroid, s.occupation, sr.start_date1, sr.start_date2, sr.end_date1, sr.end_date2, src.exceptional, src.knowledge_type, srctr.bibliography, srctr.comment, c.id as charac_id FROM site s LEFT JOIN site_range sr ON s.id = sr.site_id LEFT JOIN site_tr str ON s.id = str.site_id LEFT JOIN site_range__charac src ON sr.id = src.site_range_id LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id LEFT JOIN charac c ON src.charac_id = c.id WHERE s.database_id = $1 AND str.lang_isocode IS NULL OR str.lang_isocode = $2 ORDER BY s.id, sr.id"

	rows2, err := tx.Query(q, d.Id, d.Default_language)
	if err != nil {
		rows2.Close()
		return
	}
	for rows2.Next() {
		var (
			code           string
			name           string
			city_name      string
			city_geonameid int
			longitude      float64
			latitude       float64
			longitude3d    float64
			latitude3d     float64
			altitude3d     float64
			centroid       bool
			occupation     string
			start_date1    int
			start_date2    int
			end_date1      int
			end_date2      int
			knowledge_type string
			exceptional    bool
			bibliography   string
			comment        string
			charac_id      int
			slongitude     string
			slatitude      string
			saltitude      string
			scentroid      string
			soccupation    string
			scharacs       string
			scharac_name   string
			scharac_lvl1   string
			scharac_lvl2   string
			scharac_lvl3   string
			scharac_lvl4   string
			sexceptional   string
			// description    string
		)
		if err = rows2.Scan(&code, &name, &city_name, &city_geonameid, &longitude, &latitude, &longitude3d, &latitude3d, &altitude3d, &centroid, &occupation, &start_date1, &start_date2, &end_date1, &end_date2, &exceptional, &knowledge_type, &bibliography, &comment, &charac_id); err != nil {
			log.Println(err)
			rows2.Close()
			return
		}
		// Geonameid
		var cgeonameid string
		if city_geonameid != 0 {
			cgeonameid = strconv.Itoa(city_geonameid)
		}
		// Longitude
		slongitude = strconv.FormatFloat(longitude, 'f', -1, 32)
		// Latitude
		slatitude = strconv.FormatFloat(latitude, 'f', -1, 32)
		// Altitude
		if longitude3d == 0 && latitude3d == 0 && altitude3d == 0 {
			saltitude = ""
		} else {
			saltitude = strconv.FormatFloat(altitude3d, 'f', -1, 32)
		}
		// Centroid
		if centroid {
			scentroid = translate.T(d.Default_language, "IMPORT.CSVFIELD_ALL.T_LABEL_YES")
		} else {
			scentroid = translate.T(d.Default_language, "IMPORT.CSVFIELD_ALL.T_LABEL_NO")
		}
		// Occupation
		switch occupation {
		case "not_documented":
			soccupation = translate.T(d.Default_language, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_NOT_DOCUMENTED")
		case "single":
			soccupation = translate.T(d.Default_language, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_SINGLE")
		case "continuous":
			soccupation = translate.T(d.Default_language, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_CONTINUOUS")
		case "multiple":
			soccupation = translate.T(d.Default_language, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_MULTIPLE")
		}
		// State of knowledge
		switch knowledge_type {
		case "not_documented":
			knowledge_type = translate.T(d.Default_language, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_NOT_DOCUMENTED")
		case "literature":
			knowledge_type = translate.T(d.Default_language, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_LITERATURE")
		case "prospected_aerial":
			knowledge_type = translate.T(d.Default_language, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_AERIAL")
		case "prospected_pedestrian":
			knowledge_type = translate.T(d.Default_language, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_PEDESTRIAN")
		case "surveyed":
			knowledge_type = translate.T(d.Default_language, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_SURVEYED")
		case "dig":
			knowledge_type = translate.T(d.Default_language, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_DIG")
		}
		// Revert hack on dates
		if start_date1 < 0 && start_date1 > math.MinInt32 {
			start_date1--
		}
		if start_date2 < 0 && start_date2 > math.MinInt32 {
			start_date2--
		}
		if end_date1 < 0 && end_date1 > math.MinInt32 {
			end_date1--
		}
		if end_date2 < 0 && end_date2 > math.MinInt32 {
			end_date2--
		}
		// Starting period
		startingPeriod := ""
		if start_date1 != math.MinInt32 {
			startingPeriod += strconv.Itoa(start_date1)
		}
		if start_date1 != math.MinInt32 && start_date2 != math.MaxInt32 && start_date1 != start_date2 {
			startingPeriod += ":"
		}
		if start_date2 != math.MaxInt32 && start_date1 != start_date2 {
			startingPeriod += strconv.Itoa(start_date2)
		}
		if startingPeriod == "" {
			startingPeriod = translate.T(d.Default_language, "IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED")
		}
		// Ending period
		endingPeriod := ""
		if end_date1 != math.MinInt32 {
			endingPeriod += strconv.Itoa(end_date1)
		}
		if end_date1 != math.MinInt32 && end_date2 != math.MaxInt32 && end_date1 != end_date2 {
			endingPeriod += ":"
		}
		if end_date2 != math.MaxInt32 && end_date1 != end_date2 {
			endingPeriod += strconv.Itoa(end_date2)
		}
		if endingPeriod == "" {
			endingPeriod = translate.T(d.Default_language, "IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED")
		}
		// Caracs
		var characPath = characs[charac_id]
		// fmt.Println(code, characPath)
		num := strings.Count(characPath, ";")
		if num < 4 {
			scharacs += characPath + strings.Repeat(";", 4-num)
		} else {
			scharacs = characPath
		}
		scharac_lvl2 = ""
		scharac_lvl3 = ""
		scharac_lvl4 = ""
		for i, c := range strings.Split(scharacs, ";") {
			// fmt.Println(i, c)
			switch i {
			case 0:
				scharac_name = c
			case 1:
				scharac_lvl1 = c
			case 2:
				scharac_lvl2 = c
			case 3:
				scharac_lvl3 = c
			case 4:
				scharac_lvl4 = c
			}

		}
		// fmt.Println(scharac_name, scharac_lvl1, scharac_lvl2, scharac_lvl3, scharac_lvl4)
		// fmt.Println(startingPeriod, endingPeriod)
		// Caracs exp
		if exceptional {
			sexceptional = translate.T(d.Default_language, "IMPORT.CSVFIELD_ALL.T_LABEL_YES")
		} else {
			sexceptional = translate.T(d.Default_language, "IMPORT.CSVFIELD_ALL.T_LABEL_NO")
		}

		line := []string{code, name, city_name, cgeonameid, "4326", slongitude, slatitude, saltitude, scentroid, knowledge_type, soccupation, startingPeriod, endingPeriod, scharac_name, scharac_lvl1, scharac_lvl2, scharac_lvl3, scharac_lvl4, sexceptional, bibliography, comment}

		err := w.Write(line)
		w.Flush()
		if err != nil {
			log.Println("database::ExportCSV : ", err.Error())
		}
	}

	return buff.String(), nil
}

// GetTranslations lists all translated fields from database
func (d *DatabaseFullInfos) GetTranslations(tx *sqlx.Tx) (err error) {
	tr := []Database_tr{}
	err = tx.Select(&tr, "SELECT * FROM database_tr WHERE database_id = $1", d.Id)
	if err != nil {
		return
	}
	d.Description = MapSqlTranslations(tr, "Lang_isocode", "Description")
	d.Geographical_limit = MapSqlTranslations(tr, "Lang_isocode", "Geographical_limit")
	d.Bibliography = MapSqlTranslations(tr, "Lang_isocode", "Bibliography")
	d.Context_description = MapSqlTranslations(tr, "Lang_isocode", "Context_description")
	d.Source_description = MapSqlTranslations(tr, "Lang_isocode", "Source_description")
	d.Source_relation = MapSqlTranslations(tr, "Lang_isocode", "Source_relation")
	d.Copyright = MapSqlTranslations(tr, "Lang_isocode", "Copyright")
	d.Subject = MapSqlTranslations(tr, "Lang_isocode", "Subject")
	return
}
