package main

import (
	"github.com/Merovius/git2go"
	"flag"
	"fmt"
	"log"
)

type Blob struct{
	*git.Blob
	id *git.Oid
}

type Tree struct {
	*git.Tree
	id *git.Oid
	entries []*git.TreeEntry
}

type Commit struct {
	*git.Commit
	id *git.Oid
	tree *git.Oid
}

var shorten int

func (b *Blob) String() string {
	return fmt.Sprintf("\"%s\" [shape=plaintext]", b.id.String()[:shorten])
}

func (t *Tree) String() string {
	s := fmt.Sprintf("\"%s\" [shape=oval,style=filled,fillcolor=\"#99ff99\"]", t.id.String()[:shorten])
	for _, e := range(t.entries) {
		s += fmt.Sprintf("\n\"%s\" -> \"%s\" [label=\"%s\"]", t.id.String()[:shorten], e.Id.String()[:shorten], e.Name)
	}
	return s
}

func (c *Commit) String() string {
	s := fmt.Sprintf("\"%s\" [shape=hexagon,style=filled,fillcolor=\"#ffff99\"]\n", c.id.String()[:shorten])
	s += fmt.Sprintf("\"%s\" -> \"%s\"", c.id.String()[:shorten], c.tree.String()[:shorten])
	return s
}

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

	stuff := make(map[string]fmt.Stringer)
	var oids []*git.Oid

	for oid := range(odb.ForEach()) {
		obj, err := repo.Lookup(oid)
		if err != nil {
			log.Fatal(err)
		}
		switch obj := obj.(type) {
		default:
		case *git.Blob:
			bl := &Blob{ obj, oid }
			stuff[oid.String()] = bl
			oids = append(oids, oid)
		case *git.Tree:
			tr := &Tree{ obj, oid, nil }
			stuff[oid.String()] = tr
			for i := uint64(0); i < tr.EntryCount(); i++ {
				tr.entries = append(tr.entries, tr.EntryByIndex(i))
			}
			oids = append(oids, oid)
		case *git.Commit:
			co := &Commit{ obj, oid, obj.TreeId() }
			stuff[oid.String()] = co
			oids = append(oids, oid)
		}

	}
	shorten, err = git.ShortenOids(oids, 4)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("digraph G {")
	for _, str := range(stuff) {
		fmt.Println(str.String())
	}
	fmt.Println("}")
}
