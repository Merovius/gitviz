package main

import (
	"flag"
	"fmt"
	"github.com/Merovius/git2go"
	"log"
	"io"
	"os"
)

type Dumper interface {
	Dump(io.Writer)
}

type Blob struct {
	*git.Blob
	id *git.Oid
}

type Tree struct {
	*git.Tree
	id      *git.Oid
	entries []*git.TreeEntry
}

type Commit struct {
	*git.Commit
	id      *git.Oid
	tree    *git.Oid
	parents []*git.Oid
}

type Reference struct {
	name string
	id   *git.Oid
}

type SymbolicReference struct {
	name string
	*git.Reference
}

var shorten int

func (b *Blob) Dump(w io.Writer) {
	fmt.Fprintf(w, "\"%s\" [shape=box,style=filled,fillcolor=\"#ddddff\",color=\"#bbbbff\"]\n", b.id.String()[:shorten])
}

func (t *Tree) Dump(w io.Writer) {
	fmt.Fprintf(w, "\"%s\" [shape=oval,style=filled,fillcolor=\"#99ff99\"]\n", t.id.String()[:shorten])
	for _, e := range t.entries {
		fmt.Fprintf(w, "\"%s\" -> \"%s\" [label=\"%s\",fontcolor=\"#666666\"]\n", t.id.String()[:shorten], e.Id.String()[:shorten], e.Name)
	}
}

func (c *Commit) Dump(w io.Writer) {
	fmt.Fprintf(w, "\"%s\" [shape=hexagon,style=filled,fillcolor=\"#ffff99\"]\n",
		c.id.String()[:shorten])
	fmt.Fprintf(w, "\"%s\" -> \"%s\"\n", c.id.String()[:shorten], c.tree.String()[:shorten])
	for _, p := range c.parents {
		fmt.Fprintf(w, "\"%s\" -> \"%s\"\n", c.id.String()[:shorten], p.String()[:shorten])
	}
}

func (r *Reference) Dump(w io.Writer) {
	fmt.Fprintf(w, "\"%s\" [shape=box,style=filled,fillcolor=\"#9999ff\"]\n", r.name)
	fmt.Fprintf(w, "\"%s\" -> \"%s\"\n", r.name, r.id.String()[:shorten])
}

func (r *SymbolicReference) Dump(w io.Writer) {
	if r.Type() == git.SYMBOLIC {
		fmt.Fprintf(w, "\"%s\" [shape=box,style=filled,fillcolor=\"#ff9999\"]\n", r.name)
		fmt.Fprintf(w, "\"%s\" -> \"%s\"\n", r.name, r.SymbolicTarget())
	} else {
		fmt.Fprintf(w, "\"%s\" [shape=box,style=filled,fillcolor=\"#ff9999\"]\n", r.name)
		fmt.Fprintf(w, "\"%s\" -> \"%s\"\n", r.name, r.Target().String()[:shorten])
	}
}

func main() {
	noBrokenHead := flag.Bool("no-broken-head", false, "Hide a broken HEAD-ref")
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

	stuff := make(map[string]Dumper)
	var oids []*git.Oid

	for oid := range odb.ForEach() {
		obj, err := repo.Lookup(oid)
		if err != nil {
			log.Fatal("Lookup:", err)
		}
		switch obj := obj.(type) {
		default:
		case *git.Blob:
			bl := &Blob{obj, oid}
			stuff[oid.String()] = bl
			oids = append(oids, oid)
		case *git.Tree:
			tr := &Tree{obj, oid, nil}
			stuff[oid.String()] = tr
			for i := uint64(0); i < tr.EntryCount(); i++ {
				tr.entries = append(tr.entries, tr.EntryByIndex(i))
			}
			oids = append(oids, oid)
		case *git.Commit:
			co := &Commit{obj, oid, obj.TreeId(), nil}
			for i := uint(0); i < obj.ParentCount(); i++ {
				co.parents = append(co.parents, obj.ParentId(i))
			}
			stuff[oid.String()] = co
			oids = append(oids, oid)
		}
	}

	iter, err := repo.NewReferenceIterator()
	if err != nil {
		log.Fatal(err)
	}

	for refname := range iter.Iter() {
		ref, err := repo.LookupReference(refname)
		if err != nil {
			log.Fatal(err)
		}
		stuff[refname] = &Reference{refname, ref.Target()}
	}

	ref, err := repo.LookupReference("HEAD")
	if err != nil {
		log.Fatal(err)
	}

	if !*noBrokenHead ||
			(ref.Type() == git.SYMBOLIC && stuff[ref.SymbolicTarget()] != nil) ||
			(ref.Type() == git.OID && stuff[ref.Target().String()] != nil) {
		stuff["HEAD"] = &SymbolicReference{"HEAD", ref}
	}

	shorten, err = git.ShortenOids(oids, 4)
	if err != nil {
		shorten = 40
	}

	fmt.Println("digraph G {")
	for _, dp := range stuff {
		dp.Dump(os.Stdout)
	}
	fmt.Println("}")
}
