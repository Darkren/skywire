package vpn

import (
	"bytes"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

// ClientMobile is a VPN client used for mobile devices.
type ClientMobile struct {
	cfg       ClientConfig
	log       logrus.FieldLogger
	conn      net.Conn
	closeC    chan struct{}
	closeOnce sync.Once
}

// NewClientMobile create VPN client instance to be used on mobile devices.
func NewClientMobile(cfg ClientConfig, l logrus.FieldLogger, conn net.Conn) (*ClientMobile, error) {
	return &ClientMobile{
		cfg:    cfg,
		log:    l,
		conn:   conn,
		closeC: make(chan struct{}),
	}, nil
}

// GetConn returns VPN server connection.
func (c *ClientMobile) GetConn() net.Conn {
	return c.conn
}

// Close closes client.
func (c *ClientMobile) Close() {
	c.closeOnce.Do(func() {
		close(c.closeC)
	})
}

// ShakeHands performs client/server handshake.
func (c *ClientMobile) ShakeHands() (TUNIP, TUNGateway, serverTUNIP, serverTUNGateway net.IP, err error) {
	cHello := ClientHello{
		Passcode: c.cfg.Passcode,
	}

	return DoClientHandshake(c.log, c.GetConn(), cHello)
}

// TODO (darkrengarius): pack ip/gateway into a separate struct

func (c *ClientMobile) ShakeHandsKeepingCreds(cTUNIP, cTUNGateway, sTUNIP, sTUNGateway net.IP) (TUNIP, TUNGateway, serverTUNIP, serverTUNGateway net.IP, err error) {
	cHello := ClientHello{
		Passcode:         c.cfg.Passcode,
		ClientTUNIP:      &cTUNIP,
		ClientTUNGateway: &cTUNGateway,
		ServerTUNIP:      &sTUNIP,
		ServerTUNGateway: &sTUNGateway,
	}

	return DoClientHandshake(c.log, c.GetConn(), cHello)
}

func (c *ClientMobile) ioCopy(dst io.Writer, src io.Reader, buf []byte, isOutgouing bool) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	if buf == nil {
		size := 32 * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			if bytes.Contains(buf[:nr], []byte{195, 201, 201, 32}) {
				c.log.Infof("IS_OUTGOING: %v, READ PACKET: %v", isOutgouing, buf[:nr])
			}
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				if bytes.Contains(buf[:nr], []byte{195, 201, 201, 32}) {
					c.log.Infof("IS_OUTGOING: %v, WROTE PACKET: %v", isOutgouing, buf[:nw])
				}
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// Serve starts handling traffic.
func (c *ClientMobile) Serve(udpConnRW *UDPConnWriter) error {
	tunnelConn := c.GetConn()

	connToUDPDoneCh := make(chan error)
	udpToConnCh := make(chan error)
	// read all system traffic and pass it to the remote VPN server
	go func() {
		defer close(connToUDPDoneCh)

		//if _, err := io.Copy(udpConnRW, tunnelConn); err != nil {
		if _, err := c.ioCopy(udpConnRW, tunnelConn, nil, false); err != nil {
			c.log.WithError(err).Errorln("ERROR RESENDING TRAFFIC FROM VPN SERVER")
			//c.log.WithError(err).Errorln("Error resending traffic from VPN server to mobile app UDP conn")
			connToUDPDoneCh <- err
		} else {
			c.log.Errorln("NO ERROR RESENDING TRAFFIC FROM VPN SERVER")
		}
	}()
	go func() {
		defer close(udpToConnCh)

		//if _, err := io.Copy(tunnelConn, udpConnRW.conn); err != nil {
		if _, err := c.ioCopy(tunnelConn, udpConnRW.conn, nil, true); err != nil {
			c.log.WithError(err).Errorln("ERROR RESENDING TRAFFIC FROM MOBILE APP")
			//c.log.WithError(err).Errorln("Error resending traffic from mobile app UDP conn to VPN server")
			udpToConnCh <- err
		} else {
			c.log.Errorln("NO ERROR RESENDING TRAFFIC FROM MOBILE APP")
		}
	}()

	// only one side may fail here, so we wait till at least one fails
	select {
	case err := <-connToUDPDoneCh:
		return err
	case err := <-udpToConnCh:
		return err
	case <-c.closeC:
	}

	return nil
}
