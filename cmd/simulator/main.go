package main

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"
)

type Packet struct {
	FrameID      uint32
	PacketIndex  uint32
	TotalPackets uint32
	Timestamp    uint64
	DataLength   uint16
	Data         []byte
}

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "localhost:9000")
	conn, _ := net.DialUDP("udp", nil, addr)
	defer conn.Close()

	frameID := uint32(0)
	for {
		frameID++
		totalPackets := uint32(rand.Intn(5) + 1) // 1..5 пакетов на кадр
		dataSize := 1024                         // байт на пакет
		for i := uint32(0); i < totalPackets; i++ {
			data := make([]byte, dataSize)
			rand.Read(data) // случайные данные
			pkt := Packet{
				FrameID:      frameID,
				PacketIndex:  i,
				TotalPackets: totalPackets,
				Timestamp:    uint64(time.Now().UnixMicro()),
				DataLength:   uint16(dataSize),
				Data:         data,
			}
			// Сериализация в бинарный вид (big endian)
			buf := make([]byte, 20+len(pkt.Data))
			binary.BigEndian.PutUint32(buf[0:4], pkt.FrameID)
			binary.BigEndian.PutUint32(buf[4:8], pkt.PacketIndex)
			binary.BigEndian.PutUint32(buf[8:12], pkt.TotalPackets)
			binary.BigEndian.PutUint64(buf[12:20], pkt.Timestamp)
			// DataLength не сохраняем, т.к. он известен из len(data)
			copy(buf[20:], pkt.Data)
			conn.Write(buf)
			time.Sleep(10 * time.Millisecond) // имитация потока
		}
		time.Sleep(100 * time.Millisecond) // пауза между кадрами
	}
}
