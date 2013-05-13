package glow

import (
	"log"
	"net"
	"strings"
	"strconv"
	"time"
	"bytes"
	"encoding/binary"
)

var magicBytes = []byte{0xfe, 0xfd}
var challengeByte byte = 0x09
var queryByte byte = 0x00
var clientBytes = []byte{0x67, 0x6c, 0x6f, 0x77}

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

	buffer := make([]byte, 16)
	conn.Write(constructChallengeRequest())
	n, err := conn.Read(buffer)

	tokenString := string(buffer[5:(n - 1)])
	token64, _ := strconv.ParseInt(tokenString, 0, 32)
	token := int32(token64)

	buffer = make([]byte, 2048)
	conn.Write(constructQueryRequest(token))
	n, err = conn.Read(buffer)

	parseStatus(buffer)

	return status, nil
}

func constructChallengeRequest() []byte {
	buffer := new(bytes.Buffer)

	buffer.Write(magicBytes)
	buffer.WriteByte(challengeByte)
	buffer.Write(clientBytes)

	return buffer.Bytes()
}

func constructQueryRequest(token int32) []byte {
	buffer := new(bytes.Buffer)

	buffer.Write(magicBytes)
	buffer.WriteByte(queryByte)
	buffer.Write(clientBytes)
	binary.Write(buffer, binary.BigEndian, token)
	buffer.Write(clientBytes)

	return buffer.Bytes()
}

func parseStatus(statusBytes []byte) {
	log.Println(statusBytes)
}
