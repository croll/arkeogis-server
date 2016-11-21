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

package session

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/croll/arkeogis-server/model"
)

var sessionDuration time.Duration

// Session struct
type Session struct {
	LastAccess time.Time
	Values     map[string]interface{}
	mutex      sync.Mutex
}

// Get return a value of any type. ok is false if no value was found
func (session *Session) Get(key string) (value interface{}, ok bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	val, ok := session.Values[key]
	return val, ok
}

// GetDef return a value of any type. return the "def" parameter if no value was found
func (session *Session) GetDef(key string, def interface{}) interface{} {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	val, ok := session.Values[key]
	if !ok {
		return def
	}
	return val
}

// GetString return a value of string type. ok is false if no value was found
func (session *Session) GetString(key string) (value string, ok bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(string); castok {
			return v, castok
		}
		return "", false
	}
	return "", false
}

// GetStringDef return a value of string type. return the "def" parameter if no value was found
func (session *Session) GetStringDef(key string, def string) string {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(string); castok {
			return v
		}
		return def
	}
	return def
}

// GetInt return a value of integer type. ok is false if no value was found
func (session *Session) GetInt(key string) (value int, ok bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(int); castok {
			return v, castok
		}
		return 0, false
	}
	return 0, false
}

// GetIntDef return a value of integer type. return the "def" parameter if no value was found
func (session *Session) GetIntDef(key string, def int) int {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(int); castok {
			return v
		}
		return def
	}
	return def
}

// Set any value to a key string.
func (session *Session) Set(key string, value interface{}) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	session.Values[key] = value
	SaveSessions()
}

var sessions map[string]*Session
var GeneralMutex sync.Mutex

func init() {
	//sessionDuration = 7 * 24 * time.Hour // 7 days
	sessionDuration = 20 * time.Minute // 20 minutes
	sessions = make(map[string]*Session, 0)
	rand.Seed(time.Now().UnixNano())
	gob.Register(map[string]*Session{})
	gob.Register(model.User{})
	LoadSessions()
}

// GetSession return the Session instance of the given token. if no session was found for this token, a transient session will be given
func GetSession(token string) *Session {
	GeneralMutex.Lock()
	defer GeneralMutex.Unlock()
	cleanup()
	if s, ok := sessions[token]; ok {
		s.LastAccess = time.Now()
		return s
	} else {
		s = &Session{
			LastAccess: time.Now(),
			Values:     make(map[string]interface{}),
		}
		// removed bellow, because we don't save a transient session
		//sessions[token] = s
		return s
	}
}

// DestroySession destroy a session by token
func DestroySession(token string) {
	GeneralMutex.Lock()
	defer GeneralMutex.Unlock()
	if _, ok := sessions[token]; ok {
		delete(sessions, token)
	}
}

// NewSession return a new Session instance, with a new token.
func NewSession() (token string, s *Session) {
	GeneralMutex.Lock()
	defer GeneralMutex.Unlock()
	token = BuildRandomToken()
	s = &Session{
		LastAccess: time.Now(),
		Values:     make(map[string]interface{}),
	}
	// save it
	sessions[token] = s

	//fmt.Println(token, " => session created")

	return token, s
}

func cleanup() {
	expire := time.Now().Add(-sessionDuration)
	for token, session := range sessions {
		if session.LastAccess.Before(expire) {
			//fmt.Println(token, " => session DELETED")
			delete(sessions, token)
		}
	}
}

// BuildRandomToken build a random string used for token.
func BuildRandomToken() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	n := 42
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func SaveSessions() {
	GeneralMutex.Lock()
	defer GeneralMutex.Unlock()
	cleanup()

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(sessions)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
		return
	}
	sessions_str := base64.StdEncoding.EncodeToString(b.Bytes())

	err = ioutil.WriteFile("/tmp/arkeogis-sessions.gob.b64", ([]byte)(sessions_str), 0777)
	if err != nil {
		fmt.Println(`failed to save sessions`, err)
		return
	}
}

// go binary decoder
func LoadSessions() {
	GeneralMutex.Lock()
	defer GeneralMutex.Unlock()
	in, err := ioutil.ReadFile("/tmp/arkeogis-sessions.gob.b64")
	if err != nil {
		fmt.Println(`failed to open session file`, err)
		return
	}

	m := map[string]*Session{}
	by, err := base64.StdEncoding.DecodeString((string)(in))
	if err != nil {
		fmt.Println(`failed base64 Decode`, err)
		return
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(&m)
	if err != nil {
		fmt.Println(`failed gob Decode`, err)
		return
	}

	sessions = m

	fmt.Println("sessions loaded: ", m)
}
