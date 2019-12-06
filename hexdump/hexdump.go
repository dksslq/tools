package main

import "fmt"
import "log"
import "os"

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

	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal("[stat file] ", err)
	}
	filesize := fileinfo.Size()

	buf := make([]byte, 40960, 40960)
	fmt.Printf("%s%d%s\n", "unsigned char buf[", filesize, "] = {")
	fmt.Printf("\t")

	var count int64 = 1

	for {
		n, err := file.Read(buf)
		if err != nil {
			break
		}

		i := 0
		for i < n {
			fmt.Printf("0x%02X", buf[i])

			if count != filesize {
				/*逗号分隔*/
				fmt.Printf(", ")
				/*每十六字节并且不在文件尾时 换行*/
				if count%16 == 0 {
					fmt.Printf("\n\t")
				}
			}
			i++
			count++
		}
	}

	fmt.Println("\n};")
}
