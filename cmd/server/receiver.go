package main

import (
	"encoding/binary"
	"net"
	"sync"
	"time"
)

type FrameBuffer struct {
	FrameID      uint32
	TotalPackets uint32
	Packets      map[uint32][]byte // index -> data
	Received     uint32
	CreatedAt    time.Time
	mu           sync.Mutex
}

type Frame struct {
	FrameID   uint32
	Timestamp uint64
	Data      []byte // склеенные данные
}

var (
	buffers   = make(map[uint32]*FrameBuffer)
	buffersMu sync.Mutex
	frameChan = make(chan Frame, 100) // канал для сохранения готовых кадров
	timeout   = 2 * time.Second
)

func startReceiver(port string) {
	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	go cleanupStaleBuffers() // периодическая очистка

	buf := make([]byte, 65536)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		// Парсим заголовок (20 байт)
		if n < 20 {
			continue
		}
		frameID := binary.BigEndian.Uint32(buf[0:4])
		packetIndex := binary.BigEndian.Uint32(buf[4:8])
		totalPackets := binary.BigEndian.Uint32(buf[8:12])
		timestamp := binary.BigEndian.Uint64(buf[12:20])
		data := make([]byte, n-20)
		copy(data, buf[20:n])

		// Обработка пакета
		processPacket(frameID, packetIndex, totalPackets, timestamp, data)
	}
}

func processPacket(frameID, packetIndex, totalPackets uint32, timestamp uint64, data []byte) {
	buffersMu.Lock()
	defer buffersMu.Unlock()

	fb, ok := buffers[frameID]
	if !ok {
		fb = &FrameBuffer{
			FrameID:      frameID,
			TotalPackets: totalPackets,
			Packets:      make(map[uint32][]byte),
			Received:     0,
			CreatedAt:    time.Now(),
		}
		buffers[frameID] = fb
	}

	fb.mu.Lock()
	defer fb.mu.Unlock()

	if _, exists := fb.Packets[packetIndex]; exists {
		return // дубликат, игнорируем
	}
	fb.Packets[packetIndex] = data
	fb.Received++

	if fb.Received == fb.TotalPackets {
		// Собираем кадр
		var fullData []byte
		for i := uint32(0); i < fb.TotalPackets; i++ {
			part, ok := fb.Packets[i]
			if !ok {
				return // ошибка, хотя теоретически не должно случиться
			}
			fullData = append(fullData, part...)
		}
		frame := Frame{
			FrameID:   frameID,
			Timestamp: timestamp,
			Data:      fullData,
		}
		// Отправляем в канал для сохранения
		frameChan <- frame
		// Удаляем буфер
		delete(buffers, frameID)
	}
}

func cleanupStaleBuffers() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		now := time.Now()
		buffersMu.Lock()
		for id, fb := range buffers {
			fb.mu.Lock()
			if now.Sub(fb.CreatedAt) > timeout {
				// Логируем потерю кадра
				delete(buffers, id)
			}
			fb.mu.Unlock()
		}
		buffersMu.Unlock()
	}
}
