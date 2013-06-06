package glow

import "testing"
import "log"

func TestScan(t *testing.T) {
	for {
		status, err := Scan("mc.craftervillemc.com")
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("%+v\n", status)
		}
	}
}
