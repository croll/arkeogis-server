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
	"fmt"

	"strings"

	db "github.com/croll/arkeogis-server/db"
)

func GetCharacPathsFromLang(name string, lang string) (caracs map[string]int, err error) {
	caracs = map[string]int{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.iso_code = $2 AND ca.id = (SELECT ca.id FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON lang.isocode = cat.lang_isocode WHERE lang.iso_code = $2 AND lower(cat.name) = lower($1)) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.iso_code = $2 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;", name, lang)
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
	fmt.Println(caracs)
	return
}

func GetCharacPathsFromLangID(name string, langIsocode string) (caracs map[string]int, err error) {
	caracs = map[string]int{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $2 AND ca.id = (SELECT ca.id FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON lang.isocode = cat.lang_isocode WHERE lang.isocode = $2 AND lower(cat.name) = lower($1)) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $2 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;", name, langIsocode)
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
		path = strings.ToLower(path)
		caracs[path] = id
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func GetAllCharacPathIDsFromLangIsocode(langIsocode string) (caracs map[int]string, err error) {
	caracs = map[int]string{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.charac_id::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path|| '->' || ca.id) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC", langIsocode)
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
		caracs[id] = path
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func GetAllCharacsRootFromLangIsocode(langIsocode string) (caracsRoot map[string]int, err error) {
	caracsRoot = map[string]int{}
	rows, err := db.DB.Query("SELECT isocode, name FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id WHERE ca.parent_id = 0 AND cat.lang_isocode = $1", langIsocode)
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

//WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 47 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 47 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC

//WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.charac_id::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 47 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path|| '->' || ca.id) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 47 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC

// SELECT s.code, s.name, sr.start_date1, sr.start_date2, sr.end_date1, sr.end_date2, src.exceptional, src.knowledge_type, srctr.comment, srctr.bibliography FROM site s LEFT JOIN site_range sr ON s.id = sr.site_id LEFT JOIN site_range__charac src ON sr.id = src.site_range_id LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id;
