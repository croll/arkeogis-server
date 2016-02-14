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
 * Company Object
 */

// Get the company from the database
func (u *Company) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"company\" WHERE "
	if len(u.Name) > 0 {
		q += "name=:name"
	} else {
		q += "id=:id"
	}
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the company by inserting it in the database
func (u *Company) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"company\" (" + Company_InsertStr + ") VALUES (" + Company_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Id, u)
}

// Update the company in the database
func (u *Company) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"company\" SET "+Company_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the company from the database
func (u *Company) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"company\" WHERE id=:id", u)
	return err
}
