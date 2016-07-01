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
	"errors"
	"log"

	"strings"

	db "github.com/croll/arkeogis-server/db"
	"github.com/jmoiron/sqlx"
)

/*
 * Charac Object
 */

// Get the charac from the database
func (u *Charac) Get(tx *sqlx.Tx) (err error) {
	var q = "SELECT * FROM \"charac\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return errors.New("model.charac::Get " + err.Error())
	}
	defer stmt.Close()
	err = stmt.Get(u, u)
	if err != nil {
		err = errors.New("model.charac::Get " + err.Error())
	}
	return
}

// Create the charac by inserting it in the database
func (u *Charac) Create(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("INSERT INTO \"charac\" (" + Charac_InsertStr + ") VALUES (" + Charac_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return errors.New("model.charac::Create " + err.Error())
	}
	defer stmt.Close()
	err = stmt.Get(&u.Id, u)
	if err != nil {
		err = errors.New("model.charac::Create " + err.Error())
	}
	return
}

// Update the charac in the database
func (u *Charac) Update(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("UPDATE \"charac\" SET "+Charac_UpdateStr+" WHERE id=:id", u)
	if err != nil {
		err = errors.New("model.charac::Update " + err.Error())
	}
	return
}

// Delete the charac from the database
func (u *Charac) Delete(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"charac\" WHERE id=:id", u)
	if err != nil {
		err = errors.New("model.charac::Delete " + err.Error())
	}
	return
}

// Childs return Charac childs
func (u *Charac) Childs(tx *sqlx.Tx) (answer []Charac, err error) {
	answer = []Charac{}
	var q = "SELECT * FROM \"charac\" WHERE parent_id=:id order by \"order\""
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		if err != nil {
			err = errors.New("model.charac::Childs" + err.Error())
		}
		return
	}
	err = stmt.Select(&answer, u)
	if err != nil {
		err = errors.New("model.charac::Childs" + err.Error())
	}
	stmt.Close()
	return answer, err
}

/*
 * Charac_root Object
 */

// Get the charac_root from the database
func (u *Charac_root) Get(tx *sqlx.Tx) (err error) {
	var q = "SELECT root_charac_id, admin_group_id FROM \"charac_root\" WHERE root_charac_id=:root_charac_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		if err != nil {
			err = errors.New("model.charac_root::Get" + err.Error())
		}
		return
	}
	defer stmt.Close()
	err = stmt.Get(u, u)
	if err != nil {
		err = errors.New("model.charac_root::Get" + err.Error())
	}
	return
}

// Create the charac_root by inserting it in the database
func (u *Charac_root) Create(tx *sqlx.Tx) (err error) {
	//stmt, err := tx.PrepareNamed("INSERT INTO \"charac_root\" (" + Charac_root_InsertStr + ", root_charac_id) VALUES (" + Charac_root_InsertValuesStr + ", :root_charac_id) RETURNING root_charac_id")
	stmt, err := tx.PrepareNamed("INSERT INTO \"charac_root\" (\"admin_group_id\", root_charac_id) VALUES (:admin_group_id, :root_charac_id) RETURNING root_charac_id")
	if err != nil {
		if err != nil {
			err = errors.New("model.charac_root::Create" + err.Error())
		}
	}
	defer stmt.Close()
	err = stmt.Get(&u.Root_charac_id, u)
	if err != nil {
		err = errors.New("model.charac_root::Create" + err.Error())
	}
	return
}

// Update the charac_root in the database
func (u *Charac_root) Update(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("UPDATE \"charac_root\" SET \"admin_group_id\" = :admin_group_id WHERE root_charac_id=:root_charac_id", u)
	if err != nil {
		err = errors.New("model.charac_root::Update" + err.Error())
	}
	return
}

// Delete the charac_root from the database
func (u *Charac_root) Delete(tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec("DELETE FROM \"charac_root\" WHERE root_charac_id=:root_charac_id", u)
	if err != nil {
		err = errors.New("model.charac_root::Delete" + err.Error())
	}
	return
}

/*
 * Project_hidden_characs Object
 */

// List return Project_hidden_characs of a Project
func (u *Project_hidden_characs) List(tx *sqlx.Tx) ([]Project_hidden_characs, error) {
	answer := []Project_hidden_characs{}
	var q = "SELECT * FROM \"project_hidden_characs\" WHERE project_id=:project_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return answer, err
	}
	err = stmt.Select(&answer, u)
	stmt.Close()
	return answer, err
}

// Get the project_hidden_characs from the database
func (u *Project_hidden_characs) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"project_hidden_characs\" WHERE charac_id=:charac_id AND project_id=:project_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the project_hidden_characs by inserting it in the database
func (u *Project_hidden_characs) Create(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("INSERT INTO \"project_hidden_characs\" (\"project_id\", charac_id) VALUES (:project_id, :charac_id)", u)
	return err
}

// Delete the project_hidden_characs from the database
func (u *Project_hidden_characs) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"project_hidden_characs\" WHERE charac_id=:charac_id AND project_id=:project_id", u)
	return err
}

/*
 * Charac_tr Object
 */

// Get the charac_tr from the database
func (u *Charac_tr) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"charac_tr\" WHERE charac_id=:charac_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the charac_tr by inserting it in the database
func (u *Charac_tr) Create(tx *sqlx.Tx) error {
	log.Println("saving : ", u)
	stmt, err := tx.PrepareNamed("INSERT INTO \"charac_tr\" (" + Charac_tr_InsertStr + ", charac_id, lang_isocode) VALUES (" + Charac_tr_InsertValuesStr + ", :charac_id, :lang_isocode) RETURNING charac_id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Charac_id, u)
}

// Update the charac_tr in the database
func (u *Charac_tr) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"charac_tr\" SET "+Charac_tr_UpdateStr+" WHERE charac_id=:charac_id AND lang_isocode=:lang_isocode", u)
	return err
}

// Delete the charac_tr from the database
func (u *Charac_tr) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"charac_tr\" WHERE charac_id=:charac_id", u)
	return err
}

/*
 * some utils on characs
 */

func GetCharacPathsFromLang(name string, lang string) (caracs map[string]int, err error) {
	caracs = map[string]int{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.iso_code = $2 AND ca.id = (SELECT ca.id FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON lang.isocode = cat.lang_isocode WHERE lang.iso_code = $2 AND lower(cat.name) = lower($1)) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.iso_code = $2 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;", name, lang)
	if err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetCharacPathsFromLang " + err.Error())
		}
		return
	}
	defer rows.Close()
	var (
		id   int
		path string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &path); err != nil {
			if err != nil {
				err = errors.New("model.charac_root::GetCharacPathsFromLang " + err.Error())
			}
			return
		}
		caracs[path] = id
	}
	if err = rows.Err(); err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetCharacPathsFromLang " + err.Error())
		}
		return
	}
	return
}

func GetCharacPathsFromLangID(name string, langIsocode string) (caracs map[string]int, err error) {
	caracs = map[string]int{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $2 AND ca.id = (SELECT ca.id FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON lang.isocode = cat.lang_isocode WHERE lang.isocode = $2 AND lower(cat.name) = lower($1) AND ca.parent_id = 0) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $2 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;", name, langIsocode)
	if err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetCharacPathsFromLangID " + err.Error())
		}
		return
	}
	defer rows.Close()
	var (
		id   int
		path string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &path); err != nil {
			if err != nil {
				err = errors.New("model.charac_root::GetCharacPathsFromLangID " + err.Error())
			}
			return
		}
		path = strings.ToLower(path)
		caracs[path] = id
	}
	if err = rows.Err(); err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetCharacPathsFromLangID " + err.Error())
		}
		return
	}
	return
}

func GetAllCharacPathIDsFromLangIsocode(langIsocode string) (caracs map[int]string, err error) {
	caracs = map[int]string{}
	rows, err := db.DB.Query("WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.charac_id::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path|| '->' || ca.id) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC", langIsocode)
	if err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetAllCharacPathIDsFromLangIsocode " + err.Error())
		}
		return
	}
	defer rows.Close()
	var (
		id   int
		path string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &path); err != nil {
			if err != nil {
				err = errors.New("model.charac_root::GetAllCharacPathIDsFromLangIsocode " + err.Error())
			}
			return
		}
		caracs[id] = path
	}
	if err = rows.Err(); err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetAllCharacPathIDsFromLangIsocode " + err.Error())
		}
		return
	}
	return
}

func GetAllCharacsRootFromLangIsocode(langIsocode string) (caracsRoot map[string]int, err error) {
	caracsRoot = map[string]int{}
	rows, err := db.DB.Query("SELECT id, name FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id WHERE ca.parent_id = 0 AND cat.lang_isocode = $1", langIsocode)
	if err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetAllCharacsRootFromLangIsocode " + err.Error())
		}
		return
	}
	defer rows.Close()
	var (
		id   int
		name string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			if err != nil {
				err = errors.New("model.charac_root::GetAllCharacsRootFromLangIsocode " + err.Error())
			}
			return
		}
		caracsRoot[name] = id
	}
	if err := rows.Err(); err != nil {
		if err != nil {
			err = errors.New("model.charac_root::GetAllCharacsRootFromLangIsocode " + err.Error())
		}
		return caracsRoot, err
	}
	return
}

//WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 'fr' AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 'fr' AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC

//WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.charac_id::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 47 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path|| '->' || ca.id) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 47 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC

// SELECT s.code, s.name, sr.start_date1, sr.start_date2, sr.end_date1, sr.end_date2, src.exceptional, src.knowledge_type, srctr.comment, srctr.bibliography FROM site s LEFT JOIN site_range sr ON s.id = sr.site_id LEFT JOIN site_range__charac src ON sr.id = src.site_range_id LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id;

//WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 'en' AND ca.id = (SELECT ca.id FROM charac ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON lang.isocode = cat.lang_isocode WHERE lang.isocode = 'en' AND lower(cat.name) = lower('Furniture') AND ca.parent_id = 0) UNION ALL SELECT ca.id, (p.path || '->' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = 'en' AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;
