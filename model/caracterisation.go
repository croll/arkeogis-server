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
	db "github.com/croll/arkeogis-server/db"
)

func GetCaracterisationPathsFromLang(name string, lang string) (caracs map[string]int, err error) {
	caracs = map[string]int{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM caracterisation AS ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON cat.lang_id = lang.id WHERE lang.iso_code = $2 AND ca.id = (SELECT ca.id FROM caracterisation ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON lang.id = cat.lang_id WHERE lang.iso_code = $2 AND cat.name = $1) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, caracterisation AS ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON cat.lang_id = lang.id WHERE lang.iso_code = $2 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;", name, lang)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		id   int
		path string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &path); err != nil {
			return
		}
		caracs[path] = id
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func GetCaracterisationPathsFromLangID(name string, langID int) (caracs map[string]int, err error) {
	caracs = map[string]int{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM caracterisation AS ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON cat.lang_id = lang.id WHERE lang.id= $2 AND ca.id = (SELECT ca.id FROM caracterisation ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON lang.id = cat.lang_id WHERE lang.id = $2 AND cat.name = $1) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, caracterisation AS ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON cat.lang_id = lang.id WHERE lang.id= $2 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;", name, langID)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		id   int
		path string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &path); err != nil {
			return
		}
		caracs[path] = id
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func GetAllCaracterisationsRootFromLangId(langId int) (caracsRoot map[string]int, err error) {
	caracsRoot = map[string]int{}
	rows, err := db.DB.Query("SELECT id, name FROM caracterisation ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id WHERE ca.parent_id = 0 AND cat.lang_id = $1", langId)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		id   int
		name string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			return
		}
		caracsRoot[name] = id
	}
	if err := rows.Err(); err != nil {
		return caracsRoot, err
	}
	return
}

//WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM caracterisation AS ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON cat.lang_id = lang.id WHERE lang.id = 48 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, caracterisation AS ca LEFT JOIN caracterisation_translation cat ON ca.id = cat.caracterisation_id LEFT JOIN lang ON cat.lang_id = lang.id WHERE lang.id = 48 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC
