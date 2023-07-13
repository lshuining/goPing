package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	timeOut int64
	size    int
	count   int
	typ     uint8 = 8
	code    uint8 = 0

	sendCount     int
	sueccessCount int
	failCount     int
	totalTs       int64
	minTs         int64
	maxTs         int64

	totalBalan []int64
)

type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	ID          uint16
	SequenceNum uint16
}

func checkSum(data []byte) uint16 {
	length := len(data)
	index := 0
	var sum uint32
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		length -= 2
		index += 2
	}
	if length != 0 {
		sum += uint32(data[index])
	}
	hi16 := sum >> 16
	for hi16 != 0 {
		sum = hi16 + uint32(uint16(sum))
		hi16 = sum >> 16
	}
	return uint16(^sum)
}

func getCommandArgs() {
	flag.Int64Var(&timeOut, "w", 1000, "timeout，请求超时时常,ms.")
	flag.IntVar(&size, "l", 32, "请求发送缓冲区大小，字节")
	flag.IntVar(&count, "n", 6, "发送请求数.")
	flag.Parse()
}

func main() {
	// var a uint16 = 33
	// var b uint16 = 32
	// fmt.Println(a + b)

	getCommandArgs()
	fmt.Println(timeOut, size, count)
	desIp := os.Args[len(os.Args)-1]

	conn, err := net.DialTimeout("ip:icmp", desIp, time.Duration(timeOut)*time.Millisecond)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	fmt.Printf("PING %s [%s] %d bytes of data.\n", desIp, conn.RemoteAddr(), size)
	for i := 0; i < count; i++ {
		t1 := time.Now()
		sendCount++

		icmp := &ICMP{
			Type:        typ,
			Code:        code,
			CheckSum:    0,
			ID:          1,
			SequenceNum: 1,
		}

		data := make([]byte, size)
		// fmt.Println(data)

		var buffer bytes.Buffer
		binary.Write(&buffer, binary.BigEndian, icmp)
		// fmt.Println(buffer)
		buffer.Write(data)
		data = buffer.Bytes()
		// fmt.Println(data)

		checkSum := checkSum(data)
		data[2] = byte(checkSum >> 8) //高位
		data[3] = byte(checkSum)      //低位
		conn.SetDeadline(time.Now().Add(time.Duration(timeOut) * time.Millisecond))
		n, err := conn.Write(data)
		if err != nil {
			failCount++
			log.Println(err)
			continue
		}

		buf := make([]byte, 65535)
		n, err = conn.Read(buf)
		if err != nil {
			failCount++
			log.Panicln(err)
			continue
		}
		// fmt.Println(n, buf)
		ts := time.Since(t1).Milliseconds()
		sueccessCount++
		totalTs += ts
		totalBalan = append(totalBalan, ts)
		// if minTs > ts {
		// 	minTs = ts
		// }
		// if maxTs < ts {
		// 	maxTs = ts
		// }

		fmt.Printf("%d bytes from %d.%d.%d.%d: size=%d icmp_seq=%d ttl=%d time=%d ms\n", n-28, buf[12], buf[13],
			buf[14], buf[15], n-28, i+1, buf[8], ts,
		)
		time.Sleep(time.Second)
	}

	maxTs = totalBalan[0]
	minTs = totalBalan[0]
	for i := 0; i < len(totalBalan); i++ {
		if totalBalan[i] > maxTs {
			maxTs = totalBalan[i]
		} else if totalBalan[i] < minTs {
			minTs = totalBalan[i]
		}
	}

	fmt.Printf("--- %s ping statistics --- \n"+
		"%d packets transmitted, %d received, %.2f%%(%d) packet loss, time %d ms \n"+
		"rtt min/avg/max/mdev = %d/%d/%d/%d ms\n", conn.RemoteAddr(), sendCount, sueccessCount,
		float64(failCount)/float64(sendCount), failCount, totalTs, minTs, totalTs/int64(sendCount), maxTs, 0,
	)
}
