package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/ramonsaboya/myrpc/commons"
)

type SRH struct {
	tcpListener *net.Listener
	udpConn     *net.UDPConn

	protocol commons.Protocol

	replyToTCP *net.Conn
	replyToUDP *net.UDPAddr
}

var errUnknownProtocolError error = errors.New("Unknown protocol")
var errReplyToNotSet error = errors.New("reply-to field is not set")

func NewSRH(protocol commons.Protocol, host string, port int) (*SRH, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	if protocol == commons.TCP {
		listener, err := net.Listen(string(protocol), address)
		if err != nil {
			return nil, err
		}
		return &SRH{
			tcpListener: &listener,
			protocol:    protocol,
		}, nil
	} else if protocol == commons.UDP {
		udpAddr, err := net.ResolveUDPAddr(string(protocol), address)
		if err != nil {
			return nil, err
		}
		udpConn, err := net.ListenUDP(string(protocol), udpAddr)
		if err != nil {
			return nil, err
		}
		return &SRH{
			udpConn:  udpConn,
			protocol: protocol,
		}, nil
	}
	return nil, errUnknownProtocolError
}

func (srh *SRH) Receive() ([]byte, error) {
	data := make([]byte, 1024)
	if srh.protocol == commons.TCP {
		return srh.receiveTCP(data)
	} else if srh.protocol == commons.UDP {
		return srh.receiveUDP(data)
	}
	return nil, errUnknownProtocolError
}

func (srh *SRH) Send(data []byte) error {
	if srh.protocol == commons.TCP {
		return srh.sendTCP(data)
	} else if srh.protocol == commons.UDP {
		return srh.sendUDP(data)
	}
	return errUnknownProtocolError
}

func (srh *SRH) receiveTCP(data []byte) ([]byte, error) {
	listener := (*srh.tcpListener).(*net.TCPListener)
	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	n, err := conn.Read(data)
	if err != nil {
		return nil, err
	}
	if srh.replyToTCP == nil || !sameClient(srh.replyToTCP, &conn) {
		srh.replyToTCP = &conn
	}
	return data[:n], nil
}

func (srh *SRH) receiveUDP(data []byte) ([]byte, error) {
	conn := *srh.udpConn
	n, addr, err := conn.ReadFromUDP(data)
	if err != nil {
		return nil, err
	}
	srh.replyToUDP = addr
	return data[:n], nil
}

func (srh *SRH) sendTCP(data []byte) error {
	if srh.replyToTCP == nil {
		return errReplyToNotSet
	}
	_, err := (*srh.replyToTCP).Write(data)
	(*srh.replyToTCP).Close()
	if err != nil {
		return err
	}
	return nil
}

func (srh *SRH) sendUDP(data []byte) error {
	if srh.replyToUDP == nil {
		return errReplyToNotSet
	}
	_, err := (*srh.udpConn).WriteToUDP(data, srh.replyToUDP)
	if err != nil {
		return err
	}
	return nil
}

func sameClient(aConn, bConn *net.Conn) bool {
	a := (*aConn).RemoteAddr()
	b := (*bConn).RemoteAddr()
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Network() != b.Network() {
		return false
	}
	return a.String() == b.String()
}
