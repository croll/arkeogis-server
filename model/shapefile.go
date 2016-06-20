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

import "github.com/jmoiron/sqlx"

// Get the shapefile from the database
func (u *Shapefile) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"shapefile\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create a shapefile by inserting it in the database
func (u *Shapefile) Create(tx *sqlx.Tx) error {
	q := "INSERT INTO \"shapefile\" (\"creator_user_id\", \"filename\", \"md5sum\", \"geom\", \"geojson\", \"start_date\", \"end_date\", \"geographical_extent_geom\", \"published\", \"license\", \"license_id\", \"declared_creation_date\", \"created_at\", \"updated_at\") VALUES (:creator_user_id, :filename, :md5sum, ST_GeomFromGeoJSON(':geojson'), :geojson, :start_date, :end_date, ST_GeomFromGeoJSON(':geographical_extent_geom'), :published, :license, :license_id, :declared_creation_date, now(), now()) RETURNING id"
	// fmt.Println(q)
	_, err := tx.NamedExec(q, u)
	return err
}

// Update the shapefile in the database
func (u *Shapefile) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"shapefile\" SET "+Shapefile_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the shapefile from the database
func (u *Shapefile) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"shapefile\" WHERE id=:id", u)
	return err
}

// Set publication state of the shapefile
func (u *Shapefile) SetPublicationState(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"shapefile\" SET published = :published WHERE id=:id", u)
	return err
}
