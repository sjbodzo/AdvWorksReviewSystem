package main

import (
	"flag"
	"log"
	"os"

	"github.com/sjbodzo/review_system/db"
	"github.com/sjbodzo/review_system/queue"
	"github.com/sjbodzo/review_system/server"
)

var dbflags struct {
	port     int
	endpoint string
	database string
	user     string
	pw       string
}
var apiflags struct {
	version string
	port    int
}
var redisflags struct {
	endpoint string
	port     int
}

func init() {
	flag.IntVar(&apiflags.port, "apiPort", 8080, "Port server listens for apiflags on")
	flag.StringVar(&apiflags.version, "apiVersion", "v1", "Server-side apiflags version to run")
	flag.IntVar(&dbflags.port, "dbPort", 5432, "Port to connect to database with")
	flag.StringVar(&dbflags.endpoint, "dbEndpoint", "", "Database endpoint to connect to")
	flag.StringVar(&dbflags.database, "database", "", "Which database to connect to")
	flag.StringVar(&dbflags.pw, "dbPw", "", "Password to use when connecting to the database")
	flag.StringVar(&dbflags.user, "dbUser", "", "User to use when connecting to the database")
	flag.IntVar(&redisflags.port, "redisPort", 6379, "Port to connect to database with")
	flag.StringVar(&redisflags.endpoint, "redisEndpoint", "", "Database endpoint to connect to")
	flag.Parse()
}

func main() {
	f, err := os.OpenFile("/tmp/receiverd-logs", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Unable to open log file for writing")
	}
	defer f.Close()
	log.SetOutput(f)

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	wrapper, err := db.New(dbflags.endpoint, dbflags.port, dbflags.user,
		dbflags.pw, dbflags.database)
	if err != nil {
		return err
	}

	pool := queue.NewWorkerPool(redisflags.endpoint, redisflags.port)
	srv, err := server.New(apiflags.port, apiflags.version, wrapper, pool)
	if err != nil {
		return err
	}
	log.Println("Server live at", srv.Addr)
	defer srv.Close()
	return srv.ListenAndServe()
}
