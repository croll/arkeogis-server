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
	//"strings"
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

func InteroperableExportXml(tx *sqlx.Tx, w io.Writer, databaseId int, lang string) (err error) {
	type StringL struct {
		Content				string		`xml:",innerxml"`
		Lang				string		`xml:"xml:lang,attr"`
	}
	type Metadata struct {
		XMLName   			xml.Name 	`xml:"metadata"`
		Xmlns				string   	`xml:"xmlns,attr"`
		Xmlnsxsi			string   	`xml:"xmlns:xsi,attr"`
		XsischemaLocation	string   	`xml:"xsi:schemaLocation,attr"`
		Xmlnsdc				string		`xml:"xmlns:dc,attr"`
		Xmlnsdcterms		string		`xml:"xmlns:dcterms,attr"`

		DcTitle			    string		`xml:"dc:title"`
		DcCreator			[]string	`xml:"dc:creator"`
		DcSubject			[]string	`xml:"dc:subject"`
		DcDescription		[]StringL		`xml:"dc:description"`
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

	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(v); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	return nil
}


