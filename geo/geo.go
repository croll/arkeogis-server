/* ArkeoGIS - The Arkeolog Geographical Information Server Program
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
	db "github.com/croll/arkeogis-server/db"
	"github.com/lukeroth/gdal"
	"strconv"
)

// Point stores coordinates of a location
type Point struct {
	X    float64
	Y    float64
	Z    float64
	EPSG int
	CS   string
	Geom gdal.Geometry
}

// NewPoint returns a Point from ESPG and X/Y/Z coordinates
func NewPoint(epsg int, xyz ...float64) (*Point, error) {

	point := &Point{EPSG: epsg}
	var err error
	if len(xyz) < 2 || len(xyz) > 3 {
		return nil, errors.New("Wrong number of params for function ToWGS84")
	}
	str := ""
	spref := gdal.CreateSpatialReference("")
	if err = spref.FromEPSG(epsg); err != nil {
		return nil, err
	}
	point.CS, _ = spref.AttrValue("PROJCS", 0)
	i := 0
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
	point.Geom, err = gdal.CreateFromWKT("POINT("+str+")", spref)
	if err != nil {
		return nil, err
	}
	return point, nil
}

// ToWKT is a conveniance function which returns the WKT string of a Point
func (p *Point) ToWKT() (string, error) {
	if !p.IsValid() {
		return "", errors.New("Invalid point")
	}
	return p.Geom.ToWKT()
}

// ToWGS84 is used to convert a x,y and z (optional) coordinates from specified
// epsg to WGS84
// A new point is returned, leaving original point untouched
func (p *Point) ToWGS84() (*Point, error) {
	var err error
	p2 := p
	spref := gdal.CreateSpatialReference("")
	if err = spref.FromEPSG(4326); err != nil {
		return nil, err
	}
	p2.CS, _ = spref.AttrValue("PROJCS", 0)
	if err != nil {
		return nil, err
	}
	p2.Geom.TransformTo(spref)
	p2.X, p2.Y, p2.Z = p.Geom.Point(0)
	return p2, nil
}

// IsValid verify is point is valid or not
func (p *Point) IsValid() bool {
	if p.X == 0 || p.Y == 0 {
		return false
	}
	return true
}

// NewPoinByGeonameID returns a Point from a geoname id
func NewPointByGeonameID(geonameID int) (*Point, error) {

	var coords = struct {
		X float64
		Y float64
	}{}

	err := db.DB.Get(&coords, "SELECT ST_X(geom_centroid::geometry) AS x, ST_Y(geom_centroid::geometry) AS y FROM city WHERE geonameid = $1", geonameID)

	if err == sql.ErrNoRows {
		return NewPoint(4326, 0, 0)
	} else if err != nil {
		return nil, err
	}

	point, err := NewPoint(4326, coords.X, coords.Y)

	if err != nil {
		return nil, err
	}

	return point, nil
}
