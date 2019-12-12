package main

import "fmt"
import "io"
import "net"
import "os"
import "strconv"

// import "syscall"

import "github.com/dksslq/xor"

var argKey string = "0123456789" // -k <key>

// for listen mode
var argL bool // -l

var argNp string // -np <lport>
var argS string  // -s <laddr>

// headless args
var argRIP string   // remote ip address
var argRport string // remote port

func usageExit(n int) {
	fmt.Println("xorcat [v1.44]")
	fmt.Println("Usage:")
	fmt.Println("\txorcat [-k SECRET] [-l] [-np LPORT] [-s LADDR] [<RADDR> [RPORT]]")
	os.Exit(n)
}

func init() {

	isVal := false
	var val *string
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if isVal {
			*val = arg
			isVal = false
			continue
		}

		switch arg {
		// headed args
		case "-h":
			fallthrough
		case "-help":
			fallthrough
		case "--help":
			usageExit(0)
		case "-k":
			isVal = true
			val = &argKey
		case "-l":
			argL = true
		case "-np":
			isVal = true
			val = &argNp
		case "-s":
			isVal = true
			val = &argS
		default:
			if arg[0] == '-' {
				fmt.Printf("invalid option -- '%s'\n", arg[1:])
				os.Exit(1)
			}
			switch "" {
			// headless args
			case argRIP:
				argRIP = arg
			case argRport:
				argRport = arg
			default:
				fmt.Println("invalid arg", arg)
				os.Exit(1)
			}
		}
	}
}

func main() {
	laddr, err := net.ResolveTCPAddr("tcp", argS+":"+argNp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if argL {
		l(laddr, argRIP, argRport)
	} else {
		if argRIP == "" {
			fmt.Println("no dest to connect to")
			os.Exit(1)
		} else if argRport == "" {
			fmt.Println("no port to connect to")
			os.Exit(1)
		}
		c(laddr, argRIP, argRport)
	}
}

func l(laddr *net.TCPAddr, rip, rport string) {
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer listener.Close()

	conn, err := listener.AcceptTCP()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	// 1 conn limit
	go func() {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	invalid_client := func() {
		fmt.Printf("invalid connection to [%s] from [%s] %d\n", localAddr.IP.String(), remoteAddr.IP.String(), remoteAddr.Port)
		conn.Close()
		os.Exit(1)
	}

	if rip != "" {
		if remoteAddr.IP.String() != rip {
			invalid_client()
		} else if rport != "" {
			irport, err := strconv.Atoi(rport)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if remoteAddr.Port != irport {
				invalid_client()
			}
		}
	}

	Transport(conn)
}

func c(laddr *net.TCPAddr, rip, rport string) {
	raddr, err := net.ResolveTCPAddr("tcp", rip+":"+rport)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Transport(conn)
}

func Transport(conn *net.TCPConn) {
	streamxor_recv := xor.New([]byte(argKey))
	streamxor_send := xor.New([]byte(argKey))

	recvBuf := make([]byte, 65536, 65536)
	sendBuf := make([]byte, 65536, 65536)

	// net recv
	go func() {
		for {
			n, cerr := conn.Read(recvBuf)

			if n > 0 {
				streamxor_recv.Write(recvBuf[:n])
				n, _ = streamxor_recv.Read(recvBuf)

				_, err := os.Stdout.Write(recvBuf[:n])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				connErrorHandler(cerr)
			}
		}
	}()

	// net send
	for {
		// note - linux:    terminal input buffer 4096 bytes. limits.h: #define PIPE_BUF        4096
		n, ferr := os.Stdin.Read(sendBuf)

		if n > 0 {
			streamxor_send.Write(sendBuf[:n])
			n, _ = streamxor_send.Read(sendBuf)

			_, err := conn.Write(sendBuf[:n])

			connErrorHandler(err)
		} else {
			if ferr != nil {
				// note - windows:    Conn.Close() will send unsent bytes from the underfull buffer, then close its socket
				// note:    socket actually sends bytes only when the buffer is full
				conn.Close()
				if ferr == io.EOF {
					os.Exit(0)
				} else {
					fmt.Println(ferr)
					os.Exit(1)
				}
			}
		}
	}
}

func connErrorHandler(err error) {
	if err != nil {
		if err == io.EOF { // handle connection close
			os.Exit(0)
		} else if _, ok := err.(*net.OpError); ok { // handle operror like "connection reset by peer"
			os.Exit(1)
		} else { // other errors
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}
