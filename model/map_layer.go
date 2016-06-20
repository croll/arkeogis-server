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
	stmt, err := tx.PrepareNamed("INSERT INTO \"map_layer\" (" + Map_layer_InsertStr + ") VALUES (" + Map_layer_InsertValuesStr + ") RETURNING id")
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

// Set publication state of the map layer
func (u *Map_layer) SetPublicationState(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"map_layer\" SET published = :published WHERE id=:id", u)
	return err
}
