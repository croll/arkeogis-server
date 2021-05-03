/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2019 CROLL SAS
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

 package export

 import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
 	model "github.com/croll/arkeogis-server/model"
	//"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
	"encoding/xml"
	"encoding/json"
	"io"
	"errors"
)

/*
  xmlns="http://www.w3.org/1999/xhtml"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://www.w3.org/1999/xhtml http://www.w3.org/1999/xhtml.xsd"
  xmlns:dc="http://purl.org/dc/elements/1.1/" 
  xmlns:dcterms="http://purl.org/dc/terms/">
  */

type StringL struct {
	Content				string		`xml:",innerxml"`
	Lang				string		`xml:"xml:lang,attr"`
}

func readMappedToStringL(mapped map[string]string) []StringL {
	var strings []StringL
	for l, s := range mapped {
		strings = append(strings, StringL{s, l})
	}
	return strings
}

type XsiTyped struct {
	Content				string		`xml:",innerxml"`
	Xsitype				string		`xml:"xsi:type,attr,omitempty"`
	Lang				string		`xml:"xml:lang,attr,omitempty"`
}

type Geom struct {
	Type				string
	Coordinates			[][][]float64
}

func miniBounds(geom Geom) (north float64, south float64, east float64, west float64) {
	north = -math.MaxFloat64
	south = math.MaxFloat64
	west = -math.MaxFloat64
	east = math.MaxFloat64

	for _, obj := range geom.Coordinates {
		for _, point := range obj {
			if point[1] > north {
				north = point[1]
			} 
			if point[1] < south {
				south = point[1]
			} 
			if point[0] > west {
				west = point[0]
			} 
			if point[0] < east {
				east = point[0]
			} 
		}	
	}
	return north, south, east, west
}

func InteroperableExportXml(tx *sqlx.Tx, w io.Writer, databaseId int, lang string) (err error) {
	type Metadata struct {
		XMLName   			xml.Name 		`xml:"metadata"`
		Xmlns				string   		`xml:"xmlns,attr"`
		Xmlnsxsi			string   		`xml:"xmlns:xsi,attr"`
		XsischemaLocation	string   		`xml:"xsi:schemaLocation,attr"`
		Xmlnsdc				string			`xml:"xmlns:dc,attr"`
		Xmlnsdcterms		string			`xml:"xmlns:dcterms,attr"`

		DcTitle			    string			`xml:"dc:title"`
		DcCreator			[]string		`xml:"dc:creator"`
		DcSubject			[]StringL		`xml:"dc:subject"`
		DcDescription		[]StringL		`xml:"dc:description"`
		DcPublishers		[]XsiTyped	    `xml:"dc:publisher"`
		DcContributors      []string		`xml:"dc:contributor"`
		DcDate              XsiTyped        `xml:"dc:date"`
		DctermsIssued       XsiTyped        `xml:"dcterms:issued"`
		DctermsModified     XsiTyped        `xml:"dcterms:modified"`
		DcType			    XsiTyped		`xml:"dc:type"`   // @TODO: check if this is ok
		DcFormat			string			`xml:"dc:format"` // @TODO: check if this is ok
		DcIdentifier        []XsiTyped		`xml:"dc:identifier"`
		DcBibliographicCitation			    []StringL		`xml:"dc:bibliographicCitation"`
		DcSource	        XsiTyped		`xml:"dc:source,omitempty"`
		DcRelation			[]string		`xml:"dc:relation"`
		DcLanguage			XsiTyped		`xml:"dc:language"`
		DcTermsConformsTo   []XsiTyped	    `xml:"dcterms:conformsTo"` // @TODO: check if this is ok
		DcCoverage			[]XsiTyped		`xml:"dc:coverage"`
		DcTermsSpatial		XsiTyped		`xml:"dcterms:spatial,omitempty"`
		//Dc			    string		`xml:"dc:"`


	}

	d := model.Database{}
	d.Id = databaseId

	dbInfos, err := d.GetFullInfos(tx, lang)

	if err != nil {
		log.Println("Error getting database infos", err)
		tx.Rollback()
		return err
	}

	log.Printf("%+v\n", dbInfos)

	v := &Metadata{}
	v.Xmlns = "http://www.w3.org/1999/xhtml"
	v.Xmlnsxsi = "http://www.w3.org/2001/XMLSchema-instance"
	v.XsischemaLocation = "http://www.w3.org/1999/xhtml http://www.w3.org/1999/xhtml.xsd"
	v.Xmlnsdc = "http://purl.org/dc/elements/1.1/"
	v.Xmlnsdcterms = "http://purl.org/dc/terms/"

	v.DcTitle = dbInfos.Name
	v.DcCreator = dbInfos.GetAuthorsStrings()
	v.DcSubject = readMappedToStringL(dbInfos.Subject)
	v.DcDescription = readMappedToStringL(dbInfos.Description)
	v.DcPublishers = []XsiTyped{
		XsiTyped{dbInfos.Editor, "", ""},
		XsiTyped{dbInfos.Editor_url, "dcterms:URI", ""},
	}
	v.DcContributors = strings.Split(dbInfos.Contributor, ", ")
	v.DcDate = XsiTyped{dbInfos.Declared_creation_date.Format("2006-01-02"), "dcterms:W3CDTF", ""}
	v.DctermsIssued = XsiTyped{dbInfos.Created_at.Format("2006-01-02"), "dcterms:W3CDTF", ""}
	v.DctermsModified = XsiTyped{dbInfos.Updated_at.Format("2006-01-02"), "dcterms:W3CDTF", ""}
	v.DcType = XsiTyped{"dataset", "dcterms:DCMIType", ""}
	v.DcFormat = "text/csv"

	if len(dbInfos.Handles) > 0 {
		v.DcIdentifier = []XsiTyped{
			XsiTyped{dbInfos.Handles[0].Url, "dcterms:URI", ""},
		}
	}

	v.DcBibliographicCitation = readMappedToStringL(dbInfos.Bibliography)

	if source, ok := dbInfos.Source_description[dbInfos.Default_language]; ok {
		v.DcSource = XsiTyped{source, "dcterms:URI", ""}
	}

	if relations, ok := dbInfos.Source_relation[dbInfos.Default_language]; ok {
		// split using ','
		for _, relation := range strings.Split(relations, ",") {
			v.DcRelation = append(v.DcRelation, relation)
		}
	}

	langs := map[string]string{
		"fr": "fra",
		"de": "deu",
		"es": "spa",
		"en": "eng",
	}
	v.DcLanguage = XsiTyped{langs[dbInfos.Default_language], "dcterms:ISO639-3", ""}

	v.DcTermsConformsTo = []XsiTyped{
		XsiTyped{"https://www.frantiq.fr/pactols/le-thesaurus", "dcterms:URI", ""},
		XsiTyped{"https://epsg.org/", "dcterms:URI", ""},
		XsiTyped{"https://tools.ietf.org/id/draft-kunze-ark-21.html", "dcterms:URI", ""},
	}

	// if dbInfos.Geographical_extent == "country" {
	if len(dbInfos.Countries) > 0 {
		v.DcCoverage = append(v.DcCoverage, XsiTyped{"Pays", "", ""})
		v.DcCoverage = append(v.DcCoverage, XsiTyped{"https://www.geonames.org/"+strconv.Itoa(dbInfos.Countries[0].Geonameid), "dcterms:URI", ""})

	}

	// if dbInfos.Geographical_extent == "continent" {
		if len(dbInfos.Continents) > 0 {
		v.DcCoverage = append(v.DcCoverage, XsiTyped{"Continent", "", ""})
		v.DcCoverage = append(v.DcCoverage, XsiTyped{"https://www.geonames.org/"+strconv.Itoa(dbInfos.Continents[0].Geonameid), "dcterms:URI", ""})
	}

	// if dbInfos.Geographical_extent == "?" {
	//}

	for isocode, geolim := range dbInfos.Geographical_limit {
		v.DcCoverage = append(v.DcCoverage, XsiTyped{geolim, "", isocode})
	}

	log.Printf("Geographical_extent : %+v\n", dbInfos.Geographical_extent)
	log.Printf("Geographical_extent_geom : %+v\n", dbInfos.Geographical_extent_geom)

	if dbInfos.Geographical_extent_geom != "" {
		var geom Geom
		json.Unmarshal([]byte(dbInfos.Geographical_extent_geom), &geom)

		if (geom.Type != "Polygon") {
			return errors.New("geom not recognised for Geographical_extent_geom")
		}

		north, south, east, west := miniBounds(geom)
		northlimit := fmt.Sprintf("%f", north)
		eastlimit := fmt.Sprintf("%f", east)
		southlimit := fmt.Sprintf("%f", south)
		westlimit := fmt.Sprintf("%f", west)
		v.DcTermsSpatial = XsiTyped{"northlimit="+northlimit+";eastlimit="+eastlimit+";southlimit="+southlimit+";westlimit="+westlimit+";projection=EPSG4326;", "dcterms:Box", ""}
	}

	//v.DcCreator = strings.Split(dbInfos.Editor, ",")

	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n"))
	w.Write([]byte(`<?xml-stylesheet href="https://arkeogis.org/css/dataset-dublin-core.xsl" type="text/xsl"?>`+"\n"))

	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(v); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	return nil
}


