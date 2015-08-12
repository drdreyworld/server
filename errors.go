package server

import (
	"log"
)

func FatalIfError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func PanicIfError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func LogIfError(err error) {
	if err != nil {
		log.Println(err)
	}
}
