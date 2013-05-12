package glow

import (
	"log"
	"net"
	"strings"
	"strconv"
	"time"
)

type Status struct {
	host string
}

func Scan(server string) (status *Status, err error) {
	// parse the server into host/port
	host, port, err := net.SplitHostPort(server)
	if err != nil {
		// we weren't given a port; try to find one via dns
		_, addrs, srvErr := net.LookupSRV("minecraft", "udp", server)
		if srvErr != nil {
			_, addrs, srvErr = net.LookupSRV("minecraft", "tcp", server)
		}

		if srvErr != nil {
			host = server
			port = "25565"
		} else {
			addr := addrs[0]
			host = strings.TrimRight(addr.Target, ".")
			port = strconv.FormatInt(int64(addr.Port), 10)
		}
	}

	log.Println(host, port)

	conn, err := net.DialTimeout("udp", net.JoinHostPort(host, port), 3 * time.Second)
	if err != nil {
		log.Println(err)
	}

	defer conn.Close()

	return status, nil
}
