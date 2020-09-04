package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tfaller/go-jsonera"
	"github.com/tfaller/go-jsonera/pkg/jsonp"
)

var jsonFile = flag.String("json", "", "the json file")
var eraFile = flag.String("era", "", "the era file")
var eraPrettyPrint = flag.Bool("p", false, "era-file pretty print")

func main() {
	flag.Parse()

	if *jsonFile == "" || *eraFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// first try to load the json
	var doc interface{}
	err := loadJSONFile(*jsonFile, &doc)
	if err != nil {
		log.Fatal(err)
	}

	// try to load era file, it might
	// not exist at the moment
	var eraDoc *jsonera.EraDocument
	err = loadJSONFile(*eraFile, &eraDoc)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	if eraDoc == nil {
		// no era file exists ... initialize new one
		eraDoc = jsonera.NewEraDocument(doc)
	}

	changes := eraDoc.UpdateDoc(doc)
	if len(changes) > 0 {
		fmt.Println("json-pointer,era,mode")
		for _, c := range changes {
			fmt.Printf("%q,%v,%v\n", jsonp.Format(c.Path), c.Era, c.Mode)
		}
	}

	// crate/update era document file
	eraOut, err := os.Create(*eraFile)
	if err != nil {
		log.Fatal(err)
	}
	defer eraOut.Close()

	encoder := json.NewEncoder(eraOut)
	if *eraPrettyPrint {
		encoder.SetIndent("", "    ")
	}
	err = encoder.Encode(eraDoc)
	if err != nil {
		log.Fatal(err)
	}
}

// loadJSONFile loads and parses a json file
func loadJSONFile(file string, target interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(target)
}
