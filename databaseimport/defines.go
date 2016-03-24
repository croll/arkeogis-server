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

package databaseimport

// Fields structure list all the fields we can have in the import file
// It is used by the Parse() function to validate the file and structure the datas for processing
type Fields struct {
	DATABASE_SOURCE_NAME string
	SITE_SOURCE_ID       string
	SITE_NAME            string
	MAIN_CITY_NAME       string
	GEONAME_ID           string
	PROJECTION_SYSTEM    string
	LONGITUDE            string
	LATITUDE             string
	ALTITUDE             string
	STATE_OF_KNOWLEDGE   string
	CITY_CENTROID        string
	OCCUPATION           string
	STARTING_PERIOD      string
	ENDING_PERIOD        string
	CARAC_NAME           string
	CARAC_LVL1           string
	CARAC_LVL2           string
	CARAC_LVL3           string
	CARAC_LVL4           string
	CARAC_EXP            string
	BIBLIOGRAPHY         string
	COMMENTS             string
}

var (
	mandatoryCsvColumns = map[string]bool{"SITE_SOURCE_ID": false, "CARAC_NAME": false, "CARAC_LVL1": false, "CARAC_LVL2": false, "CARAC_LVL3": false, "CARAC_LVL4": false, "CARAC_EXP": false, "MAIN_CITY_NAME": false, "CITY_CENTROID": false, "STATE_OF_KNOWLEDGE": false, "OCCUPATION": false, "SITE_NAME": false, "STARTING_PERIOD": false, "ENDING_PERIOD": false, "BIBLIOGRAPHY": false, "COMMENTS": false, "GEONAME_ID": false}
	//mandatoryFields       = [11]string{"SITE_SOURCE_ID", "DATABASE_SOURCE_NAME", "MAIN_CITY_NAME", "CITY_CENTROID", "STATE_OF_KNOWLEDGE", "OCCUPATION"}
	//	mandatoryFields = map[string]bool{"SITE_SOURCE_ID": false, "DATABASE_SOURCE_NAME": false, "MAIN_CITY_NAME": false, "CITY_CENTROID": false, "STATE_OF_KNOWLEDGE": false, "OCCUPATION": false}
	//mandatoryHeaderFields = [2]string{"SITE_NAME", "BIBLIOGRAPHY", "COMMENTS"}
	//	mandatoryHeaderFields = map[string]bool{"SITE_NAME": false, "BIBLIOGRAPHY": false, "COMMENTS": false}
	//optionalFields        = [4]string{"START_DATE_QUALIFIER", "START_DATE", "END_DATE_QUALIFIER", "END_DATE"}
	//optionalFields = map[string]bool{"SITE_NAME": false, "START_DATE_QUALIFIER": false, "START_DATE": false, "END_DATE_QUALIFIER": false, "END_DATE": false, "BIBLIOGRAPHY": false, "COMMENTS": false}
	//conditionalColumns     = [6]string{"GEONAME_ID", "CITY_CENTROID"}
	geonamesColumns = map[string]bool{"GEONAME_ID": false, "CITY_CENTROID": false}
	//conditionalGeoFields  = [4]string{"PROJECTION_SYSTEM", "LONGITUDE", "LATITUDE", "ALTITUDE"}
	geoFieldsColumns = map[string]bool{"PROJECTION_SYSTEM": false, "LONGITUDE": false, "LATITUDE": false, "ALTITUDE": false}
	//caracterisationsColumns = [5]string{"REAL_ESTATE_LVL1", "FURNITURE_LVL1", "PRODUCTION_LVL1", "LANDSCAPE_LVL1", "TEXT_ICONOGRAPHY_LVL1"}
)
