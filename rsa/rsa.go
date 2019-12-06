package main

import "bytes"
import "crypto/rand"
import "crypto/rsa"
import "crypto/x509"
import "fmt"
import "log"
import "os"
import "runtime"
import "strconv"

func usage() {
	fmt.Println("Usage:")
	fmt.Println("\trsa <-der> <BITS> <PRIVKEY> <PUBKEY>")
	fmt.Println("\trsa [-d] <KEY> <F_IN> [F_OUT]")
	os.Exit(-1)
}

func main() {
	switch {
	// rsa KEY F_IN
	case len(os.Args) == 3:
		pubkey := readFile(os.Args[1])
		buf_i := readFile(os.Args[2])

		buf_o, err := RsaEncryptPerChunk(buf_i, pubkey)
		if err != nil {
			log.Fatal("[Encrypt] ", err)
		}

		_, err = os.Stdout.Write(buf_o)
		if err != nil {
			os.Exit(-6)
		}
	case len(os.Args) == 4:
		// rsa -d KEY F_IN
		if os.Args[1] == "-d" {
			privkey := readFile(os.Args[2])
			buf_i := readFile(os.Args[3])

			buf_o, err := RsaDecryptPerChunk(buf_i, privkey)
			if err != nil {
				log.Fatal("[Dncrypt] ", err)
			}

			_, err = os.Stdout.Write(buf_o)
			if err != nil {
				os.Exit(-6)
			}
		} else {
			// rsa KEY F_IN F_OUT
			pubkey := readFile(os.Args[1])
			RsaEncryptFilePerChunk(os.Args[2], os.Args[3], pubkey)
		}
	case len(os.Args) == 5:
		// rsa -der BITS PRIVKEY PUBKEY
		if os.Args[1] == "-der" {
			keylen, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatal("[ParseKeyLen] ", err)
			}

			privkey, err := rsa.GenerateKey(rand.Reader, keylen)
			if err != nil {
				log.Fatal("[GenKey] ", err)
			}

			writeFile(os.Args[3], x509.MarshalPKCS1PrivateKey(privkey))
			writeFile(os.Args[4], x509.MarshalPKCS1PublicKey(&privkey.PublicKey))
		} else if os.Args[1] == "-d" {
			// rsa -d KEY F_IN F_OUT
			privkey := readFile(os.Args[2])
			RsaDecryptFilePerChunk(os.Args[3], os.Args[4], privkey)
		} else {
			usage()
		}

	default:
		usage()

	}

}

func readFile(fname string) []byte {
	file, err := os.OpenFile(fname, os.O_RDONLY, 4)
	if err != nil {
		log.Fatal("[OpenFile] ", err)
	}
	defer file.Close()
	file_info, err := file.Stat()
	if err != nil {
		log.Fatal("[StatFile] ", err)
	}

	buf := make([]byte, file_info.Size())

	_, err = file.Read(buf)
	if err != nil {
		log.Fatal("[ReadFile] ", err)
	}

	return buf
}

func writeFile(fname string, buf []byte) {
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 2)
	if err != nil {
		log.Fatal("[OpenFile] ", err)
	}
	defer file.Close()

	_, err = file.Write(buf)
	if err != nil {
		log.Fatal("[WriteFile] ", err)
	}
}

func RsaEncryptFilePerChunk(srcf, dstf string, publicKeyBytes []byte) {
	fsrc, err := os.OpenFile(srcf, os.O_RDONLY, 4)
	if err != nil {
		log.Fatal("[OpenFile] ", err)
	}
	defer fsrc.Close()
	fsrc_info, err := fsrc.Stat()
	if err != nil {
		log.Fatal("[StatFile] ", err)
	}
	fdst, err := os.OpenFile(dstf, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 2)
	if err != nil {
		log.Fatal("[OpenFile] ", err)
	}
	defer fdst.Close()
	defer fdst.Sync()

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBytes)
	if err != nil {
		log.Fatal("[ParseKey] ", err)
	}
	// log.Println("[EncryptInfo] ", "KeySize:", publicKey.Size())

	chunkSize := publicKey.Size() - 11
	chunk := make([]byte, chunkSize)
	nOtimes := fsrc_info.Size() / int64(chunkSize)
	if fsrc_info.Size()%int64(chunkSize) != 0 {
		nOtimes += 1
		// log.Println("[EncryptWarning] ", "Not Divisible")
	}
	for range make([]interface{}, nOtimes) {
		ln, err := fsrc.Read(chunk)
		if err != nil {
			log.Fatal("[ReadFile] ", err)
		}

		chunk = chunk[:ln]
		e_chunk, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, chunk)
		if err != nil {
			log.Fatal("[Encrypt] ", err)
		}

		_, err = fdst.Write(e_chunk)
		if err != nil {
			log.Fatal("[WriteFile] ", err)
		}
	}
	runtime.GC()
}

func RsaDecryptFilePerChunk(srcf, dstf string, privateKeyBytes []byte) {
	fsrc, err := os.OpenFile(srcf, os.O_RDONLY, 4)
	if err != nil {
		log.Fatal("[OpenFile] ", err)
	}
	defer fsrc.Close()
	fsrc_info, err := fsrc.Stat()
	if err != nil {
		log.Fatal("[StatFile] ", err)
	}
	fdst, err := os.OpenFile(dstf, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 2)
	if err != nil {
		log.Fatal("[OpenFile] ", err)
	}
	defer fdst.Close()
	defer fdst.Sync()

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		log.Fatal("[ParseKey] ")
	}
	// log.Println("[DecryptInfo] ", "KeySize:", privateKey.Size())

	chunkSize := privateKey.Size()
	chunk := make([]byte, chunkSize)
	nOtimes := fsrc_info.Size() / int64(chunkSize)
	if fsrc_info.Size()%int64(chunkSize) != 0 {
		nOtimes += 1
		// log.Println("[DecryptInfo] ", "Not Divisible")
	}

	for range make([]interface{}, nOtimes) {
		ln, err := fsrc.Read(chunk)
		if err != nil {
			log.Fatal("[ReadFile] ", err)
		}

		chunk = chunk[:ln]
		d_chunk, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, chunk)
		if err != nil {
			log.Fatal("[Encrypt] ", err)
		}

		_, err = fdst.Write(d_chunk)
		if err != nil {
			log.Fatal("[WriteFile] ", err)
		}
	}
	runtime.GC()
}

func RsaEncryptPerChunk(src, publicKeyBytes []byte) (bytesEncrypted []byte, err error) {
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBytes)
	if err != nil {
		return
	}

	chunkSize := publicKey.Size() - 11
	chunks := split(src, chunkSize)
	buffer := bytes.Buffer{}

	for _, chunk := range chunks {
		e_chunk, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, chunk)
		if err != nil {
			return nil, err
		}

		buffer.Write(e_chunk)
	}
	bytesEncrypted = buffer.Bytes()
	runtime.GC()
	return
}

func RsaDecryptPerChunk(src, privateKeyBytes []byte) (bytesDecrypted []byte, err error) {
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return
	}

	chunkSize := privateKey.Size()
	chunks := split(src, chunkSize)
	buffer := bytes.Buffer{}

	for _, chunk := range chunks {
		d_chunk, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, chunk)
		if err != nil {
			return nil, err
		}

		buffer.Write(d_chunk)
	}
	bytesDecrypted = buffer.Bytes()
	runtime.GC()
	return
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	ln := len(buf) / lim
	if len(buf)%lim != 0 {
		ln += 1
	}
	chunks := make([][]byte, 0, ln)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}
