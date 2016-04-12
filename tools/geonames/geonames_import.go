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
	Langs                map[string]int
	ContinentsById       map[int]bool
	CachedContinentsById map[int]bool
	Countries            map[string]int
	CachedCountries      map[string]int
	CountriesById        map[int]string
	CitiesById           map[int]bool
	CachedCitiesById     map[int]bool
	acNum                int
)

const geom_nowhere = "ST_GeometryFromText('POINT(2.5559 49.0083)', 4326)" // Paris (CDG)

func downloadFromUrl(url string) error {
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
	var err error
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
		rc, err = os.Open(fileName)
		if err != nil {
			return err
		}
	default:
		return errors.New("File extension not recognized")
	}
	switch fileNameWithoutExt {
	case "iso-languagecodes":
		err = importLanguageCodes(rc)
	case "allCountries":
		switch acNum {
		case 0:
			err = importCountries(rc)
			if err != nil {
				return err
			}
			acNum++
		case 1:
			err = importCities(rc)
			if err != nil {
				return err
			}
		}
	case "alternateNames":
		err = importAlternateNames(rc)
	}
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

// func cacheExistingLangs makes an index of existing langs in database
func cacheExistingLangs() {
	rows, err := db.DB.Query("SELECT id, iso_code FROM lang")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var id int
	var iso_code string
	for rows.Next() {
		if err := rows.Scan(&id, &iso_code); err != nil {
			log.Fatal(err)
		}
		Langs[strings.TrimSpace(iso_code)] = id
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

// importLanguageCodes processes the iso-languages.txt file from geonames and imports it in database

func importLanguageCodes(rc io.Reader) error {
	// Check if database allready contains lang
	scan := bufio.NewScanner(rc)
	cacheExistingLangs()
	var nbLang int
	err := db.DB.QueryRow("SELECT count(*) FROM lang").Scan(&nbLang)
	if err != nil {
		log.Fatal("Unable to fetch langs in database")
	}
	// If true, show a warning message and exit
	if nbLang > 1 {
		return errors.New("/!\\ Database already contains langs. As everything is linked to langs in arkeogis, feel free to destroys the db yourself.")
	} else {
		if err := insertDefaultValues(); err != nil {
			log.Fatal(err)
		}
	}
	// Langs map stores the langs and associates them to their id for further use.
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	// Store langs in database
	stmt, err := tx.Prepare("INSERT INTO lang (iso_code, active) VALUES ($1, false) RETURNING id")
	lineNum := 0
	if err != nil {
		return err
	}
	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// Skip first line
		if lineNum == 0 {
			lineNum += 1
			continue
		}
		// Skip blank lang iso code
		if strings.TrimSpace(s[2]) == "" {
			continue
		}
		var id int
		if _, ok := Langs[s[2]]; !ok {
			if err = stmt.QueryRow(s[2]).Scan(&id); err != nil {
				tx.Rollback()
				return err
			}
			Langs[s[2]] = id
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	err = importContinents()
	if err != nil {
		return err
	}
	return nil
}

// func cacheExistingCountries makes an index of existing counries in database
func cacheExistingCountries() {
	rows, err := db.DB.Query("SELECT geonameid, iso_code FROM country")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var geonameid int
	var iso_code string
	for rows.Next() {
		if err := rows.Scan(&geonameid, &iso_code); err != nil {
			log.Fatal(err)
		}
		CachedCountries[iso_code] = geonameid
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

// func importCountries parses the geoname file "countryInfo.txt" and populates it in the database
// If country already set in database it updates it
// Else it adds it
// Warning: No country is removed from database by this function !

func importCountries(rc io.Reader) error {
	scan := bufio.NewScanner(rc)
	// Make an index from existing countries in database
	cacheExistingCountries()
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	var stmtInsert1, stmtInsert2, stmtUpdate1, stmtUpdate2 *sql.Stmt
	var errStmt1, errStmt2 error
	if len(CachedCountries) > 1 {
		fmt.Println("- Updating existing countries")
		stmtUpdate1, errStmt1 = tx.Prepare("UPDATE country SET iso_code = $2 WHERE geonameid = $1")
		stmtUpdate2, errStmt2 = tx.Prepare("UPDATE country_tr SET lang_id = $2, name = $3, name_ascii = $4 WHERE country_geonameid = $1")
	} else {
		fmt.Println("- Inserting countries")
		stmtInsert1, errStmt1 = tx.Prepare("INSERT INTO country (geonameid, iso_code, created_at, updated_at) VALUES ($1, $2, $3, $4)")
		stmtInsert2, errStmt2 = tx.Prepare("INSERT INTO country_tr (country_geonameid, lang_id, name, name_ascii) VALUES ($1, $2, $3, $4)")
	}
	if errStmt1 != nil {
		return errStmt1
	}
	if errStmt2 != nil {
		return errStmt2
	}

	rgxp_onlycities, _ := regexp.Compile("^PCLI")

	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// ignore line with comment
		if strings.Index(line, "#") == 0 {
			continue
		}
		// get only cities
		if !rgxp_onlycities.MatchString(strings.TrimSpace(s[7])) {
			continue
		}
		if strings.TrimSpace(s[0]) == "" {
			fmt.Println("GeonameID not found:", s[0])
		}
		GeonameID, _ := strconv.Atoi(s[0])
		if strings.TrimSpace(s[1]) == "" {
			fmt.Println("Country name not found:", s[1])
			continue
		}
		if strings.TrimSpace(s[2]) == "" {
			fmt.Println("ASCII name not found:", s[1])
			continue
		}
		name_ascii := strings.ToLower(s[2])
		if _, ok := CachedCountries[s[0]]; ok {
			if _, err := stmtUpdate1.Exec(GeonameID, s[8]); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtUpdate2.Exec(GeonameID, Langs["D"], s[1], name_ascii); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			fmt.Println("S8", s[8])
			if _, err := stmtInsert1.Exec(GeonameID, s[8], time.Now(), time.Now()); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtInsert2.Exec(GeonameID, Langs["D"], s[1], name_ascii); err != nil {
				tx.Rollback()
				return err
			}
			CountriesById[GeonameID] = s[8]
			Countries[s[8]] = GeonameID
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
	var geonameid int
	for rows.Next() {
		if err := rows.Scan(&geonameid); err != nil {
			log.Fatal(err)
		}
		CachedCitiesById[geonameid] = true
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
	if len(CachedCitiesById) > 1 {
		fmt.Println("- Updating existing cities")
		stmtUpdate1, errStmt1 = tx.Prepare("UPDATE city SET country_geonameid = $2, geom_centroid = ST_GeomFromText($3, 4326), updated_at=now() WHERE geonameid = $1")
		stmtUpdate2, errStmt2 = tx.Prepare("UPDATE city_tr SET lang_id = $2, name = $3, name_ascii = $4 WHERE city_geonameid = $1")
	} else {
		fmt.Println("- Inserting cities")
		stmtInsert1, errStmt1 = tx.Prepare("INSERT INTO city (geonameid, country_geonameid, geom_centroid, created_at, updated_at) VALUES ($1, $2, ST_GeomFromText($3, 4326), now(), now())")
		stmtInsert2, errStmt2 = tx.Prepare("INSERT INTO city_tr (city_geonameid, lang_id, name, name_ascii) VALUES ($1, $2, $3, $4)")
	}
	if errStmt1 != nil {
		return errStmt1
	}
	if errStmt2 != nil {
		return errStmt2
	}

	rgxp_onlycities, _ := regexp.Compile("^PPL.*")

	for scan.Scan() {
		line := scan.Text()
		s := strings.Split(line, "\t")
		// ignore line with comment
		if strings.Index(line, "#") == 0 {
			continue
		}
		// get only cities
		if !rgxp_onlycities.MatchString(strings.TrimSpace(s[7])) {
			continue
		}
		if strings.TrimSpace(s[0]) == "" {
			fmt.Println("GeonameID not found")
		}
		GeonameID, _ := strconv.Atoi(s[0])
		if strings.TrimSpace(s[1]) == "" {
			fmt.Println("City name not found")
			continue
		}
		if strings.TrimSpace(s[2]) == "" {
			fmt.Println("ASCII name not found")
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
		lat, err := strconv.ParseFloat(s[4], 64)
		if err != nil {
			fmt.Println("Unable to convert latitude to float")
		}
		lon, err := strconv.ParseFloat(s[5], 64)
		if err != nil {
			fmt.Println("Unable to convert longitude to float")
		}
		geom := fmt.Sprintf("POINT(%f %f)", lon, lat)

		name_ascii := strings.ToLower(s[2])
		if _, ok := Countries[s[8]]; !ok {
			fmt.Println("Country not found:", s[8])
			continue
		}
		if _, ok := CachedCitiesById[GeonameID]; ok {
			if _, err := stmtUpdate1.Exec(GeonameID, Countries[s[8]], geom); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtUpdate2.Exec(GeonameID, Langs["D"], s[1], name_ascii); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if _, err := stmtInsert1.Exec(GeonameID, Countries[s[8]], geom); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtInsert2.Exec(GeonameID, Langs["D"], s[1], name_ascii); err != nil {
				tx.Rollback()
				return err
			}
			CitiesById[GeonameID] = true
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
	var geonameid int
	for rows.Next() {
		if err := rows.Scan(&geonameid); err != nil {
			log.Fatal(err)
		}
		CachedContinentsById[geonameid] = true
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
	if len(CachedContinentsById) > 1 {
		fmt.Println("- Updating existing continents")
		stmtUpdate1, errStmt1 = tx.Prepare("UPDATE continent SET iso_code = $2, updated_at = $3 WHERE geonameid = $1")
		stmtUpdate2, errStmt2 = tx.Prepare("UPDATE continent_tr SET lang_id = $2, name = $3, name_ascii = $4 WHERE continent_geonameid = $1")
	} else {
		fmt.Println("- Inserting continents")
		stmtInsert1, errStmt1 = tx.Prepare("INSERT INTO continent (geonameid, iso_code, created_at, updated_at) VALUES ($1, $2, $3, $4)")
		stmtInsert2, errStmt2 = tx.Prepare("INSERT INTO continent_tr (continent_geonameid, lang_id, name, name_ascii) VALUES ($1, $2, $3, $4)")
	}
	if errStmt1 != nil {
		return errStmt1
	}
	if errStmt2 != nil {
		return errStmt2
	}
	for GeonameID, infos := range geonamesContinents {
		if _, ok := CachedContinentsById[GeonameID]; ok {
			if _, err := stmtUpdate1.Exec(GeonameID, infos["IsoCode"], time.Now()); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtUpdate2.Exec(GeonameID, Langs["D"], infos["Name"], strings.ToLower(infos["Name"])); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if _, err := stmtInsert1.Exec(GeonameID, infos["IsoCode"], time.Now(), time.Now()); err != nil {
				tx.Rollback()
				return err
			}
			if _, err := stmtInsert2.Exec(GeonameID, Langs["D"], infos["Name"], strings.ToLower(infos["Name"])); err != nil {
				tx.Rollback()
				return err
			}
		}
		ContinentsById[GeonameID] = true
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
			if _, ok := ContinentsById[GeonameID]; ok {
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
			ContinentsById[GeonameID] = s[8]
			Continents[s[8]] = GeonameID
		}
	*/
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// func importAlternateNames parses the geoname file "alternateNames.zip" and populates in the database
// As nothing rely directly to alternates names, we first try do destroy all datas with lang = D (our lang_id for "D" language) to trash older datas

func importAlternateNames(rc io.Reader) error {
	scan := bufio.NewScanner(rc)
	alreadyProcessed := map[string]bool{}
	// Delete old entries wich are not tagged with "Undefined lang"
	for _, tablename := range []string{"continent_tr", "country_tr", "city_tr"} {
		if _, err := db.DB.Exec("DELETE FROM "+tablename+" WHERE lang_id != (SELECT id FROM lang WHERE iso_code = $1)", "D"); err != nil {
			return err
		}
	}
	tx := db.DB.MustBegin()
	tx.MustExec("SET CONSTRAINTS ALL DEFERRED")
	stmtContinent, err := tx.Prepare("INSERT INTO continent_tr (continent_geonameid, lang_id, name, name_ascii) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	stmtCountry, err := tx.Prepare("INSERT INTO country_tr (country_geonameid, lang_id, name, name_ascii) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	stmtCity, err := tx.Prepare("INSERT INTO city_tr (city_geonameid, lang_id, name, name_ascii) VALUES ($1, $2, $3, $4)")
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
		iso_code := strings.TrimSpace(s[2])
		if iso_code == "" || len(iso_code) != 2 {
			continue
		}
		// Skip line with no name
		if strings.TrimSpace(s[3]) == "" {
			continue
		}
		preferred, _ := strconv.Atoi(s[4])
		// Uniq code
		uniqCode := s[1] + "_" + iso_code

		if _, ok := Langs[iso_code]; ok {
			if _, ok := CitiesById[GeonameID]; ok {
				if _, ok := alreadyProcessed[uniqCode]; ok {
					if preferred == 1 {
						_, err = tx.Exec("DELETE FROM city_tr WHERE city_geonameid = $1 AND lang_id = $2", GeonameID, Langs[iso_code])
						if err != nil {
							return err
						}
					} else {
						continue
					}
				}
				if _, err = stmtCity.Exec(GeonameID, Langs[iso_code], s[3], ""); err != nil {
					tx.Rollback()
					return err
				}
			} else if _, ok := CountriesById[GeonameID]; ok {
				if _, ok := alreadyProcessed[uniqCode]; ok {
					if preferred == 1 {
						_, err = tx.Exec("DELETE FROM country_tr WHERE country_geonameid = $1 AND lang_id = $2", GeonameID, Langs[iso_code])
						if err != nil {
							return err
						}
					} else {
						continue
					}
				}
				if _, err = stmtCountry.Exec(GeonameID, Langs[iso_code], s[3], ""); err != nil {
					tx.Rollback()
					return err
				}
			} else if _, ok := ContinentsById[GeonameID]; ok {
				if _, ok := alreadyProcessed[uniqCode]; ok {
					if preferred == 1 {
						_, err = tx.Exec("DELETE FROM continent_tr WHERE continent_geonameid = $1 AND lang_id = $2", GeonameID, Langs[iso_code])
						if err != nil {
							return err
						}
					} else {
						continue
					}
				}
				if _, err = stmtContinent.Exec(GeonameID, Langs[iso_code], s[3], ""); err != nil {
					tx.Rollback()
					return err
				}
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
	var undefinedLangId int
	err := db.DB.QueryRow("INSERT INTO lang (id, iso_code, active) VALUES (0, 'D', true) RETURNING id").Scan(&undefinedLangId)
	if err != nil {
		fmt.Println("can't insert default lang D : ", err)
		return err
	}
	Langs["D"] = undefinedLangId
	// Insert "undefined" continent and populate cache index
	var undefinedContinentId int
	err = db.DB.QueryRow("INSERT INTO continent (geonameid, iso_code, created_at, updated_at) VALUES (0, 'U', $1, $2) RETURNING geonameid", time.Now(), time.Now()).Scan(&undefinedContinentId)
	if err != nil {
		fmt.Println("can't insert default continent D : ", err)
		return err
	}
	// Insert "undefined" country and populate cache index
	var undefinedCountryId int
	err = db.DB.QueryRow("INSERT INTO country (geonameid, iso_code, created_at, updated_at) VALUES (0, 'U', $1, $2) RETURNING geonameid", time.Now(), time.Now()).Scan(&undefinedCountryId)
	if err != nil {
		fmt.Println("can't insert default country D : ", err)
		return err
	}
	Countries["U"] = undefinedCountryId
	// Insert "undefined" city
	var undefinedCityId int
	err = db.DB.QueryRow("INSERT INTO city (geonameid, country_geonameid, geom_centroid, created_at, updated_at) VALUES (0, 0, " + geom_nowhere + ", now(), now()) RETURNING geonameid").Scan(&undefinedCityId)
	if err != nil {
		fmt.Println("can't insert city : ", err)
		return err
	}
	CitiesById[undefinedCityId] = true
	return nil
}

// Function main() is responsible to list the files we want to process then call download and import functions
func main() {
	files := []string{"iso-languagecodes.txt", "allCountries.zip", "allCountries.zip", "alternateNames.zip"}
	//	files := []string{"alternateNames.zip"}
	Langs = map[string]int{}
	ContinentsById = map[int]bool{}
	CachedContinentsById = map[int]bool{}
	Countries = map[string]int{}
	CachedCountries = map[string]int{}
	CountriesById = map[int]string{}
	CachedCitiesById = map[int]bool{}
	CitiesById = map[int]bool{}
	l := len(files)
	for i := 0; i < l; i++ {
		url := "http://download.geonames.org/export/dump/" + files[i]
		if err := downloadFromUrl(url); err != nil {
			fmt.Println(err)
		}
		if err := importFile(files[i]); err != nil {
			fmt.Println(err)
			return
		}
	}
}
