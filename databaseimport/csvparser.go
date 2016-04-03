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

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/translate"
	//"strconv"
	"strings"
	"unicode/utf8"
)

// Handle errors
type ParserError struct {
	Columns []string `json:"columns"`
	ErrMsg  string   `json:"errMsg"`
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("columns %s: %s", strings.Join(e.Columns, ","), e.ErrMsg)
}

// Parser type holds the informations about the parsed file and provides functions to parse and store the datas in ArkeoGIS database
type Parser struct {
	Filename     string
	HeaderFields map[int]string
	UserChoices  UserChoices
	Line         int
	Lang         string
	Reader       *csv.Reader
	Errors       []*ParserError
}

func (p *Parser) AddError(errMsg string, columns ...string) {
	p.Errors = append(p.Errors, &ParserError{
		Columns: columns,
		ErrMsg:  translate.T(p.Lang, errMsg),
	})
}

func (p *Parser) HasError() bool {
	if len(p.Errors) > 0 {
		return true
	}
	return false
}

// NewParser open csv file and return a  *Parser
func NewParser(filename string, lang int) (*Parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	// Get lang by id
	var langIsoCode string
	err = db.DB.QueryRow("SELECT iso_code FROM lang WHERE id = $1", lang).Scan(&langIsoCode)
	if err != nil {
		return nil, err
	}

	p := &Parser{
		Line:        1,
		Filename:    filename,
		UserChoices: UserChoices{UseGeonames: false},
		Lang:        langIsoCode,
	}
	p.Reader = csv.NewReader(f)
	p.Reader.Comma = ';'
	return p, nil
}

// SetUserChoices is used to configure parser
func (p *Parser) SetUserChoices(item string, val bool) {
	reflect.Indirect(reflect.ValueOf(&p.UserChoices)).FieldByName(item).SetBool(val)
}

// Parse is the entry point of the parsing process
func (p *Parser) CheckHeader() error {
	// Variable to store database status
	record, err := p.Reader.Read()
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}
	// Parse only first line
	err = p.checkHeader(record)
	if err != nil {
		return err
	}
	return nil
}

// Parse is the entry point of the parsing process
func (p *Parser) Parse(fn func(r *Fields)) error {
	// Check encoding

	// Variable to store database status
	f := Fields{}
	r := reflect.ValueOf(&f)
	p.Line = 2
	for {
		record, err := p.Reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Parse lines after first line
		// Ok we assign fields vlues to struct Fields
		for k, v := range record {
			// utf8 validation
			if !utf8.ValidString(record[0]) {
				p.AddError("IMPORT.CSV_FILE.T_CHECK_NOT_UTF8_ENCODING", record[0])
			}
			r.Elem().FieldByName(p.HeaderFields[k]).SetString(v)
		}
		// Process line
		fn(&f)
		p.Line++
	}
	if len(p.Errors) > 0 {
		return errors.New("IMPORT.CSV_FILE.T_CHECK_ERRORS_DETECTED")
	}
	return nil
}

// checkHeader analyzes the first line of csv to check if fields names correspond to fields of struct Fields
// If not, trigger and error and exit
func (p *Parser) checkHeader(record []string) error {
	f := Fields{}
	// Store if we found header file witch defines lvl1 for a charac like furniture, realestate, etc
	p.HeaderFields = make(map[int]string)
	r := reflect.Indirect(reflect.ValueOf(&f))
	for k, v := range record {
		v = strings.TrimSpace(v)
		// Check if field name found in csv exists in Fields struct definition
		if !r.FieldByName(v).IsValid() {
			p.AddError("IMPORT.CSV_FILE.T_CHECK_UNRECOGNIZED_FIELD", v)
		} else {
			// Store detected header column
			p.HeaderFields[k] = v
			// Store if field is found in csv file
			if val, ok := mandatoryCsvColumns[v]; ok && val == false {
				mandatoryCsvColumns[v] = true
			}
			// If user wants to use geofields, check if fields are found in csv
			if p.UserChoices.UseGeonames == true {
				if val, ok := geonamesColumns[v]; ok && val == false {
					geonamesColumns[v] = true
				}
			}
		}
	}

	// Verify if all mandatory csv fields are found
	for name, set := range mandatoryCsvColumns {
		if set == false {
			p.AddError("IMPORT.CSVFIELD_ALL.T_CHECK_NOT_FOUND", name)
		}
	}

	// If user choose to use geonames, verify if mandatory csv fields are found
	if p.UserChoices.UseGeonames == true {
		for name, set := range geonamesColumns {
			if set == false {
				p.AddError("IMPORT.CSVFIELD_GEONAME.T_CHECK_MANDATORY_FIELDS_NOT_FOUND", name)
			}
		}
	}

	if len(p.Errors) > 0 {
		return errors.New("Error found when analysing header fields")
	}

	return nil
}
