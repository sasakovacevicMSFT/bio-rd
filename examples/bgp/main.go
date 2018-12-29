package main

import (
	"fmt"
	"log"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
)

func strAddr(s string) uint32 {
	ret, _ := bnet.StrToAddr(s)
	return ret
}

func main() {
	logrus.Printf("This is a BGP speaker\n")

	b := server.NewBgpServer()
	v, err := vrf.New("master")
	if err != nil {
		log.Fatal(err)
	}

	rib := startServer(b, v)

	go func() {
		for {
			fmt.Printf("LocRIB count: %d\n", rib.Count())
			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
