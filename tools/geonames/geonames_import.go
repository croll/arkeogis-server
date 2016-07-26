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

package main

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
)

var (
	isoCodes             map[string]bool
	continentsByID       map[int]bool
	cachedcontinentsByID map[int]bool
	countries            map[string]int
	cachedCountries      map[string]int
	countriesByID        map[int]string
	citiesByID           map[int]bool
	cachedCitiesByID     map[int]bool
	acNum                int
)

const geomNowhere = "ST_GeometryFromText('POINT(2.5559 49.0083)', 4326)" // Paris (CDG)

func downloadFromURL(url string) error {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]

	if _, err := os.Stat(fileName); err == nil {
		return errors.New("File " + fileName + " already exists. Delete it if you want that file to be downloaded again.")
	}

	fmt.Println("Downloading", url, "to", fileName)

	output, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return err
	}

	fmt.Println("file", fileName, "downloaded. Size:", n, "bytes")
	return nil
}

func importFile(fileName string) error {
	fileExt := filepath.Ext(fileName)
	fileNameWithoutExt := strings.Replace(fileName, fileExt, "", 1)
	var rc io.Reader
	switch fileExt {
	case ".zip":
		fileFound := false
		z, err := zip.OpenReader(fileNameWithoutExt + ".zip")
		if err != nil {
			return err
		}
		defer z.Close()
		for _, f := range z.File {
			if m, _ := regexp.MatchString(f.Name, fileNameWithoutExt+".txt"); m {
				fileFound = true
				fmt.Println("Processing", fileNameWithoutExt+".txt")
				rc, err = f.Open()
				if err != nil {
					return err
				}
			}
		}
		if !fileFound {
			return errors.New("No usable file found in " + fileNameWithoutExt + ".zip")
		}
	case ".txt":
		var err error
		rc, err = os.Open(fileName)
		if err != nil {
			return err
		}
	default:
		return errors.New("File extension not recognized")
	}
	switch fileNameWithoutExt {
	case "iso-languagecodes":
		err := importLanguageCodes(rc)
		if err != nil {
			return err
		}
	case "allCountries":
		switch acNum {
		case 0:
			err := importcountries(rc)
			if err != nil {
				return err
			}
			acNum++
		case 1:
			err := importCities(rc)
			if err != nil {
				return err
			}
		}
	case "alternateNames":
		err := importAlternateNames(rc)
		if err != nil {
			return err
		}
	}
	return nil
}

// func cacheExistingLangs makes an index of existing langs in database
// func cacheExistingLangs() {
// 	rows, err := db.DB.Query("SELECT id, iso_code FROM lang")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	var id int
// 	var iso_code string
// 	for rows.Next() {
// 		if err := rows.Scan(&id, &iso_code); err != nil {
// 			log.Fatal(err)
// 		}
// 		Langs[strings.TrimSpace(iso_code)] = id
// 	}
// 	if err := rows.Err(); err != nil {
// 		log.Fatal(err)
// 	}
// }

// importLanguageCodes processes the iso-languages.txt file from geonames and imports it in database

func importLanguageCodes(rc io.Reader) error {
	// Check if database allready contains lang
	scan := bufio.NewScanner(rc)
	//cacheExistingLangs()
	var nbLang int
	err := db.DB.QueryRow("SELECT count(*) FROM lang").Scan(&nbLang)
	if err != nil {
		log.Fatal("Unable to fetch langs in database")
	}
	// If true, show a warning message and exit
	if nbLang > 1 {
		return errors.New("/!\\ Database already contains langs. As everything is linked to langs in arkeogis, feel free to destroys the db yourself")
	}
	if err2 := insertDefaultValues(); err2 != nil {
		log.Fatal("Error inserting default values", err2.Error())
	}
	// Langs map stores the langs and associates them to their id for further use.
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	// Store langs in database
	//stmt, err := tx.Prepare("INSERT INTO lang (isocode, active) VALUES ($1, false)")
	lineNum := 0
	if err != nil {
		fmt.Println("Error inserting lang.", err)
		return err
	}
	// Insert Lang D
	fmt.Println("Inserting lang D")
	tx.Exec("INSERT INTO lang (isocode, active) VALUES ('D', false)")

	// Process
	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// Skip first line
		if lineNum == 0 {
			lineNum++
			continue
		}
		// Skip blank lang iso code
		if strings.TrimSpace(s[2]) == "" {
			continue
		}
		isoCode := strings.TrimSpace(s[2])
		// var id int
		//if _, ok := Langs[s[2]]; !ok {
		tx.Exec("INSERT INTO lang (isocode, active) VALUES ($1, false)", isoCode)
		// if err = stmt.QueryRow(s[2]).Scan(&id); err != nil {
		//  fmt.Println("Error inserting lang.", err)
		//  tx.Rollback()
		//  return err
		// }
		// Langs[s[2]] = id
		//}
		isoCodes[isoCode] = true
	}

	// activate default langs
	tx.MustExec("update lang set active='t' where isocode in ('fr','en','es','de','eu')")

	// set default lang translations for theses langs
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('fr', 'fr', 'Français')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('fr', 'en', 'French')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('fr', 'de', 'Französisch')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('fr', 'es', 'Francés')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('fr', 'eu', 'Frantsesa')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('en', 'fr', 'Anglais')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('en', 'en', 'Enflish')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('en', 'de', 'Englisch')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('en', 'es', 'Inglés')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('en', 'eu', 'Ingelesa')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('de', 'fr', 'Allemand')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('de', 'en', 'German')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('de', 'de', 'Deutsch')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('de', 'es', 'Alemán')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('de', 'eu', 'Alemana')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('es', 'fr', 'Espagnol')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('es', 'en', 'Spanish')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('es', 'de', 'Spanisch')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('es', 'es', 'Español')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('es', 'eu', 'Española')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('eu', 'fr', 'Basque')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('eu', 'en', 'Basque')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('eu', 'de', 'Baskisch')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('eu', 'es', 'Vasco')")
	tx.MustExec("insert into lang_tr (lang_isocode, lang_isocode_tr, name) values ('eu', 'eu', 'Euskera')")

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	err = importContinents()
	if err != nil {
		return err
	}
	return nil
}

// func cacheExistingcountries makes an index of existing counries in database
func cacheExistingcountries() {
	rows, err := db.DB.Query("SELECT geonameid, iso_code FROM country")
	if err != nil {
		log.Fatal("Error caching existing countries", err)
	}
	defer rows.Close()

	var geonameid int
	var isoCode string
	for rows.Next() {
		if err := rows.Scan(&geonameid, &isoCode); err != nil {
			log.Fatal(err)
		}
		cachedCountries[isoCode] = geonameid
	}
	if err := rows.Err(); err != nil {
		log.Fatal("Error caching existing countries #2", err)
	}
}

// func importcountries parses the geoname file "countryInfo.txt" and populates it in the database
// If country already set in database it updates it
// Else it adds it
// Warning: No country is removed from database by this function !

func importcountries(rc io.Reader) error {
	scan := bufio.NewScanner(rc)
	// Make an index from existing countries in database
	cacheExistingcountries()
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	var stmtInsert1, stmtInsert2, stmtUpdate1, stmtUpdate2 *sql.Stmt
	var errStmt1, errStmt2 error
	if len(cachedCountries) > 1 {
		fmt.Println("- Updating existing countries")
		stmtUpdate1, errStmt1 = tx.Prepare("UPDATE country SET iso_code = $2 WHERE geonameid = $1")
		stmtUpdate2, errStmt2 = tx.Prepare("UPDATE country_tr SET lang_isocode = $2, name = $3, name_ascii = $4 WHERE country_geonameid = $1")
	} else {
		fmt.Println("- Inserting countries")
		stmtInsert1, errStmt1 = tx.Prepare("INSERT INTO country (geonameid, iso_code, created_at, updated_at) VALUES ($1, $2, $3, $4)")
		stmtInsert2, errStmt2 = tx.Prepare("INSERT INTO country_tr (country_geonameid, lang_isocode, name, name_ascii) VALUES ($1, $2, $3, $4)")
	}
	if errStmt1 != nil {
		return errStmt1
	}
	if errStmt2 != nil {
		return errStmt2
	}

	rgxpOnlyCities, _ := regexp.Compile("^PCLI")

	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// ignore line with comment
		if strings.Index(line, "#") == 0 {
			continue
		}
		// get only cities
		if !rgxpOnlyCities.MatchString(strings.TrimSpace(s[7])) {
			continue
		}
		if strings.TrimSpace(s[0]) == "" {
			fmt.Println("GeonameID not found:", s[0])
			continue
		}
		if strings.TrimSpace(s[1]) == "" {
			fmt.Println("Country name not found:", s[1])
			continue
		}
		if strings.TrimSpace(s[2]) == "" {
			fmt.Println("Country ASCII name not found:", s[1])
			continue
		}
		GeonameID, _ := strconv.Atoi(strings.TrimSpace(s[0]))
		name := strings.TrimSpace(s[1])
		nameASCII := strings.ToLower(strings.TrimSpace(s[2]))
		isoCode := strings.TrimSpace(s[8])

		if _, ok := cachedCountries[isoCode]; ok {
			if _, err := stmtUpdate1.Exec(GeonameID, isoCode); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtUpdate2.Exec(GeonameID, "D", name, nameASCII); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if _, err := stmtInsert1.Exec(GeonameID, isoCode, time.Now(), time.Now()); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtInsert2.Exec(GeonameID, "D", name, nameASCII); err != nil {
				tx.Rollback()
				return err
			}
			countriesByID[GeonameID] = isoCode
			countries[isoCode] = GeonameID
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// func cacheExistingCities makes an index of existing cities in database
func cacheExistingCities() {
	rows, err := db.DB.Query("SELECT geonameid FROM city")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var geonameID int
	for rows.Next() {
		if err := rows.Scan(&geonameID); err != nil {
			log.Fatal(err)
		}
		cachedCitiesByID[geonameID] = true
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

// func importCities parses the geoname file "allCountries.txt" and populates in the database
// If entry already set in database it updates it
// Else it adds it
// Warning: No data is removed from database by this function !
func importCities(rc io.Reader) error {
	scan := bufio.NewScanner(rc)
	cacheExistingCities()
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	var stmtInsert1, stmtInsert2, stmtUpdate1, stmtUpdate2 *sql.Stmt
	var errStmt1, errStmt2 error
	// var lon, lat float64
	//var err error
	if len(cachedCitiesByID) > 1 {
		fmt.Println("- Updating existing cities")
		stmtUpdate1, errStmt1 = tx.Prepare("UPDATE city SET country_geonameid = $2, geom_centroid = ST_GeomFromText($3, 4326), updated_at=now() WHERE geonameid = $1")
		stmtUpdate2, errStmt2 = tx.Prepare("UPDATE city_tr SET lang_isocode = $2, name = $3, name_ascii = $4 WHERE city_geonameid = $1")
	} else {
		fmt.Println("- Inserting cities")
		stmtInsert1, errStmt1 = tx.Prepare("INSERT INTO city (geonameid, country_geonameid, geom_centroid, created_at, updated_at) VALUES ($1, $2, ST_GeomFromText($3, 4326), now(), now())")
		stmtInsert2, errStmt2 = tx.Prepare("INSERT INTO city_tr (city_geonameid, lang_isocode, name, name_ascii) VALUES ($1, $2, $3, $4)")
	}
	if errStmt1 != nil {
		return errStmt1
	}
	if errStmt2 != nil {
		return errStmt2
	}

	rgxpOnlyCities, _ := regexp.Compile("^PPL.*")

	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// ignore line with comment
		if strings.Index(line, "#") == 0 {
			continue
		}
		// get only cities
		if !rgxpOnlyCities.MatchString(strings.TrimSpace(s[7])) {
			continue
		}
		if strings.TrimSpace(s[0]) == "" {
			fmt.Println("GeonameID not found")
		}
		if strings.TrimSpace(s[1]) == "" {
			fmt.Println("City name not found")
			continue
		}
		if strings.TrimSpace(s[2]) == "" {
			fmt.Println("City ASCII name not found")
			continue
		}
		if strings.TrimSpace(s[4]) == "" {
			fmt.Println("Latitude not Found")
			continue
		}
		if strings.TrimSpace(s[5]) == "" {
			fmt.Println("Longitude not Found")
			continue
		}

		lat, err := strconv.ParseFloat(strings.TrimSpace(s[4]), 64)
		if err != nil {
			fmt.Println("Unable to convert latitude to float")
		}
		lon, err := strconv.ParseFloat(strings.TrimSpace(s[5]), 64)
		if err != nil {
			fmt.Println("Unable to convert longitude to float")
		}

		GeonameID, _ := strconv.Atoi(strings.TrimSpace(s[0]))
		isoCode := strings.TrimSpace(s[8])
		name := strings.TrimSpace(s[1])
		nameASCII := strings.ToLower(strings.TrimSpace(s[2]))
		geom := fmt.Sprintf("POINT(%f %f)", lon, lat)

		if _, ok := countries[isoCode]; !ok {
			fmt.Println("City", name, "not inserted his country iso code \""+isoCode+"\" is not found.")
			continue
		}
		if _, ok := cachedCitiesByID[GeonameID]; ok {
			if _, err := stmtUpdate1.Exec(GeonameID, countries[isoCode], geom); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtUpdate2.Exec(GeonameID, "D", name, nameASCII); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if _, err := stmtInsert1.Exec(GeonameID, countries[isoCode], geom); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtInsert2.Exec(GeonameID, "D", name, nameASCII); err != nil {
				tx.Rollback()
				return err
			}
			citiesByID[GeonameID] = true
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// func cacheExistingContinents makes an index of existing continents in database
func cacheExistingContinents() {
	rows, err := db.DB.Query("SELECT geonameid FROM continent")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var geonameID int
	for rows.Next() {
		if err := rows.Scan(&geonameID); err != nil {
			log.Fatal(err)
		}
		cachedcontinentsByID[geonameID] = true
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

// func importContinents parses the geoname file "allCountries.txt" and populates in the database
// If entry already set in database it updates it
// Else it adds it
// Warning: No data is removed from database by this function !
func importContinents() error {
	cacheExistingContinents()
	geonamesContinents := map[int]map[string]string{
		6255146: map[string]string{
			"IsoCode": "AF",
			"Name":    "Africa",
		},
		6255147: map[string]string{
			"IsoCode": "AS",
			"Name":    "Asia",
		},
		6255148: map[string]string{
			"IsoCode": "EU",
			"Name":    "Europe",
		},
		6255149: map[string]string{
			"IsoCode": "NA",
			"Name":    "North America",
		},
		6255151: map[string]string{
			"IsoCode": "OC",
			"Name":    "Oceania",
		},
		6255150: map[string]string{
			"IsoCode": "SA",
			"Name":    "South America",
		},
		625515: map[string]string{
			"IsoCode": "AN",
			"Name":    "Antartica",
		},
	}
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	var stmtInsert1, stmtInsert2, stmtUpdate1, stmtUpdate2 *sql.Stmt
	var errStmt1, errStmt2 error
	if len(cachedcontinentsByID) > 1 {
		fmt.Println("- Updating existing continents")
		stmtUpdate1, errStmt1 = tx.Prepare("UPDATE continent SET iso_code = $2, updated_at = $3 WHERE geonameid = $1")
		stmtUpdate2, errStmt2 = tx.Prepare("UPDATE continent_tr SET lang_isocode = $2, name = $3, name_ascii = $4 WHERE continent_geonameid = $1")
	} else {
		fmt.Println("- Inserting continents")
		stmtInsert1, errStmt1 = tx.Prepare("INSERT INTO continent (geonameid, iso_code, created_at, updated_at) VALUES ($1, $2, $3, $4)")
		stmtInsert2, errStmt2 = tx.Prepare("INSERT INTO continent_tr (continent_geonameid, lang_isocode, name, name_ascii) VALUES ($1, $2, $3, $4)")
	}
	if errStmt1 != nil {
		return errStmt1
	}
	if errStmt2 != nil {
		return errStmt2
	}
	for GeonameID, infos := range geonamesContinents {
		if _, ok := cachedcontinentsByID[GeonameID]; ok {
			if _, err := stmtUpdate1.Exec(GeonameID, infos["IsoCode"], time.Now()); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtUpdate2.Exec(GeonameID, "D", infos["Name"], strings.ToLower(infos["Name"])); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if _, err := stmtInsert1.Exec(GeonameID, infos["IsoCode"], time.Now(), time.Now()); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtInsert2.Exec(GeonameID, "D", infos["Name"], strings.ToLower(infos["Name"])); err != nil {
				tx.Rollback()
				return err
			}
		}
		continentsByID[GeonameID] = true
	}
	/*
		for scan.Scan() {
			line := scan.Text()
			s := strings.Split(line, "\t")
			// ignore line with comment
			if strings.Index(line, "#") == 0 {
				continue
			}
			// get only continents
			rgxp, _ := regexp.Compile("^CONT.*")
			if !rgxp.MatchString(strings.TrimSpace(s[7])) {
				continue
			}
			if strings.TrimSpace(s[0]) == "" {
				fmt.Println("GeonameID not found:", s[0])
			}
			GeonameID, _ := strconv.Atoi(s[0])
			if strings.TrimSpace(s[1]) == "" {
				fmt.Println("Continent name not found:", s[1])
				continue
			}
			if strings.TrimSpace(s[2]) == "" {
				fmt.Println("ASCII name not found:", s[1])
				continue
			}
			name_ascii := strings.ToLower(s[2])
			if _, ok := continentsByID[GeonameID]; ok {
				if _, err := stmtUpdate1.Exec(GeonameID, s[8], time.Now()); err != nil {
					tx.Rollback()
					return err
				}
				if _, err := stmtUpdate2.Exec(GeonameID, Langs["D"], s[1], name_ascii); err != nil {
					tx.Rollback()
					return err
				}
			} else {
				if _, err := stmtInsert1.Exec(GeonameID, s[8], time.Now(), time.Now()); err != nil {
					tx.Rollback()
					return err
				}
				if _, err := stmtInsert2.Exec(GeonameID, Langs["D"], s[1], name_ascii); err != nil {
					tx.Rollback()
					return err
				}
			}
			continentsByID[GeonameID] = s[8]
			Continents[s[8]] = GeonameID
		}
	*/
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// func importAlternateNames parses the geoname file "alternateNames.zip" and populates in the database
// As nothing rely directly to alternates names, we first try do destroy all datas with lang = D (our lang_isocode for "D" language) to trash older datas

func importAlternateNames(rc io.Reader) error {
	scan := bufio.NewScanner(rc)
	alreadyProcessed := map[string]bool{}
	// Delete old entries wich are not tagged with "Undefined lang"
	for _, tablename := range []string{"continent_tr", "country_tr", "city_tr"} {
		if _, err := db.DB.Exec("DELETE FROM " + tablename + " WHERE lang_isocode != 'D'"); err != nil {
			return err
		}
	}
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	stmtContinent, err := tx.Prepare("INSERT INTO continent_tr (continent_geonameid, lang_isocode, name, name_ascii) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	stmtCountry, err := tx.Prepare("INSERT INTO country_tr (country_geonameid, lang_isocode, name, name_ascii) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	stmtCity, err := tx.Prepare("INSERT INTO city_tr (city_geonameid, lang_isocode, name, name_ascii) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	fmt.Println("- Inserting alternate names")
	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// Skip line with no geonameid
		if strings.TrimSpace(s[1]) == "" {
			continue
		}
		GeonameID, _ := strconv.Atoi(s[1])
		// Skip line with no lang defined or without lang code with len of 2
		isoCode := strings.TrimSpace(s[2])
		if isoCode == "" || len(isoCode) != 2 {
			continue
		}
		// Check if isocode exists
		if _, ok := isoCodes[isoCode]; !ok {
			fmt.Println("iso code", isoCode, "does not exist in our lang table.")
			continue
		}
		// Skip line with no name
		if strings.TrimSpace(s[3]) == "" {
			continue
		}
		preferred, _ := strconv.Atoi(strings.TrimSpace(s[4]))
		// Uniq code
		uniqCode := strings.TrimSpace(s[1]) + "_" + isoCode
		// Name
		name := strings.TrimSpace(s[3])

		if _, ok := citiesByID[GeonameID]; ok {
			if _, ok := alreadyProcessed[uniqCode]; ok {
				if preferred == 1 {
					_, err = tx.Exec("DELETE FROM city_tr WHERE city_geonameid = $1 AND lang_isocode = $2", GeonameID, isoCode)
					if err != nil {
						return err
					}
				} else {
					continue
				}
			}
			if _, err = stmtCity.Exec(GeonameID, isoCode, name, ""); err != nil {
				tx.Rollback()
				return err
			}
		} else if _, ok := countriesByID[GeonameID]; ok {
			if _, ok := alreadyProcessed[uniqCode]; ok {
				if preferred == 1 {
					_, err = tx.Exec("DELETE FROM country_tr WHERE country_geonameid = $1 AND lang_isocode = $2", GeonameID, isoCode)
					if err != nil {
						return err
					}
				} else {
					continue
				}
			}
			if _, err = stmtCountry.Exec(GeonameID, isoCode, name, ""); err != nil {
				tx.Rollback()
				return err
			}
		} else if _, ok := continentsByID[GeonameID]; ok {
			if _, ok := alreadyProcessed[uniqCode]; ok {
				if preferred == 1 {
					_, err = tx.Exec("DELETE FROM continent_tr WHERE continent_geonameid = $1 AND lang_isocode = $2", GeonameID, isoCode)
					if err != nil {
						return err
					}
				} else {
					continue
				}
			}
			if _, err = stmtContinent.Exec(GeonameID, isoCode, name, ""); err != nil {
				tx.Rollback()
				return err
			}
		}
		alreadyProcessed[uniqCode] = true
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Function insertDefaultValues() inserts datas considered as "default value" helping avoiding null values
func insertDefaultValues() error {
	// Insert "undefined" lang and populate cache index
	var err error
	if err != nil {
		fmt.Println("can't insert default lang D : ", err)
		return err
	}
	// Insert "undefined" continent and populate cache index
	var undefinedContinentID int
	err = db.DB.QueryRow("INSERT INTO continent (geonameid, iso_code, created_at, updated_at) VALUES (0, 'U', $1, $2) RETURNING geonameid", time.Now(), time.Now()).Scan(&undefinedContinentID)
	if err != nil {
		fmt.Println("can't insert default continent D : ", err)
		return err
	}
	// Insert "undefined" country and populate cache index
	var undefinedCountryID int
	err = db.DB.QueryRow("INSERT INTO country (geonameid, iso_code, created_at, updated_at) VALUES (0, 'U', $1, $2) RETURNING geonameid", time.Now(), time.Now()).Scan(&undefinedCountryID)
	if err != nil {
		fmt.Println("can't insert default country D : ", err)
		return err
	}
	countries["U"] = undefinedCountryID
	// Insert "undefined" city
	var undefinedCityID int
	err = db.DB.QueryRow("INSERT INTO city (geonameid, country_geonameid, geom_centroid, created_at, updated_at) VALUES (0, 0, " + geomNowhere + ", now(), now()) RETURNING geonameid").Scan(&undefinedCityID)
	if err != nil {
		fmt.Println("can't insert city : ", err)
		return err
	}
	citiesByID[undefinedCityID] = true
	return nil
}

// Function main() is responsible to list the files we want to process then call download and import functions
func main() {
	files := []string{"iso-languagecodes.txt", "allCountries.zip", "allCountries.zip", "alternateNames.zip"}
	//	files := []string{"alternateNames.zip"}
	//Langs = map[string]int{}
	isoCodes = map[string]bool{}
	continentsByID = map[int]bool{}
	cachedcontinentsByID = map[int]bool{}
	countries = map[string]int{}
	cachedCountries = map[string]int{}
	countriesByID = map[int]string{}
	cachedCitiesByID = map[int]bool{}
	citiesByID = map[int]bool{}
	l := len(files)
	for i := 0; i < l; i++ {
		url := "http://download.geonames.org/export/dump/" + files[i]
		if err := downloadFromURL(url); err != nil {
			fmt.Println(err)
		}
		if err := importFile(files[i]); err != nil {
			fmt.Println(err)
			return
		}
	}
}
