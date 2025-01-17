// Relay is a package that provides functionality for relaying network traffic.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var (
	// List of torrent trackers
	torrentTrackers = map[string]bool{
		"93.158.213.92":   true,
		"102.223.180.235": true,
		"23.134.88.6":     true,
		"185.243.218.213": true,
		"208.83.20.20":    true,
		"91.216.110.52":   true,
		"83.146.97.90":    true,
		"23.157.120.14":   true,
		"185.102.219.163": true,
		"163.172.29.130":  true,
		"156.234.201.18":  true,
		"209.141.59.16":   true,
		"34.94.213.23":    true,
		"192.3.165.191":   true,
		"130.61.55.93":    true,
		"109.201.134.183": true,
		"95.31.11.224":    true,
		"83.102.180.21":   true,
		"192.95.46.115":   true,
		"198.100.149.66":  true,
		"95.216.74.39":    true,
		"51.68.174.87":    true,
		"37.187.111.136":  true,
		"51.15.79.209":    true,
		"45.92.156.182":   true,
		"49.12.76.8":      true,
		"5.196.89.204":    true,
		"62.233.57.13":    true,
		"45.9.60.30":      true,
		"35.227.12.84":    true,
		"179.43.155.30":   true,
		"94.243.222.100":  true,
		"207.241.231.226": true,
		"207.241.226.111": true,
		"51.159.54.68":    true,
		"82.65.115.10":    true,
		"95.217.167.10":   true,
		"86.57.161.157":   true,
		"83.31.30.230":    true,
		"94.103.87.87":    true,
		"160.119.252.41":  true,
		"193.42.111.57":   true,
		"80.240.22.46":    true,
		"107.189.31.134":  true,
		"104.244.79.114":  true,
		"85.239.33.28":    true,
		"61.222.178.254":  true,
		"38.7.201.142":    true,
		"51.81.222.188":   true,
		"103.196.36.31":   true,
		"23.153.248.2":    true,
		"73.170.204.100":  true,
		"176.31.250.174":  true,
		"149.56.179.233":  true,
		"212.237.53.230":  true,
		"185.68.21.244":   true,
		"82.156.24.219":   true,
		"216.201.9.155":   true,
		"51.15.41.46":     true,
		"85.206.172.159":  true,
		"104.244.77.87":   true,
		"37.27.4.53":      true,
		"192.3.165.198":   true,
		"15.204.205.14":   true,
		"103.122.21.50":   true,
		"104.131.98.232":  true,
		"173.249.201.201": true,
		"23.254.228.89":   true,
		"5.102.159.190":   true,
		"65.130.205.148":  true,
		"119.28.71.45":    true,
		"159.69.65.157":   true,
		"160.251.78.190":  true,
		"107.189.7.143":   true,
		"159.65.224.91":   true,
		"185.217.199.21":  true,
		"91.224.92.110":   true,
		"161.97.67.210":   true,
		"51.15.3.74":      true,
		"209.126.11.233":  true,
		"37.187.95.112":   true,
		"167.99.185.219":  true,
		"144.91.88.22":    true,
		"88.99.2.212":     true,
		"37.59.48.81":     true,
		"95.179.130.187":  true,
		"51.15.26.25":     true,
		"192.9.228.30":    true,
	}

	// List of Cloudflare IP ranges
	cfRanges = []string{
		"127.0.0.0/8",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"104.16.0.0/12",
		"108.162.192.0/18",
		"131.0.72.0/22",
		"141.101.64.0/18",
		"162.158.0.0/15",
		"172.64.0.0/13",
		"173.245.48.0/20",
		"188.114.96.0/20",
		"190.93.240.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"::1/128",
		"2400:cb00::/32",
		"2405:8100::/32",
		"2405:b500::/32",
		"2606:4700::/32",
		"2803:f800::/32",
		"2c0f:f248::/32",
		"2a06:98c0::/29",
	}

	// IP ranges
	ipRange []*net.IPNet
)

func checkIfDestinationIsBlocked(ipAddress string) bool {
	return torrentTrackers[ipAddress]
}

type Server struct {
	host string
	port string
}

type Client struct {
	conn net.Conn
}

type Config struct {
	Host string
	Port string
}

func New(config *Config) *Server {
	return &Server{
		host: config.Host,
		port: config.Port,
	}
}

func checkIfSourceIsAllowed(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	for _, r := range ipRange {
		if r.Contains(ip) {
			return true
		}
	}

	return false
}

func init() {
	ipRange = []*net.IPNet{}

	for _, r := range cfRanges {
		_, cidr, err := net.ParseCIDR(r)
		if err != nil {
			continue
		}
		ipRange = append(ipRange, cidr)
	}
}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.host, server.port))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		ip := conn.RemoteAddr().String()
		sh, sp, err := net.SplitHostPort(ip)
		if err != nil {
			log.Printf("unable to parse host %s\n", ip)
			_ = conn.Close()
			continue
		}
		if !checkIfSourceIsAllowed(sh) {
			log.Printf("request from unacceptable source blocked: %s:%s\n", sh, sp)
			_ = conn.Close()
			continue
		}

		client := &Client{
			conn: conn,
		}
		go client.handleRequest()
	}
}

func (client *Client) handleRequest() {
	defer func() {
		_ = client.conn.Close()
	}()
	reader := bufio.NewReader(client.conn)
	header, _ := reader.ReadBytes(byte(13))
	if len(header) < 1 {
		return
	}
	inputHeader := strings.Split(string(header[:len(header)-1]), "@")
	if len(inputHeader) < 2 {
		return
	}
	network := "tcp"
	if inputHeader[0] == "udp" {
		network = "udp"
	}
	address := strings.Replace(inputHeader[1], "$", ":", -1)
	if strings.Contains(address, "temp-mail.org") {
		return
	}

	dh, _, err := net.SplitHostPort(address)
	if err != nil {
		return
	}
	// check if ip is not blocked
	blockFlag := false
	addr := net.ParseIP(dh)
	if addr != nil {
		if checkIfDestinationIsBlocked(dh) {
			blockFlag = true
		}
	} else {
		ips, _ := net.LookupIP(dh)
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				if checkIfDestinationIsBlocked(ipv4.String()) {
					blockFlag = true
				}
			}
		}
	}

	if blockFlag {
		log.Printf("destination host is blocked: %s\n", address)
		return
	}

	if network == "udp" {
		handleUDPOverTCP(client.conn, address)
		return
	}

	// transmit data
	log.Printf("%s Dialing to %s...\n", network, address)

	rConn, err := net.Dial(network, address)

	if err != nil {
		log.Println(fmt.Errorf("failed to connect to socket: %v", err))
		return
	}

	// transmit data
	go Copy(client.conn, rConn)
	Copy(rConn, client.conn)

	_ = rConn.Close()
}

// Copy reads from src and writes to dst until either EOF is reached on src or
// an error occurs. It returns the number of bytes copied and any error
// encountered. Copy uses a fixed-size buffer to efficiently copy data between
// the source and destination.
func Copy(src io.Reader, dst io.Writer) {
	buf := make([]byte, 256*1024)

	_, err := io.CopyBuffer(dst, src, buf[:cap(buf)])
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	var config Config
	flag.StringVar(&config.Host, "b", "0.0.0.0", "Server IP address")
	flag.StringVar(&config.Port, "p", "6666", "Server Port number")
	flag.Parse()
	server := New(&config)
	server.Run()
}
