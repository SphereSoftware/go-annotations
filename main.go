package main

import (
	"flag"
	ex "github.com/SphereSoftware/go-annotations/example"
	"github.com/SphereSoftware/go-annotations/registry"
	"log"
	"os"
)

func main() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var outName string
	if len(flag.Args()) > 0 {
		outName = flag.Arg(0)
	}
	pack := os.Getenv("GOPACKAGE")
	if pack != "" {
		log.Println("Generating annotations registry")
		processFile(wd, pack, outName)
	} else {
		ex.TestAnnotations()
	}
}

func processFile(path, pck, outName string) {
	log.Printf("processing package %s at folder: %s\n", pck, path)

	registry.GenerateRegistry(path, pck, outName)
	log.Printf("Registry is generated\n")
}
