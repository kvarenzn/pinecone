package main

import (
	"fmt"
	"os"

	"github.com/kvarenzn/pinecone/parser"
	"github.com/kvarenzn/pinecone/tokenizer"
)

func main() {
	code, err := os.ReadFile("../1.pine")
	if err != nil {
		panic(err)
	}
	tokens := tokenizer.Tokenize(string(code))

	fmt.Println(tokens)

	stmts, errs := parser.Parse(tokens)

	for _, stmt := range stmts {
		fmt.Printf("%#v\n", stmt)
	}

	for _, err := range errs {
		fmt.Printf("%d:%d:%s\n", err.Row, err.Col, err.Msg)
	}
}
