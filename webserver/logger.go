package webserver

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/negroni"
	config "github.com/croll/arkeogis-server/config"
)

// ALogger interface
type ALogger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	// ALogger implements just enough log.Logger interface to be compatible with other implementations
	ALogger
}

// NewLogger returns a new Logger instance
func NewLogger() *Logger {
	f, err := os.OpenFile(config.DistPath+"/logs/access.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("Error opening access log file: %v", err)
	}
	//defer f.Close()

	return &Logger{log.New(f, "", 0)}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	l.Printf(`[%s / %s] %s %s %s %d %d "%s" "%s"`, start.Format(time.RFC3339), time.Since(start), r.RemoteAddr, r.Method, r.URL.Path, res.Status(), res.Size(), r.Referer(), r.UserAgent())
}
