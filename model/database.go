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
)

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

func (d *Database) GetInfos(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT * from \"database\" WHERE id=:id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&d, d)
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
	_, err := tx.NamedExec("UPDATE \"group\" SET "+Database_UpdateStr+" WHERE id=:id", d)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetAuthors(tx *sqlx.Tx) ([]int, error) {
	authors := []int{}
	rows, err := tx.Query("SELECT user_id FROM database__author WHERE database_id = $1", d.Id)
	if err != nil {
		return authors, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return authors, err
		}
		authors = append(authors, id)
	}
	if err := rows.Err(); err != nil {
		return authors, err
	}
	return authors, nil
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

func (d *Database) DeleteAuthor(tx *sql.Tx) error {
	return nil
}
