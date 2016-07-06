/* ArkeoGIS - The Geographic Information System for Archaeologists
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
	"log"
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

// DatabaseInfos is a meta struct which stores all the informations about
// a database
type DatabaseInfos struct {
	model.Database
	model.Database_tr
	Authors    []int
	Continents []int
	Countries  []int
	Exists     bool
	Init       bool
}

// CharacsInfos holds information about characs linked to site range
type SiteRangeCharacInfos struct {
	model.Site_range__charac
	model.Site_range__charac_tr
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
		ErrMsg:   translate.T(di.Parser.UserLang, errMsg),
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
	SitesProcessed         map[string]int
	Database               *DatabaseInfos
	CurrentSite            *model.SiteInfos
	CurrentSiteRange       *model.Site_range
	CurrentSiteRangeCharac *SiteRangeCharacInfos
	Tx                     *sqlx.Tx
	Parser                 *Parser
	Uid                    int
	ArkeoCharacs           map[string]map[string]int
	//ArkeoCharacsIDs  map[int][]int
	NumberOfSites    int
	SitesWithError   map[string]bool
	CachedSiteRanges map[string]int
	Errors           []*ImportError
	Md5sum string
}

// New creates a new import process
func (di *DatabaseImport) New(parser *Parser, uid int, databaseName string, langIsocode string, filehash string) error {
	var err error
	di.Database = &DatabaseInfos{}
	di.Uid = uid
	di.Database.Owner = di.Uid
	di.Database.Default_language = langIsocode
	di.Database.Geographical_extent_geom = "0103000020E610000001000000050000000C21E7FDFF7F66C01842CEFBFF7F56C00C21E7FDFF7F66C01842CEFBFF7F56400C21E7FDFF7F66401842CEFBFF7F56400C21E7FDFF7F66401842CEFBFF7F56C00C21E7FDFF7F66C01842CEFBFF7F56C0"
	di.CurrentSite = &model.SiteInfos{}
	di.Parser = parser
	di.NumberOfSites = 0
	di.SitesWithError = map[string]bool{}
	di.Md5sum = filehash

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
		return err
	}
	/* di.ArkeoCharacsIDs = map[int][]int{}
	di.ArkeoCharacsIDs, err = di.cacheCharacsIDs()
	if err != nil {
		return err
	}*/

	// Cache site range ids
	di.CachedSiteRanges = map[string]int{}

	// Field DATABASE_SOURCE_NAME
	if di.Database.Name == "" {
		if databaseName != "" {
			if di.Database.Init == false {
				if err = di.processDatabaseName(databaseName); err != nil {
					di.AddError(databaseName, "IMPORT.CSVFIELD_DATABASE_SOURCE_NAME.T_CHECK_INVALID", "DATABASE_SOURCE_NAME")
					return err
				}
				return err
			}
		} else {
			di.AddError(databaseName, "IMPORT.CSVFIELD_DATABASE_SOURCE_NAME.T_CHECK_EMPTY", "DATABASE_SOURCE_NAME")
			return errors.New("IMPORT.CSVFIELD_DATABASE_SOURCE_NAME.T_CHECK_EMPTY")
		}
	}

	return nil
}

// periodRegexp is used to match if starting and ending periods are valid
//var periodRegexp = regexp.MustCompile(`(-?\d{0,}):(-?\d{0,})`)
var validDateRexep = regexp.MustCompile(`^-?\d{0,}\p{L}{0,}:?-?\d{0,}\p{L}{0,}$`)
var uniqDateRegexp = regexp.MustCompile(`^(-?\d+\p{L}{0,})$`)
var periodRegexpDate1 = regexp.MustCompile(`^(-?\d{0,}\p{L}{0,}):-?\d{0,}\p{L}{0,}$`)
var periodRegexpDate2 = regexp.MustCompile(`^-?\d{0,}\p{L}{0,}:(-?\d{0,}\p{L}{0,})$`)

// setDefaultValues init the di.Database object with default values
func (di *DatabaseImport) setDefaultValues() {
	di.Database.Scale_resolution = "undefined"
	di.Database.Geographical_extent = "undefined"
	di.Database.Type = "undefined"
	di.Database.Source_description = ""
	di.Database.Editor = ""
	di.Database.Contributor = ""
	di.Database.State = "undefined"
	di.Database.Published = false
	di.Database.License_id = 0
	di.Database.Declared_creation_date = time.Now()
	di.Database.Created_at = time.Now()
	di.Database.Updated_at = time.Now()
}

// ProcessRecord is triggered for each line of csv
func (di *DatabaseImport) ProcessRecord(f *Fields) {

	//fmt.Println(di.Parser.Line, " - ", f)
	var err error

	// if site id not set and no previous SITE_SOURCE_ID is set, produce an error
	if di.CurrentSite.Code == "" && f.SITE_SOURCE_ID == "" {
		di.AddError("", "IMPORT.CSVFIELD_SITE_SOURCE_ID.T_CHECK_EMPTY", "SITE_SOURCE_ID")
		return
	}

	// If site code is not empty and differs, create a new instance of SiteInfos to store datas
	if f.SITE_SOURCE_ID != "" && f.SITE_SOURCE_ID != di.CurrentSite.Code {
		di.CurrentSite = &model.SiteInfos{}
		di.CurrentSiteRange = &model.Site_range{}
		di.CurrentSiteRangeCharac = &SiteRangeCharacInfos{}
		di.CurrentSite.Code = f.SITE_SOURCE_ID
		di.CurrentSite.Name = f.SITE_NAME
		di.CurrentSite.Database_id = di.Database.Id
		di.CurrentSite.Lang_isocode = di.Database.Default_language
		di.CurrentSite.Geom_3d = "POINT(0 0 0)"
		di.NumberOfSites++
		// Process site info
		di.processSiteInfos(f)
	} else {
		//di.CurrentSite.NbSiteRanges++
		di.processSiteInfos(f)
		di.checkDifferences(f)
	}

	// Init the site range if necessary
	// if di.CurrentSite.NbSiteRanges == 0 {
	// }

	// Process site range infos
	di.processSiteRangeInfos(f)
	// Process chara infos
	di.processCharacInfos(f)

	// If no error insert site in database
	if !di.CurrentSite.HasError {
		if di.CurrentSite.Id == 0 {
			err = di.CurrentSite.Create(di.Tx)
			// Site ID
			di.CurrentSiteRange.Site_id = di.CurrentSite.Id
			//} else {
			//	err = di.CurrentSite.Update(di.Tx)
		}
		if err == nil {
			err = di.insertSiteRangeInfos()
			if err == nil {
				err = di.insertCharacInfos()
			}
		}
		if err != nil {
			log.Println(err.Error())
			di.AddError("", err.Error(), "")
		}
	}

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
		return errors.New("Database already exists with same name and owned by another user.")
	}

	di.Database.Exists, err = di.Database.DoesExist(di.Tx)
	if err != nil {
		return err
	}

	return nil
}

// ProcessEssentialDatabaseInfos store or update informations about database defined by user at step 1
func (di *DatabaseImport) ProcessEssentialDatabaseInfos(name string, geographicalExtent string, selectedContinents []int, selectedCountries []int) error {
	var err error
	if di.Database.Exists {
		// Cache infos received from web form
		// Get database infos
		err = di.Database.Get(di.Tx)
		if err != nil {
			return err
		}
		// fmt.Println(di.Database)

		// Delete linked continents
		err = di.Database.DeleteContinents(di.Tx)
		if err != nil {
			return err
		}

		// Delete linked countries
		err = di.Database.DeleteCountries(di.Tx)
		if err != nil {
			return err
		}

		// Delete linked sites
		di.Database.DeleteSites(di.Tx)

	} else {
		di.setDefaultValues()
	}

	di.Database.Name = name
	di.Database.Geographical_extent = geographicalExtent
	di.Database.Init = true

	if di.Database.Exists {
		// Update record
		err = di.Database.Update(di.Tx)
		if err != nil {
			return err
		}
		di.Database.DeleteAuthors(di.Tx)
		a := []int{di.Uid}
		err = di.Database.SetAuthors(di.Tx, a)
	} else {
		// Create record
		err = di.Database.Create(di.Tx)
		if err != nil {
			return err
		}
		a := []int{di.Uid}
		err = di.Database.SetAuthors(di.Tx, a)
	}
	if err != nil {
		return err
	}

	if len(selectedCountries) > 0 {
		di.Database.Countries = selectedCountries
		err = di.Database.AddCountries(di.Tx, selectedCountries)
		if err != nil {
			return err
		}
	}

	if len(selectedContinents) > 0 {
		di.Database.Continents = selectedContinents
		err = di.Database.AddContinents(di.Tx, selectedContinents)
		if err != nil {
			return err
		}
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
	val, err := di.valueAsBool("CITY_CENTROID", f.CITY_CENTROID)
	if err == nil {
		di.CurrentSite.Centroid = val
	}

	// If only one of lat or lon empty
	if (f.LATITUDE != "" && f.LONGITUDE == "") || (f.LONGITUDE != "" && f.LATITUDE == "") {
		di.AddError(f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_LATITUDE_OR_LONGITUDE.T_CHECK_ONE_IS_EMPTY_OTHER_NOT", "LATITUDE", "LONGITUDE")
	} else {
		// If lat and lon not empty, process geo datas
		if f.LATITUDE != "" && f.LONGITUDE != "" {
			skip := false
			point, err := di.processGeoDatas(f)
			if err == nil {
				di.CurrentSite.Point = point
				// Store lat and lon to check differences if site has multiple site ranges
				di.CurrentSite.Latitude = f.LATITUDE
				di.CurrentSite.Longitude = f.LONGITUDE
				di.CurrentSite.Altitude = f.ALTITUDE
				if strings.Contains(f.LATITUDE, ",") {
					di.AddError(f.LATITUDE, "IMPORT.CSVFIELD_GEOMETRY.T_COMMA_DETECTED", "LATITUDE")
					skip = true
				}
				if strings.Contains(f.LONGITUDE, ",") {
					di.AddError(f.LONGITUDE, "IMPORT.CSVFIELD_GEOMETRY.T_COMMA_DETECTED", "LONGITUDE")
					skip = true
				}
				if strings.Contains(f.ALTITUDE, ",") {
					di.AddError(f.ALTITUDE, "IMPORT.CSVFIELD_GEOMETRY.T_COMMA_DETECTED", "ALTITUDE")
					skip = true
				}
				if !skip {
					di.CurrentSite.Geom = di.CurrentSite.Point.ToEWKT_2d()
					if di.CurrentSite.Altitude != "" {
						di.CurrentSite.Geom_3d = di.CurrentSite.Point.ToEWKT()
					}
				}
			} else {
				log.Println("databaseimport.go:", err)
			}
		} else {
			// User don't want to use Geonames, we are stuck
			if !di.Parser.UserChoices.UseGeonames {
				di.AddError(f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_GEO.T_CHECK_LAT_OR_LON_NOT_SET_AND_NO_GEONAMES", "LATITUDE", "LONGITUDE", "GEONAME_ID")
			} else {
				// If user chose to use Geonames, and we don't have valid coordinates at this point, use geonames functionality
				di.CurrentSite.GeonameID = f.GEONAME_ID
				point, err := di.processGeonames(f)
				if err == nil {
					di.CurrentSite.Point = point
					// Has we used Geonames, site location type is "centroid"
					di.CurrentSite.Centroid = true
					di.CurrentSite.Geom = di.CurrentSite.Point.ToEWKT_2d()
					di.CurrentSite.Geom_3d = di.CurrentSite.Point.ToEWKT()
					if err != nil {
						log.Println("databaseimport.go:", err)
						di.AddError(f.GEONAME_ID, "IMPORT.CSVFIELD_GEO.T_CHECK_LAT_OR_LON_NOT_SET_AND_NO_GEONAMES", "GEONAME_ID")
					}
				}
			}
		}
	}

	// OCCUPATION
	if f.OCCUPATION == "" {
		di.AddError("", "IMPORT.CSVFIELD_ALL.T_CHECK_UNDEFINED", "OCCUPATION")
	} else {
		val, err := di.getOccupation(f.OCCUPATION)
		if err == nil {
			di.CurrentSite.Occupation = val
		}
	}

}

// checkDifferences verifies if values entered for the site are identical
func (di *DatabaseImport) checkDifferences(f *Fields) {

	// MAIN_CITY_NAME
	if f.MAIN_CITY_NAME != "" && di.CurrentSite.City_name != f.MAIN_CITY_NAME {
		di.AddError(f.MAIN_CITY_NAME, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "MAIN_CITY_NAME")
	}

	// CITY_CENTROID
	val, err := di.valueAsBool("CITY_CENTROID", f.CITY_CENTROID)
	if err == nil && val != di.CurrentSite.Centroid {
		di.AddError(f.CITY_CENTROID, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "CITY_CENTROID")
	}

	// LONGITUDE
	if f.LONGITUDE != "" && f.LONGITUDE != di.CurrentSite.Longitude {
		di.AddError(f.LONGITUDE, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "LONGITUDE")
	}

	// LATITUDE
	if f.LATITUDE != "" && f.LATITUDE != di.CurrentSite.Latitude {
		di.AddError(f.LATITUDE, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "LATITUDE")
	}

	// ALTITUDE
	if f.ALTITUDE != "" && f.ALTITUDE != di.CurrentSite.Altitude {
		di.AddError(f.LATITUDE, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "ALTITUDE")
	}

	// GEONAME ID
	if f.GEONAME_ID != "" && f.GEONAME_ID != di.CurrentSite.GeonameID {
		di.AddError(f.GEONAME_ID, "IMPORT.CSVFIELD_ALL.T_CHECK_ALREADY_DEFINED_VALUE_DIFFERS", "GEONAME_ID")
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

	switch cleanAndLower(occupation) {
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

// processGeoDatas analyzes and process csv fields related to geo informations
func (di *DatabaseImport) processGeoDatas(f *Fields) (*geo.Point, error) {
	var point *geo.Point
	var epsg int
	var err error
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
	di.CurrentSite.EPSG = epsg

	// Parse LONGITUDE
	lon, err := strconv.ParseFloat(strings.Replace(f.LONGITUDE, ",", ".", 1), 64)
	if err != nil {
		di.AddError(f.LONGITUDE, "IMPORT.CSVFIELD_LONGITUDE.T_CHECK_INCORRECT_VALUE", "LONGITUDE")
		hasError = true
	}
	// Parse LATITUDE
	lat, err := strconv.ParseFloat(strings.Replace(f.LATITUDE, ",", ".", 1), 64)
	if err != nil {
		di.AddError(f.LATITUDE, "IMPORT.CSVFIELD_LATITUDE.T_CHECK_INCORRECT_VALUE", "LATITUDE")
		hasError = true
	}
	// Parse ALTITUDE
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
		hasError = true
	}

	if hasError {
		di.AddError(f.PROJECTION_SYSTEM+" "+f.LONGITUDE+" "+f.LATITUDE, "IMPORT.CSVFIELD_GEO.T_ERROR_UNABLE_TO_CREATE_GEOMETRY", "EPSG", "LATITUDE", "LONGITUDE")
		if err != nil {
			return nil, err
		} else {
			return nil, errors.New("Error parsing coordinates")
		}
	}
	return point, nil
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
	return point, nil
}

func (di *DatabaseImport) processSiteRangeInfos(f *Fields) {

	// STARTING_PERIOD
	// fmt.Println("Starting period", f.STARTING_PERIOD)
	startingDates, err := di.parseDates(f.STARTING_PERIOD)
	if err != nil {
		di.AddError(f.STARTING_PERIOD, "IMPORT.CSVFIELD_STARTING_PERIOD.T_CHECK_INVALID", "STARTING_PERIOD")
	} else {
		di.CurrentSiteRange.Start_date1 = startingDates[0]
		di.CurrentSiteRange.Start_date2 = startingDates[1]
	}
	// fmt.Println("Parsed dates", dates)
	// fmt.Println("Start_date1", di.CurrentSiteRange.Start_date1)
	// fmt.Println("Start_date2", di.CurrentSiteRange.Start_date2)

	// ENDING_PERIOD
	endingDates, err := di.parseDates(f.ENDING_PERIOD)
	if err != nil {
		di.AddError(f.ENDING_PERIOD, "IMPORT.CSVFIELD_ENDING_PERIOD.T_CHECK_INVALID", "ENDING_PERIOD")
	} else {
		di.CurrentSiteRange.End_date1 = endingDates[0]
		di.CurrentSiteRange.End_date2 = endingDates[1]
	}

}
func (di *DatabaseImport) insertSiteRangeInfos() error {

	// If site range is not cached, create it
	siteRangeHash := strconv.Itoa(di.CurrentSite.Id) + strconv.Itoa(di.CurrentSiteRange.Start_date1) + strconv.Itoa(di.CurrentSiteRange.Start_date2) + strconv.Itoa(di.CurrentSiteRange.End_date1) + strconv.Itoa(di.CurrentSiteRange.End_date2)

	if id, ok := di.CachedSiteRanges[siteRangeHash]; !ok {
		err := di.CurrentSiteRange.Create(di.Tx)
		if err != nil {
			di.AddError("", "IMPORT.PROCESS_SITE_RANGE.T_ERROR", "")
			return err
		}
		di.CachedSiteRanges[siteRangeHash] = di.CurrentSiteRange.Id
		return nil
	} else {
		di.CurrentSiteRange.Id = id
	}

	return nil
}

// processCharacs analyses the fields of each charac for each level
// It verify if charac of any level exists and if true, assign it to the site range
func (di *DatabaseImport) processCharacInfos(f *Fields) error {
	path := ""
	lvl := 1
	if f.CARAC_NAME == "" {
		di.AddError("", "IMPORT.CSVFIELD_CARAC_NAME.T_CHECK_EMPTY", "CARAC_NAME")
		return errors.New("invalid carac name")
	}
	if f.CARAC_LVL1 != "" {
		path += "->" + cleanAndLower(f.CARAC_LVL1)
	} else {
		di.AddError("", "IMPORT.CSVFIELD_CARAC_LVL1.T_CHECK_EMPTY", "CARAC_LVL1")
		return errors.New("no lvl1 carac")
	}
	if f.CARAC_LVL2 != "" {
		path += "->" + cleanAndLower(f.CARAC_LVL2)
		lvl++
	}
	if f.CARAC_LVL3 != "" {
		path += "->" + cleanAndLower(f.CARAC_LVL3)
		lvl++
	}
	if f.CARAC_LVL4 != "" {
		path += "->" + cleanAndLower(f.CARAC_LVL4)
		lvl++
	}
	//path = strings.TrimSuffix(path, "->")
	// Check if charac exists and retrieve id
	caracNameToLowerCase := cleanAndLower(f.CARAC_NAME)
	caracID := di.ArkeoCharacs[caracNameToLowerCase][caracNameToLowerCase+path]
	if caracID == 0 {
		log.Println("NOT FOUND: ", caracNameToLowerCase+path)
		di.AddError(caracNameToLowerCase+path, "IMPORT.CSVFIELD_CARACTERISATION.T_CHECK_INVALID", "CARAC_LVL"+strconv.Itoa(lvl))
		return errors.New("invalid charac")
	}
	/*
		cs := di.ArkeoCharacsIDs[caracID]
		if len(cs) == 0 {
			di.AddError(caracNameToLowerCase+path, "IMPORT.CSVFIELD_CARACTERISATION.T_CHECK_INVALID", "CARAC_LVL"+strconv.Itoa(lvl))
			return errors.New("invalid charac")
		}
		fmt.Println(cs)
	*/

	//	STATE_OF_KNOWLEDGE
	switch cleanAndLower(f.STATE_OF_KNOWLEDGE) {
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_NOT_DOCUMENTED"):
		di.CurrentSiteRangeCharac.Knowledge_type = "not_documented"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_LITERATURE"):
		di.CurrentSiteRangeCharac.Knowledge_type = "literature"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_AERIAL"):
		di.CurrentSiteRangeCharac.Knowledge_type = "prospected_aerial"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_PROSPECTED_PEDESTRIAN"):
		di.CurrentSiteRangeCharac.Knowledge_type = "prospected_pedestrian"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_SURVEYED"):
		di.CurrentSiteRangeCharac.Knowledge_type = "surveyed"
	case di.lowerTranslation("IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_LABEL_DIG"):
		di.CurrentSiteRangeCharac.Knowledge_type = "dig"
	default:
		if f.STATE_OF_KNOWLEDGE == "" {
			di.AddError(f.STATE_OF_KNOWLEDGE, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_CHECK_EMPTY", "STATE_OF_KNOWLEDGE")
		} else {
			di.AddError(f.STATE_OF_KNOWLEDGE, "IMPORT.CSVFIELD_STATE_OF_KNOWLEDGE.T_CHECK_INVALID", "STATE_OF_KNOWLEDGE")
		}
		return errors.New("Bad value for knowledge type")
	}

	// EXCEPTIONAL
	/*
		switch strings.ToLower(f.CARAC_EXP) {
		case di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_LABEL_YES"):
			di.CurrentSiteRangeCharac.Exceptional = true
		case di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_LABEL_NO"):
			di.CurrentSiteRangeCharac.Exceptional = false
		default:
			if f.CARAC_EXP == "" {
				di.AddError(f.CARAC_EXP, "IMPORT.CSVFIELD_CARAC_EXP.T_CHECK_EMPTY", "CARAC_EXP")
			} else {
				di.AddError(f.CARAC_EXP, "IMPORT.CSVFIELD_CARAC_EXP.T_CHECK_INVALID", "CARAC_EXP")
			}
			return errors.New("Bad value for exceptional")
		}
	*/

	val, err := di.valueAsBool("CARAC_EXP", f.CARAC_EXP)
	if err == nil {
		di.CurrentSiteRangeCharac.Exceptional = val
	}

	// BIBLIOGRAPHY
	di.CurrentSiteRangeCharac.Bibliography = f.BIBLIOGRAPHY

	// COMMENTS
	di.CurrentSiteRangeCharac.Comment = f.COMMENTS

	// Set current charac id to be linked
	di.CurrentSiteRangeCharac.Charac_id = caracID

	// Set site range id to be linked
	di.CurrentSiteRangeCharac.Site_range_id = di.CurrentSiteRange.Id
	return nil

}

// cacheCharacs get all Characs from database and cache them
func (di *DatabaseImport) cacheCharacs() (map[string]map[string]int, error) {
	characs := map[string]map[string]int{}
	characsRoot, err := model.GetAllCharacsRootFromLangIsocode(di.Database.Default_language)
	if err != nil {
		return characs, err
	}
	for name := range characsRoot {
		loweredName := cleanAndLower(name)
		characs[loweredName], err = model.GetCharacPathsFromLangID(name, di.Database.Default_language)
		if err != nil {
			return characs, err
		}
	}
	return characs, nil
}

// cacheCharacsIDs get all Characs Ids from database and cache them
func (di *DatabaseImport) cacheCharacsIDs() (map[int][]int, error) {
	characs := map[int][]int{}
	c, err := model.GetAllCharacPathIDsFromLangIsocode(di.Database.Default_language)
	if err != nil {
		return characs, err
	}
	for id, path := range c {
		// Split path
		aIDs := []int{}
		for _, cid := range strings.Split(path, "->") {
			i, err := strconv.Atoi(cid)
			if err != nil {
				return characs, err
			}
			aIDs = append(aIDs, i)
		}
		characs[id] = aIDs
	}
	return characs, nil
}

func (di *DatabaseImport) insertCharacInfos() error {
	var err error
	di.CurrentSiteRangeCharac.Site_range_id = di.CurrentSiteRange.Id
	stmt, err := di.Tx.PrepareNamed("INSERT INTO \"site_range__charac\" (" + model.Site_range__charac_InsertStr + ") VALUES (" + model.Site_range__charac_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = stmt.Get(&di.CurrentSiteRangeCharac.Site_range__charac_id, di.CurrentSiteRangeCharac)
	if err != nil {
		return err
	}

	di.CurrentSiteRangeCharac.Lang_isocode = di.Database.Default_language
	_, err = di.Tx.NamedExec("INSERT INTO \"site_range__charac_tr\" (\"site_range__charac_id\", \"lang_isocode\", \"bibliography\", \"comment\") VALUES (:site_range__charac_id, :lang_isocode, :bibliography, :comment)", di.CurrentSiteRangeCharac)
	return err
}

// valueAsBool analyses YES/NO translatable values to bool
func (di *DatabaseImport) valueAsBool(fieldName, val string) (choosenValue bool, err error) {
	switch cleanAndLower(val) {
	case di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_LABEL_YES"):
		choosenValue = true
	case di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_LABEL_NO"):
		choosenValue = false
	default:
		if val == "" {
			di.AddError(val, "IMPORT.CSVFIELD_ALL.T_CHECK_EMPTY", fieldName)
		} else {
			di.AddError(val, "IMPORT.CSVFIELD_ALL.T_CHECK_INVALID", fieldName)
		}
		return choosenValue, errors.New("Bad value for " + fieldName)
	}
	return choosenValue, nil
}

// lowerTranslation return translation in lower case
func (di *DatabaseImport) lowerTranslation(s string) string {
	return cleanAndLower(translate.T(di.Parser.Lang, s))
}

// parseDates analyzes declared period and returns starting and ending dates
func (di *DatabaseImport) parseDates(period string) ([2]int, error) {
	// If empty period, set "min and max" dates
	period = strings.Replace(period, "+", "", -1)
	period = strings.Replace(period, " ", "", -1) // non breaking space
	period = strings.Replace(period, " ", "", -1)

	// fmt.Println("PERIOD", period)

	if (period == "" || cleanAndLower(period) == di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED")) || cleanAndLower(period) == "null" {
		return [2]int{math.MinInt32, math.MaxInt32}, nil
	}

	if !validDateRexep.MatchString(period) {
		return [2]int{0, 0}, errors.New("Invalid period")
	}

	var dates [2]int
	var date1 string
	var date2 string

	// If we have only a date, start date and end date are the same
	uniqDate := uniqDateRegexp.FindString(period)
	if uniqDate != "" {
		//fmt.Println("UNIQ DATE")
		ud, err := strconv.ParseInt(uniqDate, 10, 64)
		if err != nil {
			return [2]int{0, 0}, errors.New("Invalid period")
		}
		dates[0] = int(ud)
		dates[1] = int(ud)
	} else {
		// fmt.Println("MATCH ?")
		mdate1 := periodRegexpDate1.FindStringSubmatch(period)
		mdate2 := periodRegexpDate2.FindStringSubmatch(period)
		if len(mdate1) == 0 || len(mdate2) == 0 {
			return [2]int{0, 0}, errors.New("Invalid period")
		}
		tmpDate1 := mdate1[1]
		tmpDate2 := mdate2[1]

		// If tmpDate1 is empty, set min date
		if tmpDate1 == "" {
			dates[0] = math.MinInt32
		} else {
			// Check if it is numeric
			date1 = uniqDateRegexp.FindString(tmpDate1)
			if date1 != "" {
				// If not a date check if it is undefined
				sd, err := strconv.ParseInt(tmpDate1, 10, 64)
				if err != nil {
					di.AddError(period, "IMPORT.CSVFIELD_PERIOD_DATE1.T_CHECK_WRONG_VALUE", "STARTING_PERIOD")
				} else {
					dates[0] = int(sd)
				}
			} else {
				// Check if it is set explicitly as undefined
				if cleanAndLower(tmpDate1) == di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED") {
					dates[0] = math.MinInt32
				} else {
					di.AddError(period, "IMPORT.CSVFIELD_PERIOD_DATE2.T_CHECK_WRONG_VALUE", "STARTING_PERIOD")
				}
			}
		}

		// If tmpDate2 is empty, set max date
		if tmpDate2 == "" {
			// dbg
			dates[1] = math.MaxInt32
		} else {
			// Check if it is numeric
			date2 = uniqDateRegexp.FindString(tmpDate2)
			if date2 != "" {
				// If not a date check if it is undefined
				ed, err := strconv.ParseInt(tmpDate2, 10, 64)
				if err != nil {
					di.AddError(period, "IMPORT.CSVFIELD_PERIOD_DATE2.T_CHECK_WRONG_VALUE", "ENDING_PERIOD")
				} else {
					dates[1] = int(ed)
				}
			} else {
				// Check if it is set explicitly as undefined
				if cleanAndLower(tmpDate2) == di.lowerTranslation("IMPORT.CSVFIELD_ALL.T_CHECK_UNDETERMINED") {
					dates[1] = math.MaxInt32
				} else {
					di.AddError(period, "IMPORT.CSVFIELD_PERIOD_DATE2.T_CHECK_WRONG_VALUE", "ENDING_PERIOD")
				}
			}
		}

		// Arkeogis hack on negative dates
		if dates[0] < 1 && dates[0] != math.MinInt32 {
			dates[0] += 1
		}

		if dates[1] < 1 && dates[1] != math.MinInt32 {
			dates[1] += 1
		}

	}

	// fmt.Println("Date1 and Date2", dates, "---")
	// fmt.Println("----")

	return dates, nil
}

func (di *DatabaseImport) Save(filename string) (int, error) {
	var err error
	i := model.Import{Database_id: di.Database.Id, User_id: di.Uid, Filename: filename, Number_of_lines: di.Parser.Line - 1, Md5sum: di.Md5sum}
	err = i.Create(di.Tx)
	return i.Id, err
}

func cleanAndLower(s string) string {
	s = strings.Replace(s, " ", "", -1) // non breaking space
	s = strings.Replace(s, "–", "-", -1)
	s = strings.TrimSpace(s)
	return strings.ToLower(s)
}
