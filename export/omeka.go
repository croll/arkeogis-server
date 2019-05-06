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
	"bytes"
	//"database/sql"
	"encoding/csv"
	"encoding/json"
	"log"
	//"math"
	//"strconv"
	//"strings"
 	model "github.com/croll/arkeogis-server/model"
	//"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
)
 
func prettyPrint(i interface{}) string {
    s, _ := json.MarshalIndent(i, "", "\t")
	//return string(s)
	fmt.Println(string(s))
	return ""
}

 // SitesAsOmeka exports database and sites as as csv file for omeka
 func SitesAsCSV(siteIDs []int, isoCode string, includeDbName bool, tx *sqlx.Tx) (outp string, err error) {

	type MySiteRangeCharac struct {
		model.Site_range__charac
		Charac model.Charac               `json:"charac"`
		Charac_trs []model.Charac_tr		  `json:"charac_trs"`
	}

	type MySite_range struct {
		model.Site_range
		SiteRangeCharacs []MySiteRangeCharac              `json:"site_range__characs"`
	}

	type MySite struct {
		model.Site
		Site_ranges []MySite_range              `json:"site_ranges"`
		Site_trs []model.Site_tr                `json:"site_trs"`
	}
 
	type MyDatabase struct {
		model.Database
		Sites []MySite                          `json:"sites"`
		Database_trs []model.Database_tr        `json:"database_trs"`
	}


	 var buff bytes.Buffer
 
	 w := csv.NewWriter(&buff)
	 w.Comma = ';'
	 w.UseCRLF = true
 
	 err = w.Write([]string{
		"Type Données",
		"Nombre Caracterisations",
		"Dublin Core:Title",
		"Dublin Core:Creator",
		"Dublin Core:Subject",
		"Dublin Core:Date",
		"Dublin Core:Language",
		"Dublin Core:Description",
		"Dublin Core:Source",
		"Dublin core:Coverage",
		"Dublin Core:Rights",
		"URI Site",
		"ID-site",
		"source_id",
		"Titre Site",
		"Auteur Base",
		"Structure Editrice Base",
		"Sujet Base",
		"Date Realisation Base",
		"Langue Base",
		"Description site et base",
		"Source Base",
		"Nom Site",
		"Nom Commune",
		"Sujets",
		"Bibliographie Site",
		"Bibliographie Base",
		"Commentaires",
		"Licence Base",
		"Periode Debut",
		"Debut Periode",
		"Periode Fin",
		"Fin Periode",
		"Occupation",
		"Etat Connaissances",
		"Altitude",
		"Latitude",
		"Longitude",
		"geolocation:latitude",
		"geolocation:longitude",
		"geolocation:zoom_level",
		"geolocation:map_type",
		"geolocation:address",
	})

	if err != nil {
		log.Println("database::ExportCSV : ", err.Error())
	}
	w.Flush()

	q := `SELECT  db.id,  db.name, (  SELECT json_agg(items)  FROM (    SELECT s.*, (  SELECT json_agg(items)  FROM (    SELECT sr.*, (  SELECT json_agg(items)  FROM (    SELECT src.*, ( SELECT row_to_json(items)  FROM (    SELECT c.* FROM charac c WHERE c.id = src.charac_id  ) as items) AS charac     FROM site_range__charac src WHERE src.site_range_id = sr.id  ) items) AS site_range__characs     FROM site_range sr WHERE sr.site_id = s.id  ) items) AS site_ranges     FROM site s WHERE s.database_id = db.id  ) items) AS sites , (  SELECT json_agg(items)  FROM (    SELECT dtr.*    FROM database_tr dtr WHERE dtr.database_id = db.id  ) items) AS database_trs  FROM database db`
	q += ` WHERE db.id = 284`


	q = `SELECT row_to_json(items) AS database  FROM (    SELECT *, (  SELECT json_agg(items)  FROM (    SELECT s.*, (  SELECT json_agg(items)  FROM (    SELECT sr.*, (  SELECT json_agg(items)  FROM (    SELECT src.*, ( SELECT row_to_json(items)  FROM (    SELECT c.* FROM charac c WHERE c.id = src.charac_id  ) as items) AS charac     FROM site_range__charac src WHERE src.site_range_id = sr.id  ) items) AS site_range__characs     FROM site_range sr WHERE sr.site_id = s.id  ) items) AS site_ranges     FROM site s WHERE s.database_id = db.id  ) items) AS sites , (  SELECT json_agg(items)  FROM (    SELECT dtr.*    FROM database_tr dtr WHERE dtr.database_id = db.id  ) items) AS database_trs  FROM database db WHERE db.id=284  ) as items`
	q = `SELECT row_to_json(items) AS database  FROM (    SELECT *, (  SELECT json_agg(items)  FROM (    SELECT s.*, (  SELECT json_agg(items)  FROM (    SELECT sr.*, (  SELECT json_agg(items)  FROM (    SELECT src.*, ( SELECT row_to_json(items)  FROM (    SELECT c.* FROM charac c WHERE c.id = src.charac_id  ) as items) AS charac , (  SELECT json_agg(items)  FROM (    SELECT ctr.*    FROM charac_tr ctr WHERE ctr.charac_id = src.charac_id  ) items) AS charac_trs     FROM site_range__charac src WHERE src.site_range_id = sr.id  ) items) AS site_range__characs     FROM site_range sr WHERE sr.site_id = s.id  ) items) AS site_ranges     FROM site s WHERE s.database_id = db.id  ) items) AS sites , (  SELECT json_agg(items)  FROM (    SELECT dtr.*    FROM database_tr dtr WHERE dtr.database_id = db.id  ) items) AS database_trs  FROM database db WHERE db.id=284  ) as items`

	fmt.Println("query: "+q)
	rows2, err := tx.Query(q)
	fmt.Println("query done")
	if err != nil {
		fmt.Println("query done err")
		log.Println(err)
		rows2.Close()
		return
	}
	fmt.Println("query done 2")

	for rows2.Next() {

		var dbjson string

		if err = rows2.Scan(&dbjson); err != nil {
			log.Println(err)
			rows2.Close()
			return
		}

		//fmt.Println("sites json: "+dbjson)

		var database MyDatabase = MyDatabase{}
		//database.Sites = make([]MySite, 12)
		err := json.Unmarshal([]byte(dbjson), &database)
		fmt.Println(err)

		//fmt.Println("database : ")
		//fmt.Println(database)

		fmt.Println("site[0] : ")
		//fmt.Printf("%+v\n", database.Sites[0])
		prettyPrint(database.Sites[0])

		/*
		var line []string
 
		line = []string{
			// Type Données
			// Pour les sites toujours mettre la valeur : site
			"site",

			// Nombre Caracterisations
			// le nombre de caractérisation ayant le même SITE_SOURCE_ID
			// à usage affichage cluster thématique geolocation.
			"0",

			// Dublin Core:Title
			// champs : SITE_NAME, MAIN_CITY_NAME
			// type : concaténation 
			// séparateur entre champs  : ,
			name+","+city_name,

			// Dublin Core:Creator
			// champs : Prénom Nom
			// type : concaténation 
			// séparateur entre champs  : rien
			// Tous les auteurs de la base de données déclarés dans ArkeoGIS.
			// Il peut donc être multiple
			// séparateur entre les auteurs : #
			"",

			// Dublin Core:Subject
			// champs : CARAC_NAME, CARAC_LVL1, CARAC_LVL2, CARAC_LVL3, CARAC_LVL4
			// type : concaténation 
			// séparateur visuel entre champs  : ,
			// séparateur informatique à la fin du dernier champs non vide : #
			// La liste de toutes les caractérisations ayant le même SITE_SOURCE_ID
			// Il peut donc être multiple, elles sont listées dans l'ordre de l'importation de la base source ArkeoGIS
			// séparateur entre les caractérisations : #
			"",

			// Dublin Core:Date
			// champs : Date de réalisation
			// type : extrait
			// La date de réalisation de la base de données déclarés dans ArkeoGIS.
			// A partir de la date compléte dans ArkeoGIS, uniquement l'année est exportée.
			"",

			// Dublin Core:Language
			// champs : Langue de la base de données
			// type : individuel
			// Langue de la base de données déclarée dans ArkeoGIS lors de l'importation.
			"",

			// Dublin Core:Description
			// "champs : COMMENTS
			// type : extrait 
			// Uniquement celui de la ligne 1 du site si plusieurs lignes ayant le même SITE_SOURCE_ID
			// champs : Description
			// type : individuel
			// Description de la base de données dans ArkeoGIS.
			// Les deux informations sont présentées séparées par un : #
			"",

			// Dublin Core:Source
			// champs : Source de la base
			// type : individuel
			// Source de la base de donnée déclarée dans ArkeoGIS.
			"",

			// Dublin core:Coverage
			// champs : Site Name # Main City Name # STARTING_PERIOD # ENDING_PERIOD # Debut Periode # Fin Periode
			// type : concaténation 
			// séparateur entre champs  : #
			"",

			// Dublin Core:Rights
			// champs : Licence de la base
			// type : individuel
			// Licence de la base de données déclarée dans ArkeoGIS.
			"",

			// URI Site
			// champs : url fiche base de données
			// type : adresse url
			// De la fiche base de données dans ArkeoGIS.
			"",

			// ID-site
			// champs : ID unique du site dans ArkeoGIS
			// type : individuel
			"",

			// source_id
			// ID unique du site bdd origine : SITE_SOURCE_ID
			// type : individuel
			"",

			code,
			name,
			city_name,
			cgeonameid,
			"4326",
			slongitude,
			slatitude,
			saltitude,
			scentroid,
			knowledge_type,
			soccupation,
			startingPeriod,
			endingPeriod,
			scharac_name,
			scharac_lvl1,
			scharac_lvl2,
			scharac_lvl3,
			scharac_lvl4,
			sexceptional,
			bibliography,
			comment,
		}
 
		err := w.Write(line)
		w.Flush()
		if err != nil {
			log.Println("database::ExportCSV : ", err.Error())
		}
		*/
	}
 
	return buff.String(), nil
 }
 