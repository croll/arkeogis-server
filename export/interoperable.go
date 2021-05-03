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
	//"math"
	//"strconv"
	"strings"
 	model "github.com/croll/arkeogis-server/model"
	//"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
	"encoding/xml"
	"io"
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
		XsiTyped{dbInfos.Editor, ""},
		XsiTyped{dbInfos.Editor_url, "dcterms:URI"},
	}
	v.DcContributors = strings.Split(dbInfos.Contributor, ", ")
	v.DcDate = XsiTyped{dbInfos.Declared_creation_date.Format("2006-01-02"), "dcterms:W3CDTF"}
	v.DctermsIssued = XsiTyped{dbInfos.Created_at.Format("2006-01-02"), "dcterms:W3CDTF"}
	v.DctermsModified = XsiTyped{dbInfos.Updated_at.Format("2006-01-02"), "dcterms:W3CDTF"}
	v.DcType = XsiTyped{"dataset", "dcterms:DCMIType"}
	v.DcFormat = "text/csv"

	if len(dbInfos.Handles) > 0 {
		v.DcIdentifier = []XsiTyped{
			XsiTyped{dbInfos.Handles[0].Url, "dcterms:URI"},
		}
	}

	v.DcBibliographicCitation = readMappedToStringL(dbInfos.Bibliography)

	if source, ok := dbInfos.Source_description[dbInfos.Default_language]; ok {
		v.DcSource = XsiTyped{source, "dcterms:URI"}
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


