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

package main

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	db "github.com/croll/arkeogis-server/db"
)

var caracsRootByLang map[string]map[string]string = map[string]map[string]string{
	"Furniture": map[string]string{
		"en": "Furniture",
		"fr": "Mobilier",
		"de": "Funde auswählen",
		"es": "Mobiliario",
	},
	"Landscape": map[string]string{
		"en": "Landscape",
		"fr": "Paysage",
		"de": "Umwelt auswählen",
		"es": "Paisaje",
	},
	"Production": map[string]string{
		"en": "Production",
		"fr": "Production",
		"de": "Production auswählen",
		"es": "Producción",
	},
	"Realestate": map[string]string{
		"en": "Realestate",
		"fr": "Immobilier",
		"de": "Befunde auswählen",
		"es": "Inmobiliario",
	},
}

func main() {

	// TODO
	// Delete all caracs entries without desroying everything
	if _, err := db.DB.Exec("TRUNCATE TABLE charac CASCADE"); err != nil {
		log.Fatalln(err)
	}
	// reset id
	if _, err := db.DB.Exec("SELECT pg_catalog.setval('charac_id_seq', 1, false)"); err != nil {
		log.Fatalln(err)
	}
	// insert default 0
	if _, err := db.DB.Exec("insert into charac values (0, 0, 0, 0, now(), now())"); err != nil {
		log.Fatalln(err)
	}

	for _, f := range []string{"../datas/csv/Furniture_fr-de-en-es.csv", "../datas/csv/Landscape_fr-de-en-es.csv", "../datas/csv/Production_fr-de-en-es.csv", "../datas/csv/Realestate_fr-de-en-es.csv"} {
		err := processFile(f)
		if err != nil {
			log.Println(err)
		}
	}
}

func processFile(filename string) error {

	log.Println("Importing file", filename)
	// Store current level
	langsByIso := [4]string{"fr", "de", "en", "es"}
	currentLevel := 0
	parentId := 0
	rootID := 0
	lastInsertId := 0
	parentsIDs := map[int]int{}
	// Parse csv file
	file, err := os.Open(filename)
	if err != nil {
		return errors.New("Unable to open csv file")
	}
	r := csv.NewReader(file)
	r.Comma = ';'

	lines, err := r.ReadAll()
	if err != nil {
		return errors.New("Unable to open csv file")
	}

	rootNames := strings.Split(filepath.Base(filename), "_")
	rootName := rootNames[0]

	if _, ok := caracsRootByLang[rootName]; !ok {
		log.Println("Unable to define charac root name analyzing file name")
		log.Fatalln("Please, give to your csv file a name like Furniture_fr-de-en-es.csv")
	}

	// Init db transaction
	tx := db.DB.MustBegin()

	// Insert root of charac with name derived from file name like Furniture_fr_de_en.csv
	err = tx.QueryRow("INSERT INTO charac (parent_id, \"order\", author_user_id, created_at, updated_at) VALUES ($1, $2, $3, now(), now()) RETURNING id", 0, 0, 0).Scan(&rootID)
	if err != nil {
		return err
	}
	// Insert root in charac_root
	_, err = tx.Exec("INSERT INTO charac_root values ($1, $2)", rootID, 0)
	if err != nil {
		return err
	}
	for _, langIsocode := range langsByIso {
		_, err = tx.Exec("INSERT INTO charac_tr (lang_isocode, charac_id, name, description) VALUES ($1, $2, $3, '')", langIsocode, rootID, caracsRootByLang[rootName][langIsocode])
		if err != nil {
			return err
		}
	}

	// For each line
	order := 0
	for _, line := range lines {
		// For each record
		for lvl, record := range line {
			r := strings.TrimSpace(record)
			// Ignore empty records
			if r == "" {
				continue
			}
			// If we have a level 0 carac unset parentId
			if lvl == 0 {
				parentId = rootID
				// Set parend id if child detected
			} else if lvl > currentLevel {
				parentId = lastInsertId
				// Store parent id for this lvl
				parentsIDs[lvl] = lastInsertId
			} else if lvl < currentLevel {
				parentId = parentsIDs[lvl]
			}
			currentLevel = lvl

			// reset order if level change
			if currentLevel != lvl {
				order = 0
			}

			// Split each record to get label for each lang
			for i, label := range strings.Split(record, "#") {
				l := strings.TrimSpace(label)
				// Insert charac once
				if i == 0 {
					err := tx.QueryRow("INSERT INTO charac (parent_id, \"order\", author_user_id, created_at, updated_at) VALUES ($1, $2, $3, now(), now()) RETURNING id", parentId, order, 0).Scan(&lastInsertId)
					if err != nil {
						return err
					}
				}
				_, err = tx.Exec("INSERT INTO charac_tr (lang_isocode, charac_id, name, description) VALUES ($1, $2, $3, '')", langsByIso[i], lastInsertId, l)
				if err != nil {
					return err
				}
			}
		}
		order++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
