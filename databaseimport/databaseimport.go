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

package databaseimport

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/geo"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
)

// UserChoices stores user preferences for the parsing process
type UserChoices struct {
	UseGeonames bool
}

// DatabaseFullInfos is a meta struct which stores all the informations about
// a database
type DatabaseFullInfos struct {
	model.Database
	model.Database_tr
	Authors    []int
	Continents []int
	Countries  []int
	Exists     bool
	Init       bool
}

// SiteRangeFullInfos is a meta struct which stores all the informationss
// about a site range
type SiteRangeFullInfos struct {
	model.Site_range
	model.Site_range_tr
	Characs []int
}

// SiteFullInfos is a meta struct which stores all the informations about a site
type SiteFullInfos struct {
	model.Site
	CurrentSiteRange SiteRangeFullInfos
	NbSiteRanges     int
	Characs          []int
	HasError         bool
	Point            *geo.Point
	Latitude         string
	Longitude        string
}

// ImportError is the struct used to return errors, it enhances the errors struct to return more informations like line and field
type ImportError struct {
	Line     int      `json:"line"`
	SiteCode string   `json:"siteCode"`
	Value    string   `json:"value"`
	Columns  []string `json:"columns"`
	ErrMsg   string   `json:"errMsg"`
}

// The Error() func formats the error message
func (e *ImportError) Error() string {
	return fmt.Sprintf("line %d, column %s: %s", e.Line, strings.Join(e.Columns, ","), e.ErrMsg)
}

// AddError structures errors to be logged or returned to client
func (di *DatabaseImport) AddError(value string, errMsg string, columns ...string) {

	di.Errors = append(di.Errors, &ImportError{
		Line:     di.Parser.Line,
		SiteCode: di.CurrentSite.Code,
		Columns:  columns,
		Value:    value,
		ErrMsg:   translate.T(di.Parser.Lang, errMsg),
	})

	di.CurrentSite.HasError = true

	// Store site as containing error
	if di.CurrentSite.Code != "" {
		if _, ok := di.SitesWithError[di.CurrentSite.Code]; !ok {
			di.SitesWithError[di.CurrentSite.Code] = true
		}
	}
}

// DatabaseImport is a meta struct which stores all the informations about a site
type DatabaseImport struct {
	SitesProcessed map[string]int
	Database       *DatabaseFullInfos
	CurrentSite    *SiteFullInfos
	CurrentCharac  string
	Simulate       bool
	Tx             *sqlx.Tx
	Parser         *Parser
	ArkeoCharacs   map[string]map[string]int
	NumberOfSites  int
	SitesWithError map[string]bool
	Errors         []*ImportError
}

// New creates a new import process
func (di *DatabaseImport) New(parser *Parser, uid int, databaseName string, langID int, simu bool) error {
	var err error
	if uid <= 0 {
		return errors.New("Invalid user id")
	}
	di.Simulate = simu
	di.Database = &DatabaseFullInfos{}
	di.Database.Owner = uid
	di.Database.Default_language = langID
	di.CurrentSite = &SiteFullInfos{}
	di.Parser = parser
	di.NumberOfSites = 0
	di.SitesWithError = map[string]bool{}

	// Start database transaction
	di.Tx, err = db.DB.Beginx()
	if err != nil {
		return errors.New("Can't start transaction for database import")
	}

	// Cache characs defined in Arkeogis
	// TODO: Get only needed characs filtering by user id and project
	di.ArkeoCharacs = map[string]map[string]int{}
	di.ArkeoCharacs, err = di.cacheCharacs()
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Field DATABASE_SOURCE_NAME
	if di.Database.Name == "" {
		if databaseName != "" {
			if di.Database.Init == false {
				if err = di.processDatabaseName(databaseName); err != nil {
					di.AddError(databaseName, "IMPORT.CSVFIELD_DATABASE_SOURCE_NAME.T_CHECK_INVALID", "DATABASE_SOURCE_NAME")
				}
			}
		} else {
			di.AddError(databaseName, "IMPORT.CSVFIELD_DATABASE_SOURCE_NAME.T_CHECK_EMPTY", "DATABASE_SOURCE_NAME")
		}
	}

	return nil
}

// periodRegexp is used to match if starting and ending periods are valid
//var periodRegexp = regexp.MustCompile(`(-?\d{0,}):(-?\d{0,})`)
var uniqDateRexep = regexp.MustCompile(`^(-?\d+)$`)
var periodRegexpDate1 = regexp.MustCompile(`^(-?\d{0,}):?-?\d{0,}$`)
var periodRegexpDate2 = regexp.MustCompile(`^-?\d{0,}:?(-?\d{0,})$`)

// setDefaultValues init the di.Database object with default values
func (di *DatabaseImport) setDefaultValues() {
	di.Database.Scale_resolution = "site"
	di.Database.Geographical_extent = "world"
	di.Database.Type = "inventory"
	di.Database.Source_creation_date = time.Now()
	di.Database.Data_set = ""
	di.Database.Source = ""
	di.Database.Source_url = ""
	di.Database.Publisher = ""
	di.Database.Contributor = ""
	di.Database.Relation = ""
	di.Database.Coverage = ""
	di.Database.Copyright = ""
	di.Database.Context = "todo"
	di.Database.State = "in-progress"
	di.Database.Published = false
	di.Database.License_id = 1
	di.Database.Created_at = time.Now()
	di.Database.Updated_at = time.Now()
}

// ProcessRecord is triggered for each line of csv
func (di *DatabaseImport) ProcessRecord(f *Fields) {

	//fmt.Println(di.Parser.Line, " - ", f)

	// if site id not set and no previous SITE_SOURCE_ID is set, produce an error
	if di.CurrentSite.Code == "" && f.SITE_SOURCE_ID == "" {
		di.AddError("", "IMPORT.CSVFIELD_SITE_SOURCE_ID.T_CHECK_EMPTY", "SITE_SOURCE_ID")
		return
	}

	// If site code is not empty and differs, create a new instance of SiteFullInfos to store datas
	if f.SITE_SOURCE_ID != "" && f.SITE_SOURCE_ID != di.CurrentSite.Code {
		di.CurrentSite = &SiteFullInfos{}
		di.CurrentSite.Code = f.SITE_SOURCE_ID
		di.CurrentSite.Name = f.SITE_NAME
		di.NumberOfSites++
		// Process site info
		di.processSiteInfos(f)
	} else {
		di.CurrentSite.NbSiteRanges++
		di.checkDifferences(f)
	}

	// Init the site range if necessary
	if di.CurrentSite.NbSiteRanges == 0 {
		di.CurrentSite.CurrentSiteRange = SiteRangeFullInfos{}
	}

	// Process site range infos
	di.processSiteRangeInfos(f)

}

// processDatabaseName verifies if a database name already exist for a user and
// create or update the sql entry
func (di *DatabaseImport) processDatabaseName(name string) error {
	var err error
	// Store database name
	di.Database.Name = name

	// Check database name length
	if len(name) > 50 {
		di.AddError("", "IMPORT.FORM_DATABASE_NAME.T_CHECK_TOO_LONG", "DATABASE_NAME")
		return errors.New("Database name too long")
	}

	// Check if another user as a database with same name
	alreadyExists, err := di.Database.AnotherExistsWithSameName(di.Tx)
	if err != nil {
		return err
	}

	if alreadyExists {
		di.AddError("", "IMPORT.FORM_DATABASE_NAME.T_CHECK_OTHER_USER_HAS_DB_WITH_SAME_NAME", "DATABASE_NAME")
		return err
	}

	di.Database.Exists, err = di.Database.DoesExist(di.Tx)
	if err != nil {
		return err
	}

	if di.Database.Exists {
		// Get database infos
		di.Database.GetInfos(di.Tx)
		// Update record
		err = di.Database.Update(di.Tx)
	} else {
		if di.Simulate {
			fmt.Println("Simulation")
			di.setDefaultValues()
		}
		// Create record in database
		err = di.Database.Create(di.Tx)
		// Set again values if set in Simulation mode
		di.Database.Name = name
		di.Database.Init = true
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// processSiteInfos deals informations about site (not site range)
func (di *DatabaseImport) processSiteInfos(f *Fields) {

	// MAIN_CITY_NAME
	if f.MAIN_CITY_NAME != "" {
		di.CurrentSite.City_name = f.MAIN_CITY_NAME
	}

	// CITY_CENTROID
	if f.CITY_CENTROID != "" {
		val, err := di.valueAsBool("CITY_CENTROID", f.CITY_CENTROID)
		if err == nil {
			di.CurrentSite.Centroid = val
		}
	}

	// If only one of lat or lon empty
	if (f.LATITUDE != "" && f.LONGITUDE == "") || (f.LONGITUDE != "" && f.LATITUDE == "") {
		di.AddError(f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_LATITUDE_OR_LONGITUDE.T_CHECK_ONE_IS_EMPTY_OTHER_NOT", "LATITUDE", "LONGITUDE")
	} else {
		// If lat and lon not empty, process geo datas
		if f.LATITUDE != "" && f.LONGITUDE != "" {
			point, err := di.processGeoDatas(f)
			if err != nil {
			} else {
				di.CurrentSite.Point = point
				// Store lat and lon to check differences if site has multiple site ranges
				di.CurrentSite.Latitude = f.LATITUDE
				di.CurrentSite.Longitude = f.LONGITUDE
			}
		} else {
			// User don't want to use Geonames, we are stuck
			if !di.Parser.UserChoices.UseGeonames {
				di.AddError(f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_GEO.T_CHECK_LAT_OR_LON_NOT_SET_AND_NO_GEONAMES", "LATITUDE", "LONGITUDE", "GEONAME_ID")
			} else {
				// If user chose to use Geonames, and we don't have valid coordinates at this point, use geonames functionality
				point, err := di.processGeonames(f)
				if err != nil {
					di.AddError("GEONAME_ID", "IMPORT.CSVFIELD_GEONAME_ID.T_PROCESS_INVALID")
				} else {
					di.CurrentSite.Point = point
					// Has we used Geonames, site location type is "centroid"
					di.CurrentSite.Centroid = true
				}
			}
		}
	}

	// OCCUPATION
	val, err := di.getOccupation(f.OCCUPATION)
	if err == nil {
		di.CurrentSite.Occupation = val
	}
}

// checkDifferences verifies if values entered for the site are identical
func (di *DatabaseImport) checkDifferences(f *Fields) {

	// MAIN_CITY_NAME
	if f.MAIN_CITY_NAME != "" && di.CurrentSite.City_name != f.MAIN_CITY_NAME {
		di.AddError(f.MAIN_CITY_NAME, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "MAIN_CITY_NAME")
	}

	// CITY_CENTROID
	if f.CITY_CENTROID != "" {
		val, err := di.valueAsBool("CITY_CENTROID", f.CITY_CENTROID)
		if err == nil && val != di.CurrentSite.Centroid {
			di.AddError(f.CITY_CENTROID, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "CITY_CENTROID")
		}
	}

	// LONGITUDE
	if f.LONGITUDE != "" && f.LONGITUDE != di.CurrentSite.Longitude {
		di.AddError(f.LONGITUDE, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "LONGITUDE")
	}

	// LATITUDE
	if f.LATITUDE != "" && f.LATITUDE != di.CurrentSite.Latitude {
		di.AddError(f.LATITUDE, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "LATITUDE")

	}

	// OCCUPATION
	if f.OCCUPATION != "" {
		val, err := di.getOccupation(f.OCCUPATION)
		if err == nil && val != di.CurrentSite.Occupation {
			di.AddError(val, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "OCCUPATION")
		}
	}
}

// getOccupation get occupation string from field translatable in the csv file
func (di *DatabaseImport) getOccupation(occupation string) (val string, err error) {
	err = nil
	switch strings.ToLower(occupation) {
	case di.lowerTranslation("IMPORT.CSVFIELD_OCCUPATION.T_LABEL_NOT_DOCUMENTED"):
		val = "not_documented"
	case di.lowerTranslation("IMPORT.CSVFIELD_OCCUPATION.T_LABEL_SINGLE"):
		val = "single"
	case di.lowerTranslation("IMPORT.CSVFIELD_OCCUPATION.T_LABEL_CONTINUOUS"):
		val = "continuous"
	case di.lowerTranslation("IMPORT.CSVFIELD_OCCUPATION.T_LABEL_MULTIPLE"):
		val = "multiple"
	default:
		if occupation == "" {
			di.AddError(occupation, "IMPORT.CSVFIELD_OCCUPATION.T_CHECK_EMPTY", "OCCUPATION")
		} else {
			di.AddError(occupation, "IMPORT.CSVFIELD_OCCUPATION.T_CHECK_INVALID", "OCCUPATION")
		}
	}
	return
}

func (di *DatabaseImport) processSiteRangeInfos(f *Fields) {

	// CARACTERISATIONS
	characs, _ := di.processCharacs(f)
	di.CurrentSite.CurrentSiteRange.Characs = characs

	// STARTING_PERIOD
	fmt.Println("Starting period", f.STARTING_PERIOD)
	dates := di.parseDates(f.STARTING_PERIOD)
	fmt.Println("Parsed dates", dates)
	di.CurrentSite.CurrentSiteRange.Start_date1 = dates[0]
	di.CurrentSite.CurrentSiteRange.Start_date2 = dates[1]
	fmt.Println("Start_date1", di.CurrentSite.CurrentSiteRange.Start_date1)
	fmt.Println("Start_date2", di.CurrentSite.CurrentSiteRange.Start_date2)

	// ENDING_PERIOD

	// STATE_OF_KNOWLEDGE
	switch strings.ToLower(f.STATE_OF_KNOWLEDGE) {
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_NOT_DOCUMENTED"):
		di.CurrentSite.CurrentSiteRange.Knowledge_type = "not_documented"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_LITERATURE"):
		di.CurrentSite.CurrentSiteRange.Knowledge_type = "literature"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_AERIAL"):
		di.CurrentSite.CurrentSiteRange.Knowledge_type = "prospected_aerial"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_PEDESTRIAN"):
		di.CurrentSite.CurrentSiteRange.Knowledge_type = "prospected_pedestrian"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_SURVEYED"):
		di.CurrentSite.CurrentSiteRange.Knowledge_type = "surveyed"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_DIG"):
		di.CurrentSite.CurrentSiteRange.Knowledge_type = "dig"
	default:
		if f.STATE_OF_KNOWLEDGE == "" {
			di.AddError(f.STATE_OF_KNOWLEDGE, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_CHECK_EMPTY", "STATE_OF_KNOWLEDGE")
		} else {
			di.AddError(f.STATE_OF_KNOWLEDGE, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_CHECK_INVALID", "STATE_OF_KNOWLEDGE")
		}
	}

	// BIBLIOGRAPHY
	di.CurrentSite.CurrentSiteRange.Bibliography = f.BIBLIOGRAPHY

	// COMMENTS
	di.CurrentSite.CurrentSiteRange.Comment = f.COMMENTS

}

// processGeoDatas analyzes and process csv fields related to geo informations
func (di *DatabaseImport) processGeoDatas(f *Fields) (*geo.Point, error) {
	var point *geo.Point
	var epsg int
	var err error
	var pointToBeReturned *geo.Point
	hasError := false
	// If projection system is not set, we assume it's 4326 (WGS84)
	if f.PROJECTION_SYSTEM == "" {
		epsg = 4326
	} else {
		epsg, err = strconv.Atoi(f.PROJECTION_SYSTEM)
		if err != nil {
			di.AddError(f.PROJECTION_SYSTEM, "IMPORT.CSVFIELD_PROJECTION_SYSTEM.T_CHECK_INCORRECT_VALUE", "PROJECTION_SYSTEM")
			hasError = true
		}
	}
	// Parse LONGITUDE
	lon, err := strconv.ParseFloat(strings.Replace(f.LONGITUDE, ",", ".", 1), 64)
	if err != nil || lon == 0 {
		di.AddError(f.LONGITUDE, "IMPORT.CSVFIELD_LONGITUDE.T_CHECK_INCORRECT_VALUE", "LONGITUDE")
		hasError = true
	}
	// Parse LATITUDE
	lat, err := strconv.ParseFloat(strings.Replace(f.LATITUDE, ",", ".", 1), 64)
	if err != nil || lat == 0 {
		di.AddError(f.LATITUDE, "IMPORT.CSVFIELD_LATITUDE.T_CHECK_INCORRECT_VALUE", "LATITUDE")
		hasError = true
	}
	// Parse ALTITUDE
	// If no altitude set, we have a 2D Point
	if f.ALTITUDE != "" {
		alt, err := strconv.ParseFloat(f.ALTITUDE, 64)
		if err != nil {
			di.AddError(f.ALTITUDE, "IMPORT.CSVFIELD_ALTITUDE.T_CHECK_INCORRECT_VALUE", "ALTITUDE")
			hasError = true
		}
		point, err = geo.NewPoint(epsg, lon, lat, alt)
	} else {
		point, err = geo.NewPoint(epsg, lon, lat)
	}
	if err != nil {
		di.AddError(f.PROJECTION_SYSTEM, "IMPORT.CSVFIELD_GEO.T_ERROR_UNABLE_TO_GET_WKT", "EPSG", "LATITUDE", "LONGITUDE")
	}
	// Datas are already in WGS84, leave it untouched
	if epsg == 4326 {
		pointToBeReturned = point
	} else {
		// Convert datas to WGS84
		point2, err := point.ToWGS84()
		if err != nil {
			di.AddError(f.PROJECTION_SYSTEM+" "+f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_GEO.T_ERROR_UNABLE_TO_CONVERT_TO_WGS84", "EPSG", "LATITUDE", "LONGITUDE")
			pointToBeReturned = nil
		}
		/*
			if err != nil {
				di.AddError(f.PROJECTION_SYSTEM+" "+f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_GEO.T_ERROR_UNABLE_TO_GET_WKT", "EPSG", "LATITUDE", "LONGITUDE")
				pointToBeReturned = nil
			}
		*/
		pointToBeReturned = point2
	}
	if hasError {
		di.AddError(f.PROJECTION_SYSTEM+" "+f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_GEO.T_ERROR_UNABLE_TO_CREATE_GEOMETRY", "EPSG", "LATITUDE", "LONGITUDE")
		return nil, err
	}
	return pointToBeReturned, nil
}

// processGeonames get the city name/lat/lon from the database and assign it TODO
func (di *DatabaseImport) processGeonames(f *Fields) (*geo.Point, error) {
	if f.GEONAME_ID == "" {
		di.AddError("", "IMPORT.CSVFIELD_GEONAME_ID.T_CHECK_EMPTY", "GEONAME_ID")
		return nil, errors.New("Empty geonameid")
	}
	id, err := strconv.Atoi(f.GEONAME_ID)
	if err != nil {
		di.AddError(f.GEONAME_ID, "IMPORT.CSVFIELD_GEONAME_ID.T_CHECK_INVALID", "GEONAME_ID")
		return nil, err
	}
	point, err := geo.NewPointByGeonameID(id)
	if err != nil {
		di.AddError(f.GEONAME_ID, "IMPORT.CSVFIELD_GEONAME_ID.T_ERROR_NO_MATCH", "GEONAME_ID")
		return nil, err
	}
	if err != nil {
		di.AddError(f.GEONAME_ID, "IMPORT.CSVFIELD_GEONAME_ID.T_ERROR_NO_MATCH", "GEONAME_ID")
		return nil, err
	}
	return point, nil
}

// processCharacs analyses the fields of each charac for each level
// It verify if charac of any level exists and if true, assign it to the site range
func (di *DatabaseImport) processCharacs(f *Fields) ([]int, error) {
	var characs []int
	path := ""
	lvl := 1
	if f.CARAC_NAME == "" {
		di.AddError(f.CARAC_NAME, "IMPORT.CSVFIELD_CARAC_NAME.T_CHECK_EMPTY", "CARAC_NAME")
		return characs, errors.New("invalid carac name")
	}
	di.CurrentCharac = f.CARAC_NAME
	if f.CARAC_LVL1 != "" {
		path += "->" + f.CARAC_LVL1
	} else {
		di.AddError(f.CARAC_NAME, "IMPORT.CSVFIELD_CARAC_LVL1.T_CHECK_EMPTY")
		return characs, errors.New("no lvl1 carac")
	}
	if f.CARAC_LVL2 != "" {
		path += "->" + f.CARAC_LVL2
		lvl++
	}
	if f.CARAC_LVL3 != "" {
		path += "->" + f.CARAC_LVL3
		lvl++
	}
	if f.CARAC_LVL4 != "" {
		path += "->" + f.CARAC_LVL4
		lvl++
	}
	//path = strings.TrimSuffix(path, "->")
	// Check if charac exists and retrieve id
	caracID := di.ArkeoCharacs[f.CARAC_NAME][f.CARAC_NAME+path]
	if caracID == 0 {
		di.AddError(f.CARAC_NAME+path, "IMPORT.CSVFIELD_CARACTERISATION.T_CHECK_INVALID", "CARAC_LVL"+strconv.Itoa(lvl))
		return characs, errors.New("invalid charac")
	}
	characs = append(characs, caracID)
	return characs, nil
}

// cacheCharacs get all Characs from database and cache them
func (di *DatabaseImport) cacheCharacs() (map[string]map[string]int, error) {
	characs := map[string]map[string]int{}
	characsRoot, err := model.GetAllCharacsRootFromLangId(di.Database.Default_language)
	if err != nil {
		return characs, err
	}
	for name := range characsRoot {
		characs[name], err = model.GetCharacPathsFromLangID(name, di.Database.Default_language)
		if err != nil {
			return characs, err
		}
	}
	return characs, nil
}

// valueAsBool analyses YES/NO translatable values to bool
func (di *DatabaseImport) valueAsBool(fieldName, val string) (choosenValue bool, err error) {
	if strings.ToLower(val) == di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_LABEL_YES") {
		choosenValue = true
	} else if strings.ToLower(val) == di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_LABEL_NO") {
		choosenValue = false
	} else {
		di.AddError(val, "IMPORT.CSVFIELD_ALL.T_CHECK_WRONG_VALUE", fieldName)
		return
	}
	return
}

// lowerTranslation return translation in lower case
func (di *DatabaseImport) lowerTranslation(s string) string {
	return strings.ToLower(translate.T(di.Parser.Lang, s))
}

// parseDates analyzes declared period and returns starting and ending dates
func (di *DatabaseImport) parseDates(period string) [2]int {
	// If empty period, set "min and max" dates
	if (period == "" || strings.ToLower(period) == di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED")) || strings.ToLower(period) == "null" {
		return [2]int{math.MinInt32, math.MaxInt32}
	}

	period = strings.Replace(period, "+", "", -1)
	var dates [2]int
	var date1 string
	var date2 string

	// If we have only a date, start date and end date are the same
	uniqDate := uniqDateRexep.FindString(period)
	if uniqDate != "" {
		fmt.Println("UNIQ DATE")
		date1 = uniqDate
		date2 = uniqDate
	} else {
		date1 = periodRegexpDate1.FindString(period)
		date2 = periodRegexpDate2.FindString(period)
	}

	fmt.Println("Date1 and Date2", date1, date2)

	// First date
	if date1 != "" {
		sd, err := strconv.Atoi(date1)
		if err != nil {
			di.AddError("IMPORT.CSVFIELD_STARTING_PERIOD_DATE1.T_CHECK_WRONG_VALUE", "STARTING_PERIOD")
		} else {
			dates[0] = sd
		}
	} else {
		dates[0] = math.MinInt32
	}

	// Second date
	if date2 != "" {
		ed, err := strconv.Atoi(date2)
		if err != nil {
			di.AddError("IMPORT.CSVFIELD_STARTING_PERIOD_DATE1.T_CHECK_WRONG_VALUE", "STARTING_PERIOD")
		} else {
			dates[1] = ed
		}
	} else {
		dates[1] = math.MaxInt32
	}
	return dates
}
