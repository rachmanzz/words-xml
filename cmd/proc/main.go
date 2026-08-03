package main

import (
	"fmt"
	"log"
	"os"
	"github.com/rachmanzz/words-xml/words"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: proc <docx>")
	}
	doc, err := words.ProcessDOCXFileMode(os.Args[1], "lossless")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(doc.WordsXML)
}
