package main

import (
	"flag"
	"fmt"
	"github.com/vsheffer/gofun/rdf"
	"github.com/vsheffer/gofun/util"
	"strings"
)

func main() {
	var uriStrings util.StringSlice

	rdfLoader := rdf.NewRdfLoader()

	parserNames := rdfLoader.World.GetParserNames()
	namesString := strings.Join(parserNames, ",")
	namesBlurb := fmt.Sprintf("Name of the parser.  Possible names are: (%s)", namesString)

	flag.Var(&uriStrings, "inputUri", "One or more URIs to the file to import.")
	rdfLoader.Args.IsNew = flag.String("new", "yes", "Is the model new (yes|no)?")
	rdfLoader.Args.ParserName = flag.String("parser", "guess", namesBlurb)
	rdfLoader.Args.DbHost = flag.String("dbhost", "localhost", "The host name of the database.")
	rdfLoader.Args.DbName = flag.String("dbname", "", "The database to import to.")
	rdfLoader.Args.DbUser = flag.String("dbuser", "", "The database user.")
	rdfLoader.Args.DbPassword = flag.String("dbpassword", "", "The password for the dbuser.")
	flag.Parse()

	rdfLoader.LoadAll(uriStrings.Get())
}
