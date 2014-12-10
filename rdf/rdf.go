package rdf

import (
	"fmt"
	"github.com/vsheffer/golibrdf"
	"log"
)

const storageType = "postgresql"

type Args struct {
	DbName     *string
	DbHost     *string
	DbPassword *string
	DbUser     *string
	IsNew      *string
	ParserName *string
}

type RdfLoader struct {
	Args  *Args
	World *golibrdf.World
}

// Create a new RdfLoader instance.
func NewRdfLoader() *RdfLoader {
	rdfLoader := &RdfLoader{
		Args:  &Args{},
		World: golibrdf.NewWorld()}

	if err := rdfLoader.World.Open(); err != nil {
		log.Fatalf("World failed to open: %s", err.Error())
	}

	return rdfLoader
}

func (r *RdfLoader) LoadAll(uriStrings []string) {
	numChannels := len(uriStrings)
	c := make(chan int, numChannels)

	log.Printf("uriStrings = %+v, len = %d", uriStrings, numChannels)

	var storageOptions string
	for index, elem := range uriStrings {
		if index > 0 {

			// If we are loading more than 1 file in a single execution, then
			// only the first one can be "new".

			storageOptions = fmt.Sprintf("new='no',host='%s',database='%s',user='%s',password='%s'", *r.Args.DbHost, *r.Args.DbName, *r.Args.DbUser, *r.Args.DbPassword)
		} else {
			storageOptions = fmt.Sprintf("new='%s',host='%s',database='%s',user='%s',password='%s'", *r.Args.IsNew, *r.Args.DbHost, *r.Args.DbName, *r.Args.DbUser, *r.Args.DbPassword)
		}

		log.Printf("Loading %s\n", elem)
		go r.loadOne(elem, storageOptions, c)
	}

	for i := 0; i < numChannels; i++ {
		<-c
	}
}

func (r *RdfLoader) loadOne(uriStr string, storageOptions string, c chan int) {
	var err error

	log.Printf("Hello")
	log.Printf("parserName = %s", *r.Args.ParserName)
	log.Printf("uriString = %s", uriStr)
	log.Printf("storageOptions = %s", storageOptions)

	var uri *golibrdf.Uri
	if uri, err = golibrdf.NewUri(r.World, uriStr); err != nil {
		log.Printf("Failed to create URI: %s", err.Error())
		return
	}
	defer uri.Free()

	// construct a storage provider
	var storage *golibrdf.Storage
	if storage, err = golibrdf.NewStorage(r.World, storageType, "test", storageOptions); err != nil {
		log.Printf("Failed to create storage: %s", err.Error())
	}
	defer storage.Free()

	// construct a model
	var model *golibrdf.Model
	if model, err = golibrdf.NewModel(r.World, storage, ""); err != nil {
		log.Printf("Failed to construct model: %s", err.Error())
	}
	defer model.Free()

	parser, err := golibrdf.NewParser(r.World, *r.Args.ParserName, "")
	if err != nil {
		log.Printf("Error constructing a parser", err.Error())
	}
	defer parser.Free()

	if err = parser.ParseIntoModel(uri, nil, model); err != nil {
		log.Printf("Error parsing uri into model", err.Error())
	}

	c <- 1
}
