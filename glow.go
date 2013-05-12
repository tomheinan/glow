package glow

import (
	"fmt"
	"net"
	"strings"
)

type Status struct {
	host string
}

func Scan(host string) (status *Status, err error) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", host)
	if err != nil {
		return nil, err
	}

	srv := addrs[0]
	fmt.Println(strings.TrimRight(srv.Target, "."), srv.Port)

	status = new(Status)
	return status, nil
}
