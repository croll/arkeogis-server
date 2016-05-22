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
	"database/sql"
	"fmt"
	"strconv"
	"time"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
)

// DatabaseAuthor stores essential informations about an author
type DatabaseAuthor struct {
	Id        string
	Firstname string
	Lastname  string
}

// CountryInfos store country info and translations
type CountryInfos struct {
	Country
	Country_tr
	Country_geonameid int            `json:"-"`
	Lang_id           int            `json:"-"`
	Created_at        time.Time      `json:"-"`
	Updated_at        time.Time      `json:"-"`
	Geom              sql.NullString `json:"-"`
}

// ContinentInfos store country info and translations
type ContinentInfos struct {
	Continent
	Continent_tr
	Continent_geonameid int            `json:"-"`
	Lang_id             int            `json:"-"`
	Created_at          time.Time      `json:"-"`
	Updated_at          time.Time      `json:"-"`
	Geom                sql.NullString `json:"-"`
}

// DatabaseFullInfos stores all informations about a database
type DatabaseFullInfos struct {
	Database
	Imports       []Import
	Countries     []CountryInfos     `json:"countries"`
	Continents    []ContinentInfos   `json:"continents"`
	Handles       []Database_handle  `json:"handles"`
	Authors       []DatabaseAuthor   `json:"authors"`
	Contexts      []Database_context `json:"contexts"`
	Translations  []Database_tr      `json:"translations"`
	NumberOfSites int                `json:"number_of_sites"`
	Owner_name    string             `json:"owner_name"`
}

// DoesExist check if database exist with a name and an owner
func (d *Database) DoesExist(tx *sqlx.Tx) (exists bool, err error) {
	exists = false
	err = tx.QueryRowx("SELECT id FROM \"database\" WHERE name = $1 AND owner = $2", d.Name, d.Owner).Scan(&d.Id)
	switch {
	case err == sql.ErrNoRows:
		return exists, nil
	case err != nil:
		return
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
		return
	}
	return true, nil
}

// Get retrieves informations about a database stored in the main table
func (d *Database) Get(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT * from \"database\" WHERE id=:id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(d, d)
}

// GetFullInfosRepresentation returns all informations about a database
func (d *Database) GetFullInfosRepresentation(tx *sqlx.Tx, langID int) (db DatabaseFullInfos, err error) {
	db = DatabaseFullInfos{}

	if d.Id == 0 {
		db.Imports = make([]Import, 0)
		db.Countries = make([]CountryInfos, 0)
		db.Continents = make([]ContinentInfos, 0)
		db.Handles = make([]Database_handle, 0)
		db.Authors = make([]DatabaseAuthor, 0)
		db.Contexts = make([]Database_context, 0)
		db.Translations = make([]Database_tr, 0)
		db.NumberOfSites = 0
		return
	}

	err = tx.Get(&db, "SELECT name, scale_resolution, geographical_extent, type, source_creation_date, owner, data_set, identifier, source, source_url, publisher, contributor, default_language, relation, coverage, copyright, state, license_id, subject, published, soft_deleted, d.created_at, d.updated_at, firstname || ' ' || lastname as owner_name FROM \"database\" d LEFT JOIN \"user\" u ON d.owner = u.id WHERE d.id = $1", d.Id)

	db.Authors, err = d.GetAuthorsList(tx)
	if err != nil {
		return
	}
	db.Countries, err = d.GetCountriesList(tx, langID)
	if err != nil {
		return
	}
	db.Continents, err = d.GetContinentsList(tx, langID)
	if err != nil {
		return
	}
	db.Imports, err = d.GetImportsList(tx)
	if err != nil {
		return
	}
	db.NumberOfSites, err = d.GetNumberOfSites(tx)
	if err != nil {
		return
	}
	db.Translations, err = d.GetTranslations(tx)
	if err != nil {
		return
	}
	return
}

// GetDbFullInfosAsJSON returns all infos about a database
func GetDbFullInfosAsJSON(databaseID, langID int) string {

	dbid := strconv.Itoa(databaseID)
	lid := strconv.Itoa(langID)

	var q = make([]string, 6)

	q[0] = db.AsJSON("SELECT name, scale_resolution, geographical_extent, type, source_creation_date, owner, data_set, identifier, source, source_url, publisher, contributor, default_language, relation, coverage, copyright, state, license_id, subject, published, soft_deleted, db.created_at, db.updated_at, firstname || ' ' || lastname as owner_name, (SELECT count(*) FROM site WHERE database_id = db.id) as number_of_sites FROM \"database\" db LEFT JOIN \"user\" u ON db.owner = u.id WHERE db.id = d.id", "database", true)

	q[1] = db.AsJSON("SELECT u.id, u.firstname, u.lastname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = d.id", "authors", true)

	q[2] = db.AsJSON("SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM country c LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid LEFT JOIN country_tr ct ON c.geonameid = ct.country_geonameid WHERE dc.database_id = d.id and ct.lang_id = "+lid, "countries", true)

	q[3] = db.AsJSON("SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM continent c LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid LEFT JOIN continent_tr ct ON c.geonameid = ct.continent_geonameid WHERE dc.database_id = d.id AND ct.lang_id = "+lid, "continents", true)

	q[4] = db.AsJSON("SELECT i.id, u.firstname, u.lastname, i.filename, i.created_at FROM import i LEFT JOIN \"user\" u ON i.user_id = u.id WHERE database_id = d.id", "last_import", true)

	q[5] = translate.GetQueryTranslationsAsJSON("database_tr", "database_id = d.id", "translations", true)

	// fmt.Println(q[0])
	// fmt.Println(q[1])
	// fmt.Println(q[2])
	// fmt.Println(q[3])
	// fmt.Println(q[4])
	// fmt.Println(q[5])

	fmt.Println(db.JSONQueryBuilder(q, "database d", "d.id = "+dbid))

	// test

	return ""

}

// Create insert the database into arkeogis db
func (d *Database) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"database\" (" + Database_InsertStr + ") VALUES (" + Database_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&d.Id, d)
}

// Update database informations
func (d *Database) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"database\" SET "+Database_UpdateStr+" WHERE id=:id", d)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSites deletes all sites linked to a database
func (d *Database) DeleteSites(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"site\" WHERE database_id=:id", d)
	return err
}

// GetCountriesList lists all countries linked to a database
func (d *Database) GetCountriesList(tx *sqlx.Tx, langID int) ([]CountryInfos, error) {
	countries := []CountryInfos{}
	err := tx.Select(&countries, "SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM country c LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid LEFT JOIN country_tr ct ON c.geonameid = ct.country_geonameid WHERE dc.database_id = $1 and ct.lang_id = $2", d.Id, langID)
	return countries, err
}

// AddCountries links countries to a database
func (d *Database) AddCountries(tx *sqlx.Tx, countryIds []int) error {
	for _, id := range countryIds {
		_, err := tx.Exec("INSERT INTO database__country (database_id, country_geonameid) VALUES ($1, $2)", d.Id, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteCountries unlinks countries to a database
func (d *Database) DeleteCountries(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"database__country\" WHERE database_id=:id", d)
	return err
}

// GetContinentsList lists all continents linked to a database
func (d *Database) GetContinentsList(tx *sqlx.Tx, langID int) ([]ContinentInfos, error) {
	continents := []ContinentInfos{}
	err := tx.Select(continents, "SELECT ct.name, c.geonameid, c.iso_code, c.geom FROM continent c LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid LEFT JOIN continent_tr ct ON c.geonameid = ct.continent_geonameid WHERE dc.database_id = $1 AND ct.lang_id = $2", d.Id, langID)
	return continents, err
}

// AddContinents links continents to a database
func (d *Database) AddContinents(tx *sqlx.Tx, continentIds []int) error {
	for _, id := range continentIds {
		_, err := tx.Exec("INSERT INTO database__continent (database_id, continent_geonameid) VALUES ($1, $2)", d.Id, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteContinents unlinks countries to a database
func (d *Database) DeleteContinents(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"database__continent\" WHERE database_id=:id", d)
	return err
}

// GetAuthorsList lists all user designed as author of a database
func (d *Database) GetAuthorsList(tx *sqlx.Tx) (authors []DatabaseAuthor, err error) {
	err = tx.Select(&authors, "SELECT u.id, u.firstname, u.lastname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = $1", d.Id)
	return
}

// SetAuthors links users as authors to a database
func (d *Database) SetAuthors(tx *sqlx.Tx, authors []int) error {
	for _, uid := range authors {
		_, err := tx.Exec("INSERT INTO \"database__authors\" (database_id, user_id) VALUES ($1, $2)", d.Id, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetOwnerInfos get all informations about the owner of the database
func (d *Database) GetOwnerInfos(tx *sqlx.Tx) (owner DatabaseAuthor, err error) {
	err = tx.Get(owner, "SELECT * FROM \"user\" u LEFT JOIN \"database\" d ON u.id = d.owner WHERE d.id = $1", d.Id)
	return
}

// DeleteAuthors deletes the author linked to a database
func (d *Database) DeleteAuthors(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"database__authors\" WHERE database_id=:id", d)
	return err
}

// GetHandlesList lists all handles linked to a database
func (d *Database) GetHandlesList(tx *sqlx.Tx) (handles []Database_handle, err error) {
	handles = []Database_handle{}
	err = tx.Select(&handles, "SELECT import_id, name, url, created_at FROM database_handle WHERE database_id = $1", d.Id)
	return
}

// GetImportsList lists all informations about an import (date, filename, etc)
func (d *Database) GetImportsList(tx *sqlx.Tx) (imports []Import, err error) {
	imports = []Import{}
	err = tx.Select(&imports, "SELECT i.id, u.firstname, u.lastname, i.filename, i.created_at FROM import i LEFT JOIN \"user\" u ON i.user_id = u.id WHERE database_id = $1", d.Id)
	return
}

// GetLastImport lists last import informations
func (d *Database) GetLastImport(tx *sqlx.Tx) (imp Import, err error) {
	imp = Import{}
	err = tx.Get(&imp, "SELECT id, filename FROM import WHERE database_id = $1 ORDER by id DESC LIMIT 1", d.Id)
	return
}

// GetTranslations lists all translated fields from database
func (d *Database) GetTranslations(tx *sqlx.Tx) (translations []Database_tr, err error) {
	translations = []Database_tr{}
	err = tx.Select(&translations, "SELECT * FROM database_tr WHERE database_id = $1", d.Id)
	return
}

// GetNumberOfSites returns an int which represent the number of sites linked to a database
func (d *Database) GetNumberOfSites(tx *sqlx.Tx) (nb int, err error) {
	err = tx.Get(&nb, "SELECT count(*) FROM site WHERE database_id = $1", d.Id)
	return
}
