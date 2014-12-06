package main

import (
	"github.com/PhillP/golibrdf"
	"log"
	"flag"
	"fmt"
	"strings"
)

func main() {
	var err error

	// create a new world
	world := golibrdf.NewWorld()

	if err = world.Open(); err != nil {
		log.Fatalf("World failed to open: %s", err.Error())
	}
	defer world.Close()

	parserNames := world.PrintParserNames() 
	namesString := strings.Join(parserNames, ",")
	namesBlurb := fmt.Sprintf("Name of the parser.  Possible names are: (%s)", namesString)

	uriString := flag.String("inputUri", "", "The URI to the file to import.")
	isNew := flag.String("new", "yes", "Is the model new (yes|no)?")
	parserName := flag.String("parser", "guess", namesBlurb)
	dbHost := flag.String("dbhost", "localhost", "The host name of the database.")
	dbName := flag.String("dbname", "", "The database to import to.")
	dbUser := flag.String("dbuser", "", "The database user.")
	dbPassword := flag.String("dbpassword", "", "The password for the dbuser.")
	flag.Parse()

	storageType := "postgresql"
	var storageOptions string

	storageOptions = fmt.Sprintf("new='%s',host='%s',database='%s',user='%s',password='%s'", *isNew, *dbHost, *dbName, *dbUser, *dbPassword)

	log.Printf("parserName = %s", *parserName)
	log.Printf("uriString = %s", *uriString)
	log.Printf("storageOptions = %s", storageOptions)

	var uri *golibrdf.Uri
	if uri, err = golibrdf.NewUri(world, *uriString); err != nil {
		log.Fatalf("Failed to create URI: %s", err.Error())
	}
	defer uri.Free()

	// construct a storage provider
	var storage *golibrdf.Storage
	if storage, err = golibrdf.NewStorage(world, storageType, "test", storageOptions); err != nil {
		log.Fatalf("Failed to create storage: %s", err.Error())
	}
	defer storage.Free()

	// construct a model
	var model *golibrdf.Model
	if model, err = golibrdf.NewModel(world, storage, ""); err != nil {
		log.Fatalf("Failed to construct model: %s", err.Error())
	}
	defer model.Free()

	parser, err := golibrdf.NewParser(world, *parserName, "")
	if err != nil {
		log.Fatalf("Error constructing a parser", err.Error())
	}
	defer parser.Free()

	if err = parser.ParseIntoModel(uri, nil, model); err != nil {
		log.Fatalf("Error parsing uri into model", err.Error())
	}
}
