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
	"log"

	"github.com/jmoiron/sqlx"
)

/*
 * Chronology Object
 */

// Get the chronology from the database
func (u *Chronology) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"chronology\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the chronology by inserting it in the database
func (u *Chronology) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"chronology\" (" + Chronology_InsertStr + ") VALUES (" + Chronology_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Id, u)
}

// Update the chronology in the database
func (u *Chronology) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"chronology\" SET "+Chronology_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the chronology from the database
func (u *Chronology) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"chronology\" WHERE id=:id", u)
	return err
}

// Childs return Chronology childs
func (u *Chronology) Childs(tx *sqlx.Tx) ([]Chronology, error) {
	answer := []Chronology{}
	var q = "SELECT * FROM \"chronology\" WHERE parent_id=:id order by start_date,end_date,id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return answer, err
	}
	err = stmt.Select(&answer, u)
	stmt.Close()
	return answer, err
}

/*
 * Chronology_root Object
 */

// Get the chronology_root from the database
func (u *Chronology_root) Get(tx *sqlx.Tx) error {
	var q = "SELECT root_chronology_id, admin_group_id, author_user_id, \"credits\", \"active\", ST_AsGeojson(geom) as geom, cached_langs FROM \"chronology_root\" WHERE root_chronology_id=:root_chronology_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the chronology_root by inserting it in the database
func (u *Chronology_root) Create(tx *sqlx.Tx) error {
	//stmt, err := tx.PrepareNamed("INSERT INTO \"chronology_root\" (" + Chronology_root_InsertStr + ", root_chronology_id) VALUES (" + Chronology_root_InsertValuesStr + ", :root_chronology_id) RETURNING root_chronology_id")
	stmt, err := tx.PrepareNamed("INSERT INTO \"chronology_root\" (\"admin_group_id\", \"author_user_id\", \"credits\", \"active\", geom, root_chronology_id, cached_langs) VALUES (:admin_group_id, :author_user_id, :credits, :active, ST_GeomFromGeoJSON(:geom), :root_chronology_id, :cached_langs) RETURNING root_chronology_id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Root_chronology_id, u)
}

// Update the chronology_root in the database
func (u *Chronology_root) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"chronology_root\" SET \"admin_group_id\" = :admin_group_id, \"author_user_id\" = :author_user_id, \"credits\" = :credits, \"active\" = :active, \"geom\" = ST_GeomFromGeoJSON(:geom), cached_langs = :cached_langs WHERE root_chronology_id=:root_chronology_id", u)
	return err
}

// Delete the chronology_root from the database
func (u *Chronology_root) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"chronology_root\" WHERE root_chronology_id=:root_chronology_id", u)
	return err
}

/*
 * Chronology_tr Object
 */

// Get the chronology_tr from the database
func (u *Chronology_tr) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"chronology_tr\" WHERE chronology_id=:chronology_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the chronology_tr by inserting it in the database
func (u *Chronology_tr) Create(tx *sqlx.Tx) error {
	log.Println("saving : ", u)
	stmt, err := tx.PrepareNamed("INSERT INTO \"chronology_tr\" (" + Chronology_tr_InsertStr + ", chronology_id, lang_isocode) VALUES (" + Chronology_tr_InsertValuesStr + ", :chronology_id, :lang_isocode) RETURNING chronology_id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Chronology_id, u)
}

// Update the chronology_tr in the database
func (u *Chronology_tr) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"chronology_tr\" SET "+Chronology_tr_UpdateStr+" WHERE chronology_id=:chronology_id AND lang_isocode=:lang_isocode", u)
	return err
}

// Delete the chronology_tr from the database
func (u *Chronology_tr) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"chronology_tr\" WHERE chronology_id=:chronology_id", u)
	return err
}
