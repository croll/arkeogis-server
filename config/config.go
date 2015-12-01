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

package arkeogis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	//"path/filepath"
)

type Config struct {
	Server struct {
		Port         int    `json:"port"`
		CookieSecret string `json:"cookie_secret"`
	} `json:"server"`
	Database struct {
		DbName   string `json:"dbname"`
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port,omitempty"`
		SslMode  string `json:"sslmode,omitempty"`
	} `json:"database"`
}

var Main Config        // The main configuration (server port, database credentials, etc.)
var DistPath string    // Path where the server binary is
var WebPath string     // Path where the web root is
var DevDistPath string // Path to /src/server
var DevWebPath string  // Path to /src/web
var DevMode bool       // true if we are running in dev mode
var CurDistPath string // = DevMode ? DevDistPath :  DistPath
var CurWebPath string  // = DevMode ? CurDistPath : DistPath

func init() {
	//var err error

	// Root path of the server
	/*
		DistPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
	*/
	DistPath, _ = os.Getwd()

	// Web root
	WebPath = path.Join(DistPath, "public")

	DevDistPath = path.Join(DistPath, "../src/server")
	DevWebPath = path.Join(DistPath, "../src/web")

	fmt.Println("launch args: ", os.Args)

	if len(os.Args) > 1 && os.Args[1] == "dev" {
		DevMode = true
		CurDistPath = DevDistPath
		CurWebPath = DevWebPath
	} else {
		DevMode = false
		CurDistPath = DistPath
		CurWebPath = WebPath
	}

	fmt.Println("DistPath : " + DistPath)
	fmt.Println("Dev Mode : ", DevMode)
	ReadMain("config.json")
}

func ReadMain(f string) {
	filename := path.Join(DistPath, f)
	fmt.Println("Reading config file " + filename)
	c, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(c, &Main); err != nil {
		log.Fatal("error:", err)
	}
}

func Write(sc interface{}, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	e := json.NewEncoder(file)
	err = e.Encode(sc)
	if err != nil {
		log.Fatal(err)
	}
}
