/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2016 CROLL SAS
 *
 * Authors :
 *  Nicolas Dimitrijevic <nicolas@croll.fr>
 *  Christophe Beveraggi <beve@croll.fr>
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

import "github.com/jmoiron/sqlx"

/*
 * City Object
 */

// Get the city from the database
func (u *City) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"city\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the city by inserting it in the database
func (u *City) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"city\" (" + City_InsertStr + ") VALUES (" + City_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Geonameid, u)
}

// Update the city in the database
func (u *City) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"city\" SET "+City_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the city from the database
func (u *City) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"city\" WHERE id=:id", u)
	return err
}

/*
 * model.City_wtr
 */

type City_wtr struct {
	City
	City_tr
}

// Get the city from the database
func (u *City_wtr) Get(tx *sqlx.Tx) error {
	var q = "SELECT city.*, city_tr.* FROM \"city\" LEFT JOIN \"city_tr\" ON city_geonameid = geonameid WHERE geonameid=:geonameid AND (lang_id=:lang_id OR lang_id=0) GROUP BY city.geonameid, city_tr.city_geonameid, city_tr.lang_id ORDER BY city_tr.lang_id DESC LIMIT 1"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

/*
 * Country Object
 */

// Get the country from the database
func (u *Country) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"country\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the country by inserting it in the database
func (u *Country) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"country\" (" + Country_InsertStr + ") VALUES (" + Country_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Geonameid, u)
}

// Update the country in the database
func (u *Country) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"country\" SET "+Country_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the country from the database
func (u *Country) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"country\" WHERE id=:id", u)
	return err
}

/*
 * model.Country_wtr
 */

type Country_wtr struct {
	Country
	Country_tr
}

// Get the country from the database
func (u *Country_wtr) Get(tx *sqlx.Tx) error {
	var q = "SELECT country.*, country_tr.* FROM \"country\" LEFT JOIN \"country_tr\" ON country_geonameid = geonameid WHERE geonameid=:geonameid AND (lang_id=:lang_id OR lang_id=0) GROUP BY country.geonameid, country_tr.country_geonameid, country_tr.lang_id ORDER BY country_tr.lang_id DESC LIMIT 1"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

type CityAndCountry_wtr struct {
	City    City_wtr    `json:"city"`
	Country Country_wtr `json:"country"`
}

func (u *CityAndCountry_wtr) Get(tx *sqlx.Tx, cityId int, langIsocode string) error {
	u.City.Geonameid = cityId
	u.City.Lang_isocode = langIsocode
	err := u.City.Get(tx)
	if err != nil {
		return err
	}
	//log.Println("City : ", u.City)

	u.Country.Lang_isocode = langIsocode
	u.Country.Geonameid = u.City.Country_geonameid
	err = u.Country.Get(tx)
	if err != nil {
		return err
	}
	//log.Println("Country : ", u.Country)
	return nil
}
