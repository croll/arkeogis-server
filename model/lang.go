/* ArkeoGIS - The Arkeolog Geographical Information Server Program
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
)

type LangT struct {
	Id      int    `json:"id"`
	IsoCode string `json:"iso_code"`
}

func GetActiveLangs() ([]LangT, error) {

	langs := []LangT{}
	err := db.DB.Select(&langs, "SELECT id, iso_code AS isocode FROM lang WHERE active = true AND iso_code != 'D'")
	if err != nil {
		return langs, err
	}
	return langs, nil
}
