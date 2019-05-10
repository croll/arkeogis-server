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
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"log"
	//"math"
	"strconv"
	"strings"
 	model "github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"
	"bufio"
)

type MyUser struct {
	model.User
	Companies []model.Company `json:"companies"`
}

type MySiteRangeCharac struct {
	model.Site_range__charac
	Charac model.Charac               `json:"charac"`
	Charac_trs []model.Charac_tr		  `json:"charac_trs"`
	Site_range__charac_trs []model.Site_range__charac_tr `json:"srcharactrs"`
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
	OwnerUser    model.User                 `json:"owneruser"`
	Authors      []MyUser               	`json:"authors"`
	Default_language_tr model.Lang_tr			`json:"default_language_tr"`
	License		model.License				`json:"license"`
}


func prettyPrint(i interface{}) string {
    s, _ := json.MarshalIndent(i, "", "\t")
	//return string(s)
	fmt.Println(string(s))
	return ""
}

func joinusers(objs []MyUser) string {
	var r=""
	for i, obj := range objs {
		if i > 0 {
			r += " # "
		}
		r += obj.Firstname + " " + obj.Lastname
	}
	return r
}

func joinuserscompanies(users []MyUser) string {
	var r=""
	for _, u := range users {
		for _, c := range u.Companies {
			if r != "" {
				r += " # "
			}
			r += c.Name
		}
	}
	return r
}

func getFirstLine(x string) string {
	scanner := bufio.NewScanner(strings.NewReader(x))
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}


func joinCharacs(cachedCharacs *map[int]string, characIds []int) string {
	var r=""
	for i, characId := range characIds {
		if i > 0 {
			r += " # "
		}
		fmt.Println("charac i: "+strconv.Itoa(i)+", id: "+strconv.Itoa(characId)+", s: "+(*cachedCharacs)[characId])
		r += (*cachedCharacs)[characId]
	}
	return r
}

func getCachedCharacs(isoCode string, separator string, tx *sqlx.Tx) map[int]string {
	characs := make(map[int]string)

	q := "WITH RECURSIVE nodes_cte(id, path) AS (SELECT ca.id, cat.name::TEXT AS path FROM charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = 0 UNION ALL SELECT ca.id, (p.path || '"+separator+"' || cat.name) FROM nodes_cte AS p, charac AS ca LEFT JOIN charac_tr cat ON ca.id = cat.charac_id LEFT JOIN lang ON cat.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND ca.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC"

	rows, err := tx.Query(q, isoCode)
	switch {
	case err == sql.ErrNoRows:
		rows.Close()
		return nil
	case err != nil:
		rows.Close()
		return nil
	}
	for rows.Next() {
		var id int
		var path string
		if err = rows.Scan(&id, &path); err != nil {
			return nil
		}
		characs[id] = path
	}

	return characs
}

type Chronocached struct {
	Id int
	ParentId int
	Name string
	Path string
	Start_date int
	End_date int
}

func getCachedChronology(chronoRootId int, isoCode string, separator string, tx *sqlx.Tx) map[int]Chronocached {
	chronology := make(map[int]Chronocached)

	q := "WITH RECURSIVE nodes_cte(parent_id, id, name, path, start_date, end_date) AS (SELECT c.parent_id, c.id, ctr.name, ctr.name::TEXT AS path, c.start_date, c.end_date FROM chronology AS c LEFT JOIN chronology_tr ctr ON c.id = ctr.chronology_id LEFT JOIN lang ON ctr.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND c.parent_id = $2 UNION ALL SELECT c.parent_id, c.id, ctr.name, (p.path || '"+separator+"' || ctr.name), c.start_date, c.end_date FROM nodes_cte AS p, chronology AS c LEFT JOIN chronology_tr ctr ON c.id = ctr.chronology_id LEFT JOIN lang ON ctr.lang_isocode = lang.isocode WHERE lang.isocode = $1 AND c.parent_id = p.id) SELECT * FROM nodes_cte AS n ORDER BY n.id ASC"

	rows, err := tx.Query(q, isoCode, chronoRootId)
	switch {
	case err == sql.ErrNoRows:
		rows.Close()
		return nil
	case err != nil:
		rows.Close()
		return nil
	}
	for rows.Next() {
		c := Chronocached{}
		if err = rows.Scan(&c.ParentId, &c.Id, &c.Name, &c.Path, &c.Start_date, &c.End_date); err != nil {
			return nil
		}
		chronology[c.Id] = c
	}

	return chronology
}

func getChronoName(chronocached *map[int]Chronocached, startDate int, endDate int) string {
	for _, c := range *chronocached {
		if c.Start_date == startDate && c.End_date == endDate {
			return c.Name
		}
	}
	return humanYear(startDate)+" : "+humanYear(endDate)
}

func humanYear(year int) string {
	if year == -2147483648 || year == 2147483647 {
		return "indeterminé"
	}
	if year <= 0 {
		return strconv.Itoa(year - 1)
	}
	return strconv.Itoa(year)
}


// SitesAsOmeka exports database and sites as as csv file for omeka
func SitesAsCSV(siteIDs []int, isoCode string, includeDbName bool, tx *sqlx.Tx) (outp string, err error) {

	var cachedCharacs = getCachedCharacs("fr", ",", tx)
	var cachedChronology = getCachedChronology(37, "fr", ",", tx)

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

	q := `
	SELECT row_to_json(items) AS database
	FROM (
	  SELECT *, 
	  (
		SELECT json_agg(items)
		FROM (
		  SELECT s.*, 
		(
		  SELECT json_agg(items)
		  FROM (
			SELECT sr.*, 
		  (
			SELECT json_agg(items)
			FROM (
			  SELECT src.*, 
			( SELECT row_to_json(items)
			  FROM (
				SELECT c.*
				FROM "charac" "c"
				WHERE c.id = src.charac_id
				ORDER BY c.id
			  ) as items
			) AS charac, 
			(
			  SELECT json_agg(items)
			  FROM (
				SELECT ctr.*
				FROM "charac_tr" "ctr"
				WHERE ctr.charac_id = src.charac_id
				ORDER BY ctr.charac_id
			  ) items
			) AS charac_trs, 
			(
			  SELECT json_agg(items)
			  FROM (
				SELECT srcharactr.*
				FROM "site_range__charac_tr" "srcharactr"
				WHERE srcharactr.site_range__charac_id = src.Id
				ORDER BY srcharactr.site_range__charac_id
			  ) items
			) AS srcharactrs
			  FROM "site_range__charac" "src"
			  WHERE src.site_range_id = sr.id
			  ORDER BY src.id
			) items
		  ) AS site_range__characs
			FROM "site_range" "sr"
			WHERE sr.site_id = s.id
			ORDER BY sr.id
		  ) items
		) AS site_ranges
		  FROM "site" "s"
		  WHERE s.database_id = db.id
		  ORDER BY s.id
		) items
	  ) AS sites, 
	  (
		SELECT json_agg(items)
		FROM (
		  SELECT dtr.*
		  FROM "database_tr" "dtr"
		  WHERE dtr.database_id = db.id
		) items
	  ) AS database_trs, 
	  ( SELECT row_to_json(items)
		FROM (
		  SELECT u.*
		  FROM "user" "u"
		  WHERE u.id = db.owner
		  ORDER BY u.id
		) as items
	  ) AS owneruser, 
	  (
		SELECT json_agg(items)
		FROM (
		  SELECT u.*, 
		(
		  SELECT json_agg(items)
		  FROM (
			SELECT comp.*
			FROM "company" "comp"
			LEFT JOIN "user__company"
			ON user__company.company_id = comp.id
			WHERE u.id = user__company.user_id
			ORDER BY comp.id
		  ) items
		) AS companies
		  FROM "user" "u"
		  LEFT JOIN "database__authors"
		  ON database__authors.user_id = u.id
		  WHERE db.id = database__authors.database_id
		  ORDER BY u.id
		) items
	  ) AS authors, 
	  (
		SELECT json_agg(items)
		FROM (
		  SELECT dtr.*
		  FROM "database_tr" "dtr"
		  WHERE dtr.database_id = db.id
		) items
	  ) AS database_trs, 
	  ( SELECT row_to_json(items)
		FROM (
		  SELECT l.*
		  FROM "lang_tr" "l"
		  WHERE l.lang_isocode = db.Default_language AND l.lang_isocode_tr='fr'
		) as items
	  ) AS default_language_tr, 
	  ( SELECT row_to_json(items)
		FROM (
		  SELECT license.*
		  FROM "license" "license"
		  WHERE license.id = db.license_id
		) as items
	  ) AS license
	  FROM "database" "db"
	  WHERE db.id=284
	  ORDER BY db.id
	) as items
	`

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


		for _, site := range database.Sites {

			// count caracs and build caracs array, and search chronos bounds
			var caracsCount = 0
			var caracsIds []int
			leftPeriodStart := 2147483647;
			leftPeriodEnd := 2147483647;
			rightPeriodStart := -2147483648;
			rightPeriodEnd := -2147483648;
			firstSiteRangeCharacComment := ""
			firstSiteRangeCharacBibliography := ""
			
			for i, sr := range site.Site_ranges {
				caracsCount += len(sr.SiteRangeCharacs)
				for j, charac := range sr.SiteRangeCharacs {
					caracsIds = append(caracsIds, charac.Charac_id)
					if i == 0 && j == 0 {
						firstSiteRangeCharacComment = translate.GetTranslatedFromTr(charac.Site_range__charac_trs, "fr", "Comment")
						firstSiteRangeCharacBibliography = translate.GetTranslatedFromTr(charac.Site_range__charac_trs, "fr", "Bibliography")
					}
				}
				if sr.Start_date1 <= leftPeriodStart {
					leftPeriodStart = sr.Start_date1
					leftPeriodEnd = sr.End_date1
				}
				if sr.End_date2 >= rightPeriodEnd {
					rightPeriodStart = sr.Start_date2
					rightPeriodEnd = sr.End_date2
				}
			}

			var line []string
 
			line = []string{
				// Type Données
				// Pour les sites toujours mettre la valeur : site
				"site",
	
				// Nombre Caracterisations
				// le nombre de caractérisation ayant le même SITE_SOURCE_ID
				// à usage affichage cluster thématique geolocation.
				strconv.Itoa(caracsCount),
	
				// Dublin Core:Title
				// champs : SITE_NAME, MAIN_CITY_NAME
				// type : concaténation 
				// séparateur entre champs  : ,
				site.Name+", "+site.City_name,
	
				// Dublin Core:Creator
				// champs : Prénom Nom
				// type : concaténation 
				// séparateur entre champs  : rien
				// Tous les auteurs de la base de données déclarés dans ArkeoGIS.
				// Il peut donc être multiple
				// séparateur entre les auteurs : #
				//database.OwnerUser.Firstname+" "+database.OwnerUser.Lastname,
				joinusers(database.Authors),
	
				// Dublin Core:Subject
				// champs : CARAC_NAME, CARAC_LVL1, CARAC_LVL2, CARAC_LVL3, CARAC_LVL4
				// type : concaténation 
				// séparateur visuel entre champs  : ,
				// séparateur informatique à la fin du dernier champs non vide : #
				// La liste de toutes les caractérisations ayant le même SITE_SOURCE_ID
				// Il peut donc être multiple, elles sont listées dans l'ordre de l'importation de la base source ArkeoGIS
				// séparateur entre les caractérisations : #
				joinCharacs(&cachedCharacs, caracsIds),
	
				// Dublin Core:Date
				// champs : Date de réalisation
				// type : extrait
				// La date de réalisation de la base de données déclarés dans ArkeoGIS.
				// A partir de la date compléte dans ArkeoGIS, uniquement l'année est exportée.
				database.Declared_creation_date.Format("2006"),
	
				// Dublin Core:Language
				// champs : Langue de la base de données
				// type : individuel
				// Langue de la base de données déclarée dans ArkeoGIS lors de l'importation.
				database.Default_language_tr.Name,
	
				// Dublin Core:Description
				// "champs : COMMENTS
				// type : extrait 
				// Uniquement celui de la ligne 1 du site si plusieurs lignes ayant le même SITE_SOURCE_ID
				// champs : Description
				// type : individuel
				// Description de la base de données dans ArkeoGIS.
				// Les deux informations sont présentées séparées par un : #
				translate.GetTranslatedFromTr(database.Database_trs, "fr", "Description"),
	
				// Dublin Core:Source
				// champs : Source de la base
				// type : individuel
				// Source de la base de donnée déclarée dans ArkeoGIS.
				translate.GetTranslatedFromTr(database.Database_trs, "fr", "Source_relation"),
	
				// Dublin core:Coverage
				// champs : Site Name # Main City Name # STARTING_PERIOD # ENDING_PERIOD # Debut Periode # Fin Periode
				// type : concaténation 
				// séparateur entre champs  : #
				site.Name+
					" # "+site.City_name+
					" # "+getChronoName(&cachedChronology, leftPeriodStart, leftPeriodEnd)+
					" # "+getChronoName(&cachedChronology, rightPeriodStart, rightPeriodEnd)+
					" # "+humanYear(leftPeriodStart)+" : "+humanYear(leftPeriodEnd)+
					" # "+humanYear(rightPeriodStart)+" : "+humanYear(rightPeriodEnd),
	
				// Dublin Core:Rights
				// champs : Licence de la base
				// type : individuel
				// Licence de la base de données déclarée dans ArkeoGIS.
				database.License.Name,
	
				// URI Site
				// champs : url fiche base de données
				// type : adresse url
				// De la fiche base de données dans ArkeoGIS.
				"https://app.arkeogis.org/#/database/"+strconv.Itoa(database.Id),
	
				// ID-site
				// champs : ID unique du site dans ArkeoGIS
				// type : individuel
				strconv.Itoa(site.Id)+"_s",
	
				// source_id
				// ID unique du site bdd origine : SITE_SOURCE_ID
				// type : individuel
				strconv.Itoa(site.Id),

				// Titre Site
				// champs : SITE_NAME, MAIN_CITY_NAME
				// type : concaténation 
				// séparateur entre champs  : ,
				site.Name+","+site.City_name,

				// Auteur Base
				// champs : Prénom Nom
				//
				// type : concaténation 
				// séparateur entre champs  : rien
				//
				// Tous les auteurs de la base de données déclarés dans ArkeoGIS.
				//
				// Il peut donc être multiple
				// séparateur entre les auteurs : #
				joinusers(database.Authors),

				// Structure Editrice Base
				// champs : ""Structure éditrice'
				//
				// type : individuel
				//
				// De la base de données du site déclarée dans ArkeoGIS.
				database.Editor,

				// Sujet Base
				// champs ""Sujet(s) / Mots-clés'
				//
				// type : individuel
				//
				// De la base de données du site déclarée dans ArkeoGIS.
				translate.GetTranslatedFromTr(database.Database_trs, "fr", "Subject"),


				// Date Realisation Base
				// champs : Date de réalisation
				//
				// type : extrait
				//
				// La date de réalisation de la base de données déclarés dans ArkeoGIS.
				//
				// A partir de la date compléte dans ArkeoGIS, uniquement l'année est exportée.
				database.Declared_creation_date.Format("2006"),

				// Langue Base
				// champs : Langue de la base de données
				//
				// type : individuel
				//
				// Langue de la base de données déclarée dans ArkeoGIS lors de l'importation.
				database.Default_language_tr.Name,

				// Description site et base
				//
				// champs : COMMENTS
				//
				// type : extrait 
				// Uniquement celui de la ligne 1 du site si plusieurs lignes ayant le même SITE_SOURCE_ID
				//
				// champs : Description
				//
				// type : individuel
				// Description de la base de données dans ArkeoGIS.
				//
				// Les deux informations sont présentées concaténées séparées par un : #
				firstSiteRangeCharacComment + " # " + translate.GetTranslatedFromTr(database.Database_trs, "fr", "Description"),

				// Source Base
				// champs : Source de la base
				//
				// type : individuel
				// Source de la base de données déclarée dans ArkeoGIS.
				"", //-TODO: not sure

				// Nom Site
				// champs : SITE_NAME
				// type : individuel
				site.Name,

				// Nom Commune
				// champs : MAIN_CITY_NAME
				// type : individuel
				site.City_name,

				// Sujets
				// champs : CARAC_NAME, CARAC_LVL1, CARAC_LVL2, CARAC_LVL3, CARAC_LVL4
				//
				// type : concaténation 
				// séparateur visuel entre champs  : ,
				// séparateur informatique à la fin du dernier champs non vide : #
				//
				// La liste de toutes les caractérisations ayant le même SITE_SOURCE_ID
				//
				// Il peut donc être multiple, elles sont listées dans l'ordre de l'importation de la base source ArkeoGIS
				// séparateur entre les caractérisations : #
				joinCharacs(&cachedCharacs, caracsIds),

				// Bibliographie Site
				// champs : BIBLIOGRAPHY
				//
				// type : extrait 
				// Celui de la ligne 1 du site si plusieurs lignes ayant le même SITE_SOURCE_ID
				firstSiteRangeCharacBibliography,

				// Bibliographie Base
				translate.GetTranslatedFromTr(database.Database_trs, "fr", "Bibliography"),

				// Commentaires
				// Licence Base
				// Periode Debut
				// Debut Periode
				// Periode Fin
				// Fin Periode
				// Occupation
				// Etat Connaissances
				// Altitude
				// Latitude
				// Longitude
				// geolocation:latitude
				// geolocation:longitude
				// geolocation:zoom_level
				// geolocation:map_type
				// geolocation:address
	
			}
	 
			err := w.Write(line)
			w.Flush()
			if err != nil {
				log.Println("database::ExportCSV : ", err.Error())
			}

		}

	}
 
	return buff.String(), nil
 }
 