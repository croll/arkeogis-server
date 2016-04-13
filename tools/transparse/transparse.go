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

package main

import (
	"fmt"
	//"os"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	config "github.com/croll/arkeogis-server/config"
	translate "github.com/croll/arkeogis-server/translate"
)

var re_namespace *regexp.Regexp
var re_translate *regexp.Regexp

func init() {
	re_translate = regexp.MustCompile("['\">](([A-Z0-9_]+\\.)|(\\.)){0,3}T_[A-Z0-9_]+")
	re_namespace = regexp.MustCompile("translate-namespace=['\"]([A-Z0-9_]*)['\"]")
}

func parseFile(p string, trans map[string]string) {
	in, err := ioutil.ReadFile(p)
	if err != nil {
		log.Println("Unable to open file : "+p+" => ", err)
		return
	}

	matched_ns := re_namespace.FindStringSubmatch((string)(in))
	if err != nil {
		//fmt.Println("MatchString err : ", err)
	} else {
		//fmt.Println("matches : ", matched_ns)
	}

	namespace := ""
	if len(matched_ns) >= 2 {
		namespace = matched_ns[1]
	}

	matched := re_translate.FindAllString((string)(in), 9999)
	//fmt.Println("matches : ", matched)
	for _, m := range matched {
		if strings.HasPrefix(m, "'") || strings.HasPrefix(m, "\"") || strings.HasPrefix(m, ">") {
			m = m[1:]
		}
		//fmt.Println("found : ", m)
		if strings.Index(m, ".") == 0 {
			if namespace == "" {
				log.Printf("Found '%s' in file '%s', without namespace", m, p)
			} else {
				m = namespace + m
			}
		}
		trans[m] = "#!#" + m
	}

	log.Println("- ", p, " : ", len(matched))
}

func parseDir(p string, ignores []string, trans map[string]string) {
	dirs, err := ioutil.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}
	for _, dir := range dirs {
		name := dir.Name()

		cont := false
		for _, ignore := range ignores {
			if ignore == name {
				cont = true
			}
		}

		if cont || strings.Index(name, ".") == 0 {
			continue
		}

		//fmt.Println(p+"/"+name)
		if dir.IsDir() {
			parseDir(p+"/"+name, ignores, trans)
		} else {
			parseFile(p+"/"+name, trans)
		}
	}
}

func rebuildLang(lang string, domain string) {
	ignores := []string{
		"bower_components",
		"img",
		"languages",
		"fonts",
		"bower.json",
		"scss",
	}

	var parsepath string
	if domain == "web" {
		parsepath = config.CurWebPath
	} else if domain == "server" {
		parsepath = config.DevDistPath + "/src/github.com/croll/arkeogis-server"
	}

	newTrans := make(map[string]string)
	parseDir(parsepath, ignores, newTrans)

	oldTrans, err := translate.ReadTranslation(lang, domain)
	if err != nil {
		log.Println("Problem while loading origin translation : ", err, " using empty one")
		oldTrans = make(map[string]string)
	}

	translate.MergeIn(newTrans, oldTrans)

	err = translate.WriteJSON(translate.PlateToTree(newTrans), lang, domain)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	rebuildLang("fr", "web")
	rebuildLang("en", "web")
	rebuildLang("de", "web")
	rebuildLang("es", "web")
	rebuildLang("eu", "web")
	rebuildLang("fr", "server")
	rebuildLang("en", "server")
	rebuildLang("de", "server")
	rebuildLang("es", "server")
	rebuildLang("eu", "server")
}
