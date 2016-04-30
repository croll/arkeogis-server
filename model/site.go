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

import (
"github.com/jmoiron/sqlx"
)


func (s *Site) Get(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT * from \"site\" WHERE id=:id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(s, s)
}

func (s *Site) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"site\" (" + Site_InsertStr + ") VALUES (" + Site_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&s.Id, s)
}

func (s *Site) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"site\" SET "+Site_UpdateStr+" WHERE id=:id", s)
	if err != nil {
		return err
	}
	return nil
}

func (sr *Site_range) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"site_range\" (" + Site_range_InsertStr + ") VALUES (" + Site_range_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&sr.Id, sr)
}
