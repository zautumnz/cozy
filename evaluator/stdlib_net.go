package evaluator

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/zacanger/cozy/object"
)

// Almost all of this is copy-pasted from the upstream, Prologic's version,
// which is licensed MIT.

func parseAddress(address string) (ip net.IP, port int, err error) {
	var (
		h string
		p string
		n int64
	)

	h, p, err = net.SplitHostPort(address)
	if err != nil {
		return
	}

	ip = net.ParseIP(h)
	if ip == nil {
		var addrs []string
		addrs, err = net.LookupHost(address)
		if err != nil {
			err = fmt.Errorf("error resolving host '%s'", address)
			return
		}

		if len(addrs) == 0 {
			err = fmt.Errorf("host not found '%s'", address)
			return
		}

		ip = net.ParseIP(addrs[0])
		if ip == nil {
			err = fmt.Errorf("invalid IP address '%s'", address)
			return
		}
	}

	n, err = strconv.ParseInt(p, 10, 16)
	if err != nil {
		return
	}
	port = int(n)
	return
}

func parseV4Address(address string) (addr [4]byte, port int, err error) {
	var ip net.IP
	ip, port, err = parseAddress(address)
	if err != nil {
		return
	}
	copy(addr[:], ip.To4()[:4])
	return
}

func parseV6Address(address string) (addr [16]byte, port int, err error) {
	var ip net.IP
	ip, port, err = parseAddress(address)
	if err != nil {
		return
	}
	copy(addr[:], ip.To16()[:16])
	return
}

// Socket is used like let f = socket("tcp4")
func Socket(args ...object.Object) object.Object {
	var (
		domain int
		typ    int
		proto  int
	)

	arg := args[0].(*object.String).Value

	switch strings.ToLower(arg) {
	case "unix":
		domain = syscall.AF_UNIX
		typ = syscall.SOCK_STREAM
		proto = 0
	case "tcp4":
		domain = syscall.AF_INET
		typ = syscall.SOCK_STREAM
		proto = syscall.IPPROTO_TCP
	case "tcp6":
		domain = syscall.AF_INET6
		typ = syscall.SOCK_STREAM
		proto = syscall.IPPROTO_TCP
	case "udp4":
		domain = syscall.AF_INET
		typ = syscall.SOCK_DGRAM
		proto = syscall.IPPROTO_UDP
	case "udp6":
		domain = syscall.AF_INET6
		typ = syscall.SOCK_DGRAM
		proto = syscall.IPPROTO_UDP
	default:
		return NewError("ValueError: invalid socket type '%s'", arg)
	}

	fd, err := syscall.Socket(domain, typ, proto)
	if err != nil {
		return NewError("SocketError: %s", err)
	}

	if domain == syscall.AF_INET || domain == syscall.AF_INET6 {
		if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
			return NewError("SocketError: cannot enable SO_REUSEADDR: %s", err)
		}
	}

	return &object.Integer{Value: int64(fd)}
}

// Listen used like listen(f, 1)
func Listen(args ...object.Object) object.Object {
	fd := int(args[0].(*object.Integer).Value)
	backlog := int(args[1].(*object.Integer).Value)

	if err := syscall.Listen(fd, backlog); err != nil {
		return NewError("SocketError: %s", err)
	}

	// fd
	return args[0]
}

// Connect is used by the client, example: connect(fd, "0.0.0.0:8080")
func Connect(args ...object.Object) object.Object {
	var sa syscall.Sockaddr

	fd := int(args[0].(*object.Integer).Value)
	address := args[1].(*object.String).Value

	sockaddr, err := syscall.Getsockname(fd)
	if err != nil {
		return NewError("ValueError: %s", err)
	}

	if _, ok := sockaddr.(*syscall.SockaddrInet4); ok {
		addr, port, err := parseV4Address(address)
		if err != nil {
			return NewError("ValueError: Invalid IPv4 address '%s': %s", address, err)
		}
		sa = &syscall.SockaddrInet4{Addr: addr, Port: port}
	} else if _, ok := sockaddr.(*syscall.SockaddrInet6); ok {
		addr, port, err := parseV6Address(address)
		if err != nil {
			return NewError("ValueError: Invalid IPv6 address '%s': %s", address, err)
		}
		sa = &syscall.SockaddrInet6{Addr: addr, Port: port}
	} else {
		return NewError("ValueError: Invalid socket type %T for bind '%s'", sockaddr, address)
	}

	if err = syscall.Connect(fd, sa); err != nil {
		return NewError("SocketError: %s", err)
	}

	// address
	return args[1]
}

// Close is for closing a connection
func Close(args ...object.Object) object.Object {
	fd := int(args[0].(*object.Integer).Value)

	err := syscall.Close(fd)
	if err != nil {
		return NewError("IOError: %s", err)
	}

	// file descriptor
	return args[0]
}

// Bind is used like bind(fd, "0.0.0.0:8080") server-side
func Bind(args ...object.Object) object.Object {
	var (
		err      error
		sockaddr syscall.Sockaddr
	)

	fd := int(args[0].(*object.Integer).Value)
	address := args[1].(*object.String).Value

	sockaddr, err = syscall.Getsockname(fd)
	if err != nil {
		return NewError("ValueError: %s", err)
	}

	if _, ok := sockaddr.(*syscall.SockaddrInet4); ok {
		addr, port, err := parseV4Address(address)
		if err != nil {
			return NewError("ValueError: Invalid IPv4 address '%s': %s", address, err)
		}
		sockaddr = &syscall.SockaddrInet4{Addr: addr, Port: port}
	} else if _, ok := sockaddr.(*syscall.SockaddrInet6); ok {
		addr, port, err := parseV6Address(address)
		if err != nil {
			return NewError("ValueError: Invalid IPv6 address '%s': %s", address, err)
		}
		sockaddr = &syscall.SockaddrInet6{Addr: addr, Port: port}
	} else {
		return NewError("ValueError: Invalid socket type %T for bind '%s'", sockaddr, address)
	}

	err = syscall.Bind(fd, sockaddr)
	if err != nil {
		return NewError("SocketError: %s", err)
	}

	// address
	return args[1]
}

// Accept takes requests
func Accept(args ...object.Object) object.Object {
	var (
		nfd int
		err error
	)

	fd := int(args[0].(*object.Integer).Value)

	nfd, _, err = syscall.Accept(fd)
	if err != nil {
		return NewError("SocketError: %s", err)
	}

	return &object.Integer{Value: int64(nfd)}
}

// Write to a socket
func Write(args ...object.Object) object.Object {
	fd := int(args[0].(*object.Integer).Value)
	data := []byte(args[1].(*object.String).Value)

	n, err := syscall.Write(fd, data)
	if err != nil {
		return NewError("IOError: %s", err)
	}

	return &object.Integer{Value: int64(n)}
}

// DefaultBufferSize is the default buffer size
const DefaultBufferSize = 4096

// Read from connection
func Read(args ...object.Object) object.Object {
	var (
		fd int
		n  = DefaultBufferSize
	)

	fd = int(args[0].(*object.Integer).Value)

	if len(args) == 2 {
		n = int(args[1].(*object.Integer).Value)
	}

	buf := make([]byte, n)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		return NewError("IOError: %s", err)
	}

	return &object.String{Value: string(buf[:n])}
}

func init() {
	RegisterBuiltin("net.socket",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Socket(args...))
		})

	RegisterBuiltin("net.listen",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Listen(args...))
		})

	RegisterBuiltin("net.connect",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Connect(args...))
		})

	RegisterBuiltin("net.close",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Close(args...))
		})

	RegisterBuiltin("net.bind",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Bind(args...))
		})

	RegisterBuiltin("net.accept",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Accept(args...))
		})

	RegisterBuiltin("net.write",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Write(args...))
		})

	RegisterBuiltin("net.read",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (Read(args...))
		})
}
