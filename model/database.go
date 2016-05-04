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
	"github.com/jmoiron/sqlx"
	"fmt"
)

type DatabaseAuthor struct {
	id string
	firstname string
	lastname string
}

type DatabaseFullInfos struct {
	Database
	Database_tr
	Imports []Import
	Countries []Country `json: "countries"`
	Continents []Continent `json: "continents"`
	Handles []Database_handle `json: "handles"`
	Authors []DatabaseAuthor `json: "authors"`
	NumberOfSites int `json: "numberOfSites"`
	Owner_name string `json: "ownerName"`
}

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

func (d *Database) Get(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT * from \"database\" WHERE id=:id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(d, d)
}

func (d *Database) GetFullInfosRepresentation(tx *sqlx.Tx, langID int) (db DatabaseFullInfos, err error) {
	db = DatabaseFullInfos{}
	err = tx.Get(&db, "SELECT name, scale_resolution, geographical_extent, type, source_creation_date, data_set, identifier, source, source_url, publisher, contributor, default_language, relation, coverage, copyright, state, license_id, context, context_description, subject, published, soft_deleted, d.created_at, d.updated_at, firstname || ' ' || lastname as owner_name FROM \"database\" d LEFT JOIN \"user\" u ON d.owner = u.id WHERE d.id = $1", d.Id)
	if err != nil {
		return
	}
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
	fmt.Println(db)
	return
}

func (d *Database) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"database\" (" + Database_InsertStr + ") VALUES (" + Database_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&d.Id, d)
}

func (d *Database) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"database\" SET "+Database_UpdateStr+" WHERE id=:id", d)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) DeleteSites(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"site\" WHERE database_id=:id", d)
	return err
}

func (d *Database) GetCountriesList(tx *sqlx.Tx, langID int) ([]Country, error) {
	countries := []Country{}
	err := tx.Select(countries, "SELECT geonameid, iso_code, geom FROM country c LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid WHERE dc.database_id = $1", d.Id)
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

func (d *Database) GetContinentsList(tx *sqlx.Tx, langID int) ([]Continent, error) {
	continents := []Continent{}
	err := tx.Select(continents, "SELECT geonameid, iso_code, geom FROM continent c LEFT JOIN database__continentdc ON c.geonameid = dc.continent_geonameid WHERE dc.database_id = $1", d.Id)
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

func (d *Database) GetAuthorsList(tx *sqlx.Tx) (authors []DatabaseAuthor, err error) {
	err = tx.Select(&authors, "SELECT u.id, u.firstname, u.lastname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = $1", d.Id)
	return
}

func (d *Database) SetAuthors(tx *sqlx.Tx, authors []int) error {
	/*
		for _, uid := range authors {
				_, err := tx.In("INSERT INTO \"database__author\" database_id, user_id VALUES ", uid, d.Id)
				if err != nil {
					return err
				}
		}
	*/
	return nil
}

func (d *Database) DeleteAuthors(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"database__authors\" WHERE database_id=:id", d)
	return err
}

func (d *Database) GetHandlesList(tx *sqlx.Tx) (handles []Database_handle, err error) {
	handles = []Database_handle{}
	err = tx.Select(&handles, "SELECT import_id, name, url, created_at FROM database_handle WHERE database_id = $1", d.Id)
	return
}

func (d *Database) GetImportsList(tx *sqlx.Tx) (imports []Import, err error) {
	imports = []Import{}
	err = tx.Select(&imports, "SELECT i.id, u.firstname, u.lastname, i.filename, i.created_at FROM import i LEFT JOIN user u ON i.user_id = u.id WHERE database_id = $1", d.Id)
	return
}

func (d *Database) GetLastImport(tx *sqlx.Tx) (imp Import, err error) {
	imp = Import{}
	err = tx.Get(&imp, "SELECT id, filename FROM import WHERE database_id = $1 ORDER by id DESC LIMIT 1", d.Id)
	return
}

func (d *Database) GetNumberOfSites(tx *sqlx.Tx) (nb int, err error) {
	err = tx.Get(&nb, "SELECT count(*) FROM sites WHERE database_id = $1", d.Id)
	return
}
