package main

import "flag"
import "io"
import "log"
import "os"

import "github.com/dksslq/xor"

var argKey string
var argIfile string
var argOfile string

func init() {
	flag.StringVar(&argKey, "k", "0123456789", "Secret string.")
	flag.StringVar(&argIfile, "i", "", "Input file.")
	flag.StringVar(&argOfile, "o", "", "Output file.")
	flag.Parse()
}

func main() {
	var ifile *os.File
	var ofile *os.File
	var err error

	if argIfile == "" {
		ifile = os.Stdin
	} else {
		ifile, err = os.OpenFile(argIfile, os.O_RDONLY, 4)
		if err != nil {
			log.Fatal("[open file] ", err)
		}
		defer ifile.Close()
	}

	if argOfile == "" {
		ofile = os.Stdout
	} else {
		// over write
		ofile, err = os.OpenFile(argOfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 2)
		if err != nil {
			log.Fatal("[open file] ", err)
		}
		defer ofile.Close()
	}

	buf := make([]byte, 65536, 65536)

	key := []byte(argKey)
	streamxor := xor.New(key)

	var n int

	for {
		n, err = ifile.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("[read file] ", err)
			}
		}

		streamxor.Write(buf[:n])
		n, _ = streamxor.Read(buf)

		ofile.Write(buf[:n])
	}
}
