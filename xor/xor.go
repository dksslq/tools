package main

import "flag"
import "io"
import "log"
import "os"
import "encoding/hex"

import "github.com/dksslq/xor"

var argMethS bool
var argMethC bool
var argKey string
var argHey string
var argIfile string
var argOfile string

func init() {
	flag.BoolVar(&argMethS, "s", false, "sha512 infinite xor key stream, conflict with -c")
	flag.BoolVar(&argMethC, "c", false, "raw cycle xor key stream, conflict with -s")
	flag.StringVar(&argKey, "k", "", "Secret string, conflict with -x (default \"0123456789\")")
	flag.StringVar(&argHey, "x", "", "hex encoded key string, conflict with -k")
	flag.StringVar(&argIfile, "i", "", "Input file")
	flag.StringVar(&argOfile, "o", "", "Output file")
	flag.Parse()

	if argMethS && argMethC {
		flag.Usage()
		os.Exit(2)
	}

	if !argMethS && !argMethC {
		argMethS = true
	}

	if argKey != "" && argHey != "" {
		flag.Usage()
		os.Exit(2)
	}

	if argKey == "" && argHey == "" {
		argKey = "0123456789"
	}
}

func main() {
	var ifile *os.File
	var ofile *os.File
	var err error
	var key []byte
	var streamxor *xor.StreamXOR

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

	if argKey != "" {
		key = []byte(argKey)
	} else if argHey != "" {
		key, err = hex.DecodeString(argHey)
		if err != nil {
			log.Fatal("[decode hex key] ", err)
		}
	}

	// log.Println("key:", string(key))

	switch {
	case argMethS:
		streamxor, err = xor.NewC(key, xor.DIARRHEA, 65536)
	case argMethC:
		streamxor, err = xor.NewC(key, xor.CYCLE, 65536)
	}
	if err != nil {
		log.Fatal("[xor.New] ", err)
	}

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
