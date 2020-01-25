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
	fmt.Fprintln(os.Stderr, "xorcat [v1.78]")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "\txorcat [-k SECRET] [-l] [-np LPORT] [-s LADDR] [<RADDR> [RPORT]]")
	die(n)
}

func die(excode int, closers ...io.Closer) {
	dieLW(excode, "", closers...)
}

var shit = false

func dieLW(excode int, lastWords string, closers ...io.Closer) {
	shit = true
	for _, closer := range closers {
		// err := closer.Close()
		// if err != nil {
		// 	println("close a closer:", err.Error())
		// } else {
		// 	println("close a closer:", err)
		// }
		closer.Close()
	}
	if lastWords != "" {
		fmt.Fprintln(os.Stderr, lastWords)
	}
	// println("exit with:", excode)
	// panic(excode)
	os.Exit(excode)
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
				dieLW(1, fmt.Sprintf("invalid option -- '%s'", arg[1:]))
			}
			switch "" {
			// headless args
			case argRIP:
				argRIP = arg
			case argRport:
				argRport = arg
			default:
				dieLW(1, fmt.Sprintf("invalid arg %s", arg))
			}
		}
	}

	if len(argKey) == 0 {
		dieLW(1, "invalid key")
	}
}

func main() {
	laddr, err := net.ResolveTCPAddr("tcp", argS+":"+argNp)
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()))
	}

	if argL {
		l(laddr, argRIP, argRport)
	} else {
		if argRIP == "" {
			dieLW(1, "no dest to connect to")
		} else if argRport == "" {
			dieLW(1, "no port to connect to")
		}
		c(laddr, argRIP, argRport)
	}
}

func l(laddr *net.TCPAddr, rip, rport string) {
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()))
	}
	defer listener.Close()

	conn, err := listener.AcceptTCP()
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()))
	}
	defer conn.Close()

	// 1 conn limit
	go func() {
		for {
			conn, err := listener.AcceptTCP()

			if shit {
				// oh shit
				return
			}

			if err != nil {
				dieLW(1, fmt.Sprintf("%s", err.Error()))
			}

			conn.Close()
		}
	}()

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	if rip != "" {
		if remoteAddr.IP.String() != rip {
			dieLW(1, fmt.Sprintf("invalid connection to [%s] from [%s] %d", localAddr.IP.String(), remoteAddr.IP.String(), remoteAddr.Port), conn, listener)
		} else if rport != "" {
			irport, err := strconv.Atoi(rport)
			if err != nil {
				dieLW(1, fmt.Sprintf("%s", err.Error()), conn, listener)
			}

			if remoteAddr.Port != irport {
				dieLW(1, fmt.Sprintf("invalid connection to [%s] from [%s] %d", localAddr.IP.String(), remoteAddr.IP.String(), remoteAddr.Port), conn, listener)
			}
		}
	}

	Transport(conn)
}

func c(laddr *net.TCPAddr, rip, rport string) {
	raddr, err := net.ResolveTCPAddr("tcp", rip+":"+rport)
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()))
	}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()))
	}

	Transport(conn)
}

func Transport(conn *net.TCPConn) {
	streamxor_recv, err := xor.New([]byte(argKey))
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()), conn)
	}
	streamxor_send, err := xor.New([]byte(argKey))
	if err != nil {
		dieLW(1, fmt.Sprintf("%s", err.Error()), conn)
	}

	recvBuf := make([]byte, 65536, 65536)
	sendBuf := make([]byte, 65536, 65536)

	// net recv
	go func() {
		for {
			n, cerr := conn.Read(recvBuf)

			if shit {
				// oh shit
				return
			}

			if cerr != nil {
				connErrorHandler(cerr, conn)
			}

			streamxor_recv.Write(recvBuf[:n])
			n, _ = streamxor_recv.Read(recvBuf)

			_, err := os.Stdout.Write(recvBuf[:n])
			if err != nil {
				dieLW(1, fmt.Sprintf("%s", err.Error()), conn)
			}
		}
	}()

	// net send
	for {
		// note - linux:    terminal input buffer 4096 bytes. limits.h: #define PIPE_BUF        4096
		n, ferr := os.Stdin.Read(sendBuf)

		if ferr != nil {
			if ferr == io.EOF {
				// note - windows:    Conn.Close() will send unsent bytes from the underfull buffer, then close its socket
				die(0, conn)
			} else {
				dieLW(1, fmt.Sprintf("%s", ferr.Error()), conn)
			}
		}

		streamxor_send.Write(sendBuf[:n])
		n, _ = streamxor_send.Read(sendBuf)

		// note:    socket actually sends bytes only when the buffer is full
		_, err := conn.Write(sendBuf[:n])

		if err != nil {
			connErrorHandler(err, conn)
		}
	}
}

func connErrorHandler(err error, conn io.Closer) {
	if err == io.EOF { // handle connection close
		die(0, conn)
	} else if _, ok := err.(*net.OpError); ok { // handle operror like "connection reset by peer"
		die(1, conn)
	} else { // other errors
		dieLW(1, fmt.Sprintf("%s", err.Error()), conn)
	}
}
