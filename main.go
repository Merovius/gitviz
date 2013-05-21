package main

import (
	"github.com/Merovius/git2go"
	"flag"
	"fmt"
	"log"
)

func main () {
	flag.Parse()
	var dir string
	var err error
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	} else {
		dir, err = git.Discover(".", false, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	repo, err := git.OpenRepository(dir)
	if err != nil {
		log.Fatal(err)
	}

	odb, err := repo.Odb()
	if err != nil {
		log.Fatal(err)
	}

	ch := odb.ForEach()

	for oid := range(ch) {
		fmt.Println(oid)
	}
}
