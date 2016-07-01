/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2016 CROLL SAS
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

package geo

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	db "github.com/croll/arkeogis-server/db"
	// "github.com/lukeroth/gdal"
)

// Point stores coordinates of a location
type Point struct {
	X    float64
	Y    float64
	Z    float64
	EPSG int
}

// NewPoint returns a Point from ESPG and X/Y/Z coordinates
func NewPoint(epsg int, xyz ...float64) (point *Point, err error) {

	point = &Point{EPSG: epsg}
	if len(xyz) < 2 || len(xyz) > 3 {
		return nil, errors.New("Wrong number of params for function ToWGS84")
	}
	i := 0
	str := ""
	for _, p := range xyz {
		v := strconv.FormatFloat(p, 'E', -1, 64)
		str += v + " "
		switch i {
		case 0:
			point.X = p
		case 1:
			point.Y = p
		case 2:
			point.Z = p
		}
		i++
	}
	return
}

// ToWKT is a conveniance function which returns the WKT string of a Point
func (p *Point) ToEWKT() string {
	return "SRID=" + strconv.Itoa(p.EPSG) + ";POINT(" + strconv.FormatFloat(p.X, 'f', -1, 64) + " " + strconv.FormatFloat(p.Y, 'f', -1, 64) + " " + strconv.FormatFloat(p.Z, 'f', -1, 64) + ")"
}

// ToWKT_2d is a conveniance function which returns the WKT 2D string of a Point
func (p *Point) ToEWKT_2d() string {
	return "SRID=" + strconv.Itoa(p.EPSG) + ";POINT(" + strconv.FormatFloat(p.X, 'f', -1, 64) + " " + strconv.FormatFloat(p.Y, 'f', -1, 64) + ")"
}

// NewPointByGeonameID returns a Point from a geoname id
func NewPointByGeonameID(geonameID int) (*Point, error) {

	var coords = struct {
		X float64
		Y float64
	}{}

	fmt.Println(geonameID)

	err := db.DB.Get(&coords, "SELECT ST_X(geom_centroid::geometry) AS y, ST_Y(geom_centroid::geometry) AS x FROM city WHERE geonameid = $1", geonameID)

	if err == sql.ErrNoRows {
		return nil, err
	}

	point, err := NewPoint(4326, coords.X, coords.Y)

	if err != nil {
		return nil, err
	}

	return point, nil
}
