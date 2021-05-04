/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2019 CROLL SAS
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

package export

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"log"
	"math"
	"strconv"
	"strings"

	model "github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
)

// SitesAsCSV exports database and sites as as csv file
func SitesAsCSV(siteIDs []int, isoCode string, includeDbName bool, includeSiteId bool, includeInterop bool, tx *sqlx.Tx) (outp string, err error) {

	var buff bytes.Buffer

	w := csv.NewWriter(&buff)
	w.Comma = ';'
	w.UseCRLF = true

	columns := []string{}

	if includeDbName {
		columns = append(columns, "DATABASE_NAME")
	}
	if includeSiteId {
		columns = append(columns, "SITE_AKG_ID")
	}
	columns = append(columns, "SITE_SOURCE_ID")
	columns = append(columns, "SITE_NAME")
	columns = append(columns, "MAIN_CITY_NAME")
	columns = append(columns, "GEONAME_ID")
	columns = append(columns, "PROJECTION_SYSTEM")
	columns = append(columns, "LONGITUDE")
	columns = append(columns, "LATITUDE")
	columns = append(columns, "ALTITUDE")
	columns = append(columns, "CITY_CENTROID")
	columns = append(columns, "STATE_OF_KNOWLEDGE")
	columns = append(columns, "OCCUPATION")
	columns = append(columns, "STARTING_PERIOD")
	columns = append(columns, "ENDING_PERIOD")
	columns = append(columns, "CARAC_NAME")
	columns = append(columns, "CARAC_LVL1")
	columns = append(columns, "CARAC_LVL2")
	columns = append(columns, "CARAC_LVL3")
	columns = append(columns, "CARAC_LVL4")
	columns = append(columns, "CARAC_EXP")
	if includeInterop {
		columns = append(columns, "ARK_CARAC_ID")
		columns = append(columns, "Ark PACTOLS")
		columns = append(columns, "URI_SITE")
	}
	columns = append(columns, "BIBLIOGRAPHY")
	columns = append(columns, "COMMENTS")

	err = w.Write(columns)
	if err != nil {
		log.Println("database::ExportCSV : ", err.Error())
	}
	w.Flush()

	// Cache characs
	characs := make(map[int]string)

	q := "WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || ';' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC"

	rows, err := tx.Query(q, isoCode)
	switch {
	case err == sql.ErrNoRows:
		rows.Close()
		return outp, nil
	case err != nil:
		rows.Close()
		return
	}
	for rows.Next() {
		var id int
		var path string
		if err = rows.Scan(&id, &path); err != nil {
			return
		}
		characs[id] = path
	}

	q = "SELECT s.id as site_id, db.name as dbname, s.code, s.name, s.city_name, s.city_geonameid, ST_X(s.geom::geometry) as longitude, ST_Y(s.geom::geometry) as latitude, ST_X(s.geom_3d::geometry) as longitude_3d, ST_Y(s.geom_3d::geometry) as latitude3d, ST_Z(s.geom_3d::geometry) as altitude, s.centroid, s.occupation, sr.start_date1, sr.start_date2, sr.end_date1, sr.end_date2, src.exceptional, src.knowledge_type, srctr.bibliography, srctr.comment, c.id as charac_id, c.ark_id, c.pactols_id FROM site s LEFT JOIN database db ON s.database_id = db.id LEFT JOIN site_range sr ON s.id = sr.site_id LEFT JOIN site_tr str ON s.id = str.site_id LEFT JOIN site_range__charac src ON sr.id = src.site_range_id LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id LEFT JOIN charac c ON src.charac_id = c.id WHERE s.id in (" + model.IntJoin(siteIDs, true) + ") AND str.lang_isocode IS NULL OR str.lang_isocode = db.default_language ORDER BY s.id, sr.id"

	rows2, err := tx.Query(q)
	if err != nil {
		rows2.Close()
		return
	}
	for rows2.Next() {
		var (
			site_id		   string
			dbname         string
			code           string
			name           string
			city_name      string
			city_geonameid int
			longitude      float64
			latitude       float64
			longitude3d    float64
			latitude3d     float64
			altitude3d     float64
			centroid       bool
			occupation     string
			start_date1    int
			start_date2    int
			end_date1      int
			end_date2      int
			knowledge_type string
			exceptional    bool
			bibliography   string
			comment        string
			charac_id      int
			slongitude     string
			slatitude      string
			saltitude      string
			scentroid      string
			soccupation    string
			scharacs       string
			scharac_name   string
			scharac_lvl1   string
			scharac_lvl2   string
			scharac_lvl3   string
			scharac_lvl4   string
			sexceptional   string
			// description    string
			arkid          string   // "ARK_CARAC_ID
			arkpactols     string   // "Ark PACTOLS"
			urisite		   string
		)
		if err = rows2.Scan(&site_id, &dbname, &code, &name, &city_name, &city_geonameid, &longitude, &latitude, &longitude3d, &latitude3d, &altitude3d, &centroid, &occupation, &start_date1, &start_date2, &end_date1, &end_date2, &exceptional, &knowledge_type, &bibliography, &comment, &charac_id, &arkid, &arkpactols); err != nil {
			log.Println(err)
			rows2.Close()
			return
		}
		// Geonameid
		var cgeonameid string
		if city_geonameid != 0 {
			cgeonameid = strconv.Itoa(city_geonameid)
		}
		// Longitude
		slongitude = strconv.FormatFloat(longitude, 'f', -1, 32)
		// Latitude
		slatitude = strconv.FormatFloat(latitude, 'f', -1, 32)
		// Altitude
		if longitude3d == 0 && latitude3d == 0 && altitude3d == 0 {
			saltitude = ""
		} else {
			saltitude = strconv.FormatFloat(altitude3d, 'f', -1, 32)
		}
		// Centroid
		if centroid {
			scentroid = translate.T(isoCode, "IMPORT.CSVFIELD_ALL.T_LABEL_YES")
		} else {
			scentroid = translate.T(isoCode, "IMPORT.CSVFIELD_ALL.T_LABEL_NO")
		}
		// Occupation
		switch occupation {
		case "not_documented":
			soccupation = translate.T(isoCode, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_NOT_DOCUMENTED")
		case "single":
			soccupation = translate.T(isoCode, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_SINGLE")
		case "continuous":
			soccupation = translate.T(isoCode, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_CONTINUOUS")
		case "multiple":
			soccupation = translate.T(isoCode, "IMPORT.CSVFIELD_OCCUPATION.T_LABEL_MULTIPLE")
		}
		// State of knowledge
		switch knowledge_type {
		case "not_documented":
			knowledge_type = translate.T(isoCode, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_NOT_DOCUMENTED")
		case "literature":
			knowledge_type = translate.T(isoCode, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_LITERATURE")
		case "prospected_aerial":
			knowledge_type = translate.T(isoCode, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_AERIAL")
		case "prospected_pedestrian":
			knowledge_type = translate.T(isoCode, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_PEDESTRIAN")
		case "surveyed":
			knowledge_type = translate.T(isoCode, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_SURVEYED")
		case "dig":
			knowledge_type = translate.T(isoCode, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_DIG")
		}
		// Revert hack on dates
		if start_date1 < 1 && start_date1 > math.MinInt32 {
			start_date1--
		}
		if start_date2 < 1 && start_date2 > math.MinInt32 {
			start_date2--
		}
		if end_date1 < 1 && end_date1 > math.MinInt32 {
			end_date1--
		}
		if end_date2 < 1 && end_date2 > math.MinInt32 {
			end_date2--
		}
		// Starting period
		startingPeriod := ""
		if start_date1 != math.MinInt32 {
			startingPeriod += strconv.Itoa(start_date1)
		}
		if start_date1 != math.MinInt32 && start_date2 != math.MaxInt32 && start_date1 != start_date2 {
			startingPeriod += ":"
		}
		if start_date2 != math.MaxInt32 && start_date1 != start_date2 {
			startingPeriod += strconv.Itoa(start_date2)
		}
		if startingPeriod == "" {
			startingPeriod = translate.T(isoCode, "IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED")
		}
		// Ending period
		endingPeriod := ""
		if end_date1 != math.MinInt32 {
			endingPeriod += strconv.Itoa(end_date1)
		}
		if end_date1 != math.MinInt32 && end_date2 != math.MaxInt32 && end_date1 != end_date2 {
			endingPeriod += ":"
		}
		if end_date2 != math.MaxInt32 && end_date1 != end_date2 {
			endingPeriod += strconv.Itoa(end_date2)
		}
		if endingPeriod == "" {
			endingPeriod = translate.T(isoCode, "IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED")
		}
		// Caracs
		var characPath = characs[charac_id]
		// fmt.Println(code, characPath)
		num := strings.Count(characPath, ";")
		if num < 4 {
			scharacs += characPath + strings.Repeat(";", 4-num)
		} else {
			scharacs = characPath
		}
		scharac_lvl2 = ""
		scharac_lvl3 = ""
		scharac_lvl4 = ""
		for i, c := range strings.Split(scharacs, ";") {
			// fmt.Println(i, c)
			switch i {
			case 0:
				scharac_name = c
			case 1:
				scharac_lvl1 = c
			case 2:
				scharac_lvl2 = c
			case 3:
				scharac_lvl3 = c
			case 4:
				scharac_lvl4 = c
			}

		}
		// fmt.Println(scharac_name, scharac_lvl1, scharac_lvl2, scharac_lvl3, scharac_lvl4)
		// fmt.Println(startingPeriod, endingPeriod)
		// Caracs exp
		if exceptional {
			sexceptional = translate.T(isoCode, "IMPORT.CSVFIELD_ALL.T_LABEL_YES")
		} else {
			sexceptional = translate.T(isoCode, "IMPORT.CSVFIELD_ALL.T_LABEL_NO")
		}

		var line []string

		if includeDbName {
			line = append(line, dbname)
		}
		if includeSiteId {
			line = append(line, site_id)
		}
		line = append(line, code)
		line = append(line, name)
		line = append(line, city_name)
		line = append(line, cgeonameid)
		line = append(line, "4326")
		line = append(line, slongitude)
		line = append(line, slatitude)
		line = append(line, saltitude)
		line = append(line, scentroid)
		line = append(line, knowledge_type)
		line = append(line, soccupation)
		line = append(line, startingPeriod)
		line = append(line, endingPeriod)
		line = append(line, scharac_name)
		line = append(line, scharac_lvl1)
		line = append(line, scharac_lvl2)
		line = append(line, scharac_lvl3)
		line = append(line, scharac_lvl4)
		line = append(line, sexceptional)
		if includeInterop {
			line = append(line, arkid)
			line = append(line, arkpactols)
			line = append(line, urisite)
		}
		line = append(line, bibliography)
		line = append(line, comment)

		err := w.Write(line)
		w.Flush()
		if err != nil {
			log.Println("database::ExportCSV : ", err.Error())
		}
	}

	return buff.String(), nil
}
