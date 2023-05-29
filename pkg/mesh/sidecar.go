package mesh

import (
	iptables "Mini-K8s/pkg/iptable"
	"fmt"
	"io"
	"net"
	"syscall"
)

// 15001 outbound pod -> outside
// 15006 inbound outside -> pod

type Sidecar struct {
	PodIP         string
	Host          string
	forwardServer *net.TCPListener
}

func (sc *Sidecar) RunForwardServer(mode string) {

	var host string
	if mode == "outbound" {
		host = sc.Host + ":15001"
	} else {
		host = sc.Host + ":15006"
	}

	// set server
	listenAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return
	}
	server, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return
	}
	sc.forwardServer = server

	// set iptables
	ipt, err := iptables.New()
	if err != nil {
		return
	}
	if mode == "outbound" {
		exist, err := ipt.Exists("nat", "PREROUTING", "-p", "tcp", "-s", sc.PodIP, "-j", "DNAT", "--to-destination", host)
		if err != nil {
			return
		}
		if exist == false {
			fmt.Println("iptables -t nat -A PREROUTING -p tcp " + "-s" + " " + sc.PodIP + " -j DNAT --to-destination " + host)
			err = ipt.AppendNAT("PREROUTING", "-p", "tcp", "-s", sc.PodIP, "-j", "DNAT", "--to-destination", host)
			if err != nil {
				return
			}
		} else {
			fmt.Println("exist")
		}
	} else {
		exist, err := ipt.Exists("nat", "OUTPUT", "-p", "tcp", "-d", sc.PodIP, "-j", "DNAT", "--to-destination", host)
		if err != nil {
			return
		}
		if exist == false {
			fmt.Println("iptables -t nat -A OUTPUT -p tcp " + "-d" + " " + sc.PodIP + " -j DNAT --to-destination " + host)
			err = ipt.AppendNAT("OUTPUT", "-p", "tcp", "-d", sc.PodIP, "-j", "DNAT", "--to-destination", host)
			if err != nil {
				return
			}
		} else {
			fmt.Println("exist")
		}
	}

	// listen
	for {
		fmt.Println("(Sidecar) server listening...")
		conn, err := sc.forwardServer.AcceptTCP()
		if err != nil {
			continue
		}
		go sc.handleConnection(conn)
	}
}

func (sc *Sidecar) handleConnection(conn *net.TCPConn) {
	fmt.Printf("(Sidecar) Connection from %v,\t", conn.RemoteAddr().String())

	ipv4, port, conn, err := getOriginalDst(conn)
	if err != nil {
		return
	}

	fmt.Printf("to %v:%v\n", ipv4, port)

	dstAddr, err := net.ResolveIPAddr("ip", ipv4)
	if err != nil {
		return
	}
	dstTCPAddr := &net.TCPAddr{IP: dstAddr.IP, Port: int(port)}
	dstConn, err := net.DialTCP("tcp", nil, dstTCPAddr)
	if err != nil {
		fmt.Println("[ERROR] (Sidecar) Failed to connect to the target.")
		return
	}

	go copyPackages(conn, dstConn)
	go copyPackages(dstConn, conn)
}

func getOriginalDst(clientConn *net.TCPConn) (ipv4 string, port uint16, newTCPConn *net.TCPConn, err error) {

	remoteAddr := clientConn.RemoteAddr()
	if remoteAddr == nil {
		return
	}

	newTCPConn = nil

	clientConnFile, err := clientConn.File()
	if err != nil {
		return
	}

	clientConn.Close()

	addr, err := syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IP, 80)
	if err != nil {
		return
	}
	newConn, err := net.FileConn(clientConnFile)
	if err != nil {
		return
	}

	if _, ok := newConn.(*net.TCPConn); ok {
		newTCPConn = newConn.(*net.TCPConn)
		clientConnFile.Close()
	} else {
		return
	}

	ipv4 = itod(uint(addr.Multiaddr[4])) + "." +
		itod(uint(addr.Multiaddr[5])) + "." +
		itod(uint(addr.Multiaddr[6])) + "." +
		itod(uint(addr.Multiaddr[7]))
	port = uint16(addr.Multiaddr[2])<<8 + uint16(addr.Multiaddr[3])

	return
}

func itod(i uint) string {
	if i == 0 {
		return "0"
	}

	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}
	return string(b[bp:])
}

func copyPackages(dst io.ReadWriteCloser, src io.ReadWriteCloser) {
	if dst == nil || src == nil {
		return
	}

	defer dst.Close()
	defer src.Close()

	_, err := io.Copy(dst, src)
	if err != nil {
		return
	}
}
