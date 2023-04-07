package main

import (
	"log"

	"github.com/adystag/jobs-search/internal"
	"github.com/adystag/jobs-search/internal/http"
	"github.com/adystag/jobs-search/internal/provider"
)

func main() {
	module, err := internal.NewModule(
		provider.Configuration{},
		provider.DB{},
		provider.Service{},
	)
	if err != nil {
		log.Fatalln(err)
	}

	err = http.NewServer(module).Run()
	if err != nil {
		log.Fatalln(err)
	}
}
