package glow

import (
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
	Host string
	Port string
	MotD string
	GameType string
	Version string
	Plugins string
	MapName string
	NumPlayers uint16
	MaxPlayers uint16
	Players []string
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

	conn, err := net.DialTimeout("udp", net.JoinHostPort(host, port), 3 * time.Second)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	deadline := time.Now().Add(3 * time.Second)
	conn.SetDeadline(deadline)

	buffer := make([]byte, 16)
	conn.Write(constructChallengeRequest())
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	tokenString := string(buffer[5:(n - 1)])
	token64, _ := strconv.ParseInt(tokenString, 0, 32)
	token := int32(token64)

	buffer = make([]byte, 4096)
	conn.Write(constructQueryRequest(token))
	n, err = conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	packetData := new(bytes.Buffer)
	packetData.Write(buffer[0:n])
	status = parseStatus(packetData)

	status.Host = host
	status.Port = port

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

func parseStatus(buffer *bytes.Buffer) *Status {
	status := new(Status)

	byteSet := bytes.Split(buffer.Bytes(), []byte{0x00})
	for i, val := range byteSet {
		switch i {
		case 4:
			status.MotD = string(val)
		case 6:
			status.GameType = string(val)
		case 10:
			status.Version = string(val)
		case 12:
			status.Plugins = string(val)
		case 14:
			status.MapName = string(val)
		case 16:
			num64, _ := strconv.ParseUint(string(val), 10, 16)
			status.NumPlayers = uint16(num64)
		case 18:
			max64, _ := strconv.ParseUint(string(val), 10, 16)
			status.MaxPlayers = uint16(max64)
		}
	}

	playerSubset := byteSet[26:(len(byteSet) - 2)]
	status.Players = make([]string, len(playerSubset))

	for i, val := range playerSubset {
		status.Players[i] = string(val)
	}

	return status
}
