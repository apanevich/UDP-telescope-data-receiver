package main

import "sync"

var (
	frames    []Frame
	framesMu  sync.RWMutex
	maxFrames = 10000 // ограничим количество хранимых кадров
)

func storeFrame(frame Frame) {
	framesMu.Lock()
	defer framesMu.Unlock()
	frames = append(frames, frame)
	if len(frames) > maxFrames {
		// сдвиг или удаление старых
		frames = frames[len(frames)-maxFrames:]
	}
}

func getFrames(from, to uint32) []Frame {
	framesMu.RLock()
	defer framesMu.RUnlock()
	var result []Frame
	for _, f := range frames {
		if f.FrameID >= from && f.FrameID <= to {
			result = append(result, f)
		}
	}
	return result
}
