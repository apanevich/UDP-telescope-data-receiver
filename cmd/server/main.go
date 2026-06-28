package main

func main() {
	go startReceiver("9000")
	go storeWorker()
	startAPI()
}

func storeWorker() {
	for frame := range frameChan {
		storeFrame(frame)
	}
}
