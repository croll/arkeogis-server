/* ArkeoGIS - The Arkeolog Geographical Information Server Program
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
	//	"database/sql"
	//	db "github.com/croll/arkeogis-server/db"
	"github.com/jmoiron/sqlx"
)

func (s *Site) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"site\" (code, name, city_geonameid, geom, centroid, occupation, database_id, created_at, updated_at) VALUES (:code, :name, :city_geonameid, :geom, :centroid, :occupation, :database_id, :now(), now()) RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&s.Id, s)
}

func (s *Site) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"site\" SET code=:code, name=:name, city_geonameid=:city_geonameid, geom=:geom, centroid=:centroid, occupation=:occupation, database_id=:database_id, updated_at=:updated_at", s)
	if err != nil {
		return err
	}
	return nil
}

func (s *Site) AddSiteRange(tx *sqlx.Tx) error {
	return nil
}
