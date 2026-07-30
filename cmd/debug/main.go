package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rachmanzz/words-xml/words"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: debug <docx>")
	}
	doc, err := words.ProcessDOCXFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(doc.WordsXML)
}
