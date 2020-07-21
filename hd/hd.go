package main

import "fmt"
import "log"
import "os"
import "io"

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n\thexdump <file>")
		return
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("[open file] ", err)
	}
	defer file.Close()

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
			fmt.Printf("%02X", buf[i])
			i++
		}
	}
}
