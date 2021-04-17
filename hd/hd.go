package main

import "fmt"
import "log"
import "os"
import "io"
import "flag"

var File, Prefix, Suffix string

func main() {
	flag.StringVar(&File, "f", "", "file to read, if not specified, read from stdin")
	flag.StringVar(&Prefix, "p", "", "prefix string for each hexadecimal")
	flag.StringVar(&Suffix, "s", "", "suffix string for each hexadecimal")
	flag.Parse()

	var err error
	var file *os.File

	if File == "" {
		file = os.Stdin
	} else {
		file, err = os.Open(File)
		defer file.Close()
	}
	if err != nil {
		log.Fatal("[open file] ", err)
	}

	buf := make([]byte, 40960, 40960)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {

			} else {
				log.Fatal("[read file] ", err)
			}
			break
		}

		i := 0
		for i < n {
			fmt.Printf(Prefix+"%02X"+Suffix, buf[i])
			i++
		}
	}
}
