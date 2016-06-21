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

// Get the map layer from the database
func (u *Map_layer) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"map_layer\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create a map layer by inserting it in the database
func (u *Map_layer) Create(tx *sqlx.Tx) error {
	q := "INSERT INTO \"map_layer\" (\"creator_user_id\", \"type\", \"url\", \"identifier\", \"min_scale\", \"max_scale\", \"start_date\", \"end_date\", \"image_format\", \"geographical_extent_geom\", \"published\", \"license\", \"license_id\", \"max_usage_date\", \"created_at\", \"updated_at\") VALUES (:creator_user_id, :type, :url, :identifier, :min_scale, :max_scale, :start_date, :end_date, :image_format,  ST_GeomFromGeoJSON(:geographical_extent_geom), :published, :license, :license_id, :max_usage_date, now(), now()) RETURNING id"

	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Id, u)
}

// Update the map layer in the database
func (u *Map_layer) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"map_layer\" SET "+Map_layer_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the map layer from the database
func (u *Map_layer) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"map_layer\" WHERE id=:id", u)
	return err
}

// SetAuthors links users as authors to a wms layer
func (u *Map_layer) SetAuthors(tx *sqlx.Tx, authors []int) (err error) {
	for _, uid := range authors {
		_, err = tx.Exec("INSERT INTO \"map_layer__authors\" (map_layer_id, user_id) VALUES ($1, $2)", u.Id, uid)
		if err != nil {
			return
		}
	}
	return
}

// DeleteAuthors deletes the author linked to a wms layer
func (u *Map_layer) DeleteAuthors(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"map_layer__authors\" WHERE map_layer_id=:id", u)
	return
}

// Set publication state of the map layer
func (u *Map_layer) SetPublicationState(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"map_layer\" SET published = :published WHERE id=:id", u)
	return err
}

// SetTranslation set translations !
func (u *Map_layer) SetTranslations(tx *sqlx.Tx, field string, translations []struct {
	Lang_Isocode string
	Text         string
}) (err error) {

	// Check if translation entry exists for this map_layer and this lang

	var transID int

	for _, tr := range translations {
		err = tx.QueryRow("SELECT count(map_layer_id) FROM map_layer_tr WHERE map_layer_id = $1 AND lang_isocode = $2", u.Id, tr.Lang_Isocode).Scan(&transID)
		if transID == 0 {
			_, err = tx.Exec("INSERT INTO map_layer_tr (map_layer_id, lang_isocode, name, attribution, copyright, description) VALUES ($1, $2, '', '', '', '')", u.Id, tr.Lang_Isocode)
			if err != nil {
				return
			}
		}
		if tr.Text != "" {
			_, err = tx.Exec("UPDATE map_layer_tr SET "+field+" = $1 WHERE map_layer_id = $2 and lang_isocode = $3", tr.Text, u.Id, tr.Lang_Isocode)
		}
	}

	return
}
