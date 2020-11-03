package vpn

import "net"

// ClientHello is a message sent by client during the Client/Server handshake.
type ClientHello struct {
	UnavailablePrivateIPs []net.IP `json:"unavailable_private_ips"`
	Passcode              string   `json:"passcode"`
	ClientTUNIP           *net.IP  `json:"client_tun_ip,omitempty"`
	ClientTUNGateway      *net.IP  `json:"client_tun_gateway,omitempty"`
	ServerTUNIP           *net.IP  `json:"server_tun_ip,omitempty"`
	ServerTUNGateway      *net.IP  `json:"server_tun_gateway,omitempty"`
}
