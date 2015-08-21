package main

import (
	"flag"

	"github.com/benschw/chinchilla/example/ex"
)

func main() {
	bind := flag.String("bind", ":8080", "address to bind to")
	flag.Parse()

	s := ex.NewServer(*bind)
	s.Start()

}
