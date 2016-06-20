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

// Get the shapefile from the shapefile
func (u *Shapefile) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"shapefile\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create a shapefile by inserting it in the shapefile
func (u *Shapefile) Create(tx *sqlx.Tx) error {
	q := "INSERT INTO \"shapefile\" (\"creator_user_id\", \"filename\", \"md5sum\", \"geojson_with_data\", \"geojson\", \"start_date\", \"end_date\", \"geographical_extent_geom\", \"published\", \"license\", \"license_id\", \"declared_creation_date\", \"created_at\", \"updated_at\") VALUES (:creator_user_id, :filename, :md5sum, :geojson_with_data, :geojson, :start_date, :end_date, ST_GeomFromGeoJSON(:geographical_extent_geom), :published, :license, :license_id, :declared_creation_date, now(), now()) RETURNING id"
	// fmt.Println(q)
	stmt, err := tx.PrepareNamed(q)
	defer stmt.Close()
	if err != nil {
		return err
	}
	return stmt.Get(&u.Id, u)
}

// Update the shapefile in the shapefile
func (u *Shapefile) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"shapefile\" SET "+Shapefile_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the shapefile from the shapefile
func (u *Shapefile) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"shapefile\" WHERE id=:id", u)
	return err
}

// Set publication state of the shapefile
func (u *Shapefile) SetPublicationState(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"shapefile\" SET published = :published WHERE id=:id", u)
	return err
}

// SetAuthors links users as authors to a shapefile
func (u *Shapefile) SetAuthors(tx *sqlx.Tx, authors []int) (err error) {
	for _, uid := range authors {
		_, err = tx.Exec("INSERT INTO \"shapefile_authors\" (shapefile_id, user_id) VALUES ($1, $2)", u.Id, uid)
		if err != nil {
			return
		}
	}
	return
}

// DeleteAuthors deletes the author linked to a shapefile
func (u *Shapefile) DeleteAuthors(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"shapefile_authors\" WHERE shapefile_id=:id", u)
	return
}

// SetTranslation set translations !
func (u *Shapefile) SetTranslations(tx *sqlx.Tx, field string, translations []struct {
	Lang_Isocode string
	Text         string
}) (err error) {

	// Check if translation entry exists for this shapefile and this lang

	var transID int

	for _, tr := range translations {
		err = tx.QueryRow("SELECT count(shapefile_id) FROM shapefile_tr WHERE shapefile_id = $1 AND lang_isocode = $2", u.Id, tr.Lang_Isocode).Scan(&transID)
		if transID == 0 {
			_, err = tx.Exec("INSERT INTO shapefile_tr (shapefile_id, lang_isocode, name, attribution, copyright, description) VALUES ($1, $2, '', '', '', '')", u.Id, tr.Lang_Isocode)
			if err != nil {
				return
			}
		}
		if tr.Text != "" {
			_, err = tx.Exec("UPDATE shapefile_tr SET "+field+" = $1 WHERE shapefile_id = $2 and lang_isocode = $3", tr.Text, u.Id, tr.Lang_Isocode)
		}
	}

	return
}
