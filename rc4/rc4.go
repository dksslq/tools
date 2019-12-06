package main

import "crypto/rc4"
import "flag"
import "fmt"
import "io"
import "log"
import "os"

var argKey string
var argIfile string
var argOfile string

func init() {
	flag.StringVar(&argKey, "k", "66abcdef", "key string. [-]")
	flag.StringVar(&argIfile, "i", "", "Input file. [R]")
	flag.StringVar(&argOfile, "o", "", "Output file. [-]")
	flag.Parse()
}

func main() {
	if argIfile == "" {
		flag.Usage()
		os.Exit(-1)
	}

	key := []byte(argKey)
	rc4cipher, err := rc4.NewCipher(key)
	if err != nil {
		log.Fatal("[parse key] ", err)
	}

	ifile, err := os.OpenFile(argIfile, os.O_RDONLY, 4)
	if err != nil {
		log.Fatal("[open file] ", err)
	}
	defer ifile.Close()

	ifileinfo, err := ifile.Stat()
	if err != nil {
		log.Fatal("[stat file] ", err)
	}
	ifilesize := ifileinfo.Size()

	var ofile *os.File = nil
	if argOfile != "" {
		// over write
		ofile, err = os.OpenFile(argOfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 2)
		if err != nil {
			log.Fatal("[open file] ", err)
		}
		defer ofile.Close()
	}

	buf := make([]byte, 65536, 65536)
	if argOfile == "" {
		fmt.Printf("%s%d%s\n", "unsigned char buf[", ifilesize, "] = {")
		fmt.Printf("\t\"")
	}

	var count int64 = 1
	var keylen int64 = int64(len(key))
	//log.Println("keylen:", keylen)

	for {
		n, err := ifile.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("[read file] ", err)
			}
		}

		i := 0
		for i < n {
			/*do xor*/
			if argOfile == "" {
				fmt.Printf("\\x%02X", buf[i]^key[(count-1)%keylen])
			} else {
				rc4cipher.XORKeyStream(buf, buf)
			}

			//fmt.Printf("%c", buf[i])
			/*每十六字节并且不在文件尾时 换行*/
			if argOfile == "" {
				if count%16 == 0 && count != ifilesize {
					fmt.Printf("\"\n\t\"")
				}
			}
			i++
			count++
		}
		ofile.Write(buf[:n])
	}
	if argOfile == "" {
		fmt.Printf("\"")
		fmt.Println("\n};")
	}
}
