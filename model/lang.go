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
	db "github.com/croll/arkeogis-server/db"
	"github.com/jmoiron/sqlx"
)

// Get the lang from the database
func (l *Lang) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"lang\" WHERE isocode=:isocode"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(l, l)
}

// GetActiveLangs returns an array of Lang object which are actives only
func GetActiveLangs() ([]Lang, error) {

	langs := []Lang{}
	err := db.DB.Select(&langs, "SELECT isocode FROM lang WHERE active = true AND isocode != 'D'")
	if err != nil {
		return langs, err
	}
	return langs, nil
}

// GetLangs returns an array of Lang objects
func GetLangs() ([]Lang, error) {

	langs := []Lang{}
	err := db.DB.Select(&langs, "SELECT isocode FROM lang WHERE isocode != 'D'")
	if err != nil {
		return langs, err
	}
	return langs, nil
}
