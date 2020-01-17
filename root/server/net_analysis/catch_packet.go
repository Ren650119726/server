package main

import (
	"bytes"
	"root/core/log"
	"root/core/log/colorized"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"time"
)

var ip_mac map[string][]string

func captrue_packet() {
	ip_mac = make(map[string][]string)
	handle, err := pcap.OpenLive("eno1", 65535, true, pcap.BlockForever)
	if err != nil {
		fmt.Printf("Error:%v", err.Error())
		return
	}

	defer handle.Close()

	src := gopacket.NewPacketSource(handle, handle.LinkType())
	in := src.Packets()

	log.Infof(colorized.Magenta("开启抓包monitor"))
	for {
		var packet gopacket.Packet
		select {
		case packet = <-in:

			tcpLayer := packet.Layer(layers.LayerTypeIPv4)
			if tcpLayer != nil {
				tcp_ip := tcpLayer.(*layers.IPv4)

				var ip string
				var mac string

				ip = tcp_ip.DstIP.String()
				ethLayer := packet.Layer(layers.LayerTypeEthernet)
				if ethLayer != nil {
					eth := ethLayer.(*layers.Ethernet)
					mac = eth.DstMAC.String()
				}

				if arr, exist := ip_mac[ip]; exist {
					insert := true
					for _, oldmac := range arr {
						if oldmac == mac {
							insert = false
							break
						}
					}
					if insert {
						ip_mac[ip] = append(ip_mac[ip], mac)
						log.Infof("insert ip:[%v] mac:%v", ip, ip_mac[ip])
					}
				} else {
					ip_mac[ip] = make([]string, 0)
					ip_mac[ip] = append(ip_mac[ip], mac)
					log.Infof("new ip:[%v] mac:%v", ip, ip_mac[ip])
				}
			}
		}
	}
}

func test() {
	handle, err := pcap.OpenLive("eno1", 65535, false, time.Millisecond*10)
	if err != nil {
		fmt.Printf("Error:%v", err.Error())
		return
	}

	defer handle.Close()

	src := gopacket.NewPacketSource(handle, handle.LinkType())
	count := uint64(0)
	pktNonTcp := uint64(0)
	pktTcp := uint64(0)
	fmt.Println("test([]testSequence{")
	for packet := range src.Packets() {
		count++
		tcp := packet.Layer(layers.LayerTypeTCP)
		if tcp == nil {
			pktNonTcp++
			continue
		} else {
			pktTcp++
			tcp := tcp.(*layers.TCP)
			//fmt.Printf("packet: %s\n", tcp)
			var b bytes.Buffer
			b.WriteString("{\n")
			// TCP
			b.WriteString("tcp: layers.TCP{\n")
			if tcp.SYN {
				b.WriteString("  SYN: true,\n")
			}
			if tcp.ACK {
				b.WriteString("  ACK: true,\n")
			}
			if tcp.RST {
				b.WriteString("  RST: true,\n")
			}
			if tcp.FIN {
				b.WriteString("  FIN: true,\n")
			}
			b.WriteString(fmt.Sprintf("  SrcPort: %d,\n", tcp.SrcPort))
			b.WriteString(fmt.Sprintf("  DstPort: %d,\n", tcp.DstPort))
			b.WriteString(fmt.Sprintf("  Seq: %d,\n", tcp.Seq))
			b.WriteString(fmt.Sprintf("  Ack: %d,\n", tcp.Ack))
			b.WriteString("  BaseLayer: layers.BaseLayer{Payload: []byte{")
			for _, p := range tcp.Payload {
				b.WriteString(fmt.Sprintf("%d,", p))
			}
			b.WriteString("}},\n")
			b.WriteString("},\n")
			// CaptureInfo
			b.WriteString("ci: gopacket.CaptureInfo{\n")
			ts := packet.Metadata().CaptureInfo.Timestamp
			b.WriteString(fmt.Sprintf("  Timestamp: time.Unix(%d,%d),\n", ts.Unix(), ts.Nanosecond()))
			b.WriteString("},\n")
			// Struct
			b.WriteString("},\n")
			fmt.Print(b.String())
		}

	}
	fmt.Println("})")
}
