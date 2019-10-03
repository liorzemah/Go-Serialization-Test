package main

import (
	"fmt"
	"time"
	"unsafe"
)

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

type NetworkProtocol interface {
	serialize() []byte
	deserialize([]byte)
}

type FileHandler struct{
	filename string
	path string
	id uint64
	size uint64 /* file size*/
}

func(fileHandler FileHandler) print() {
	fmt.Printf("FileHandler: filename = %s, path = %s, fileID = %d, size = %d\n", fileHandler.filename, fileHandler.path, fileHandler.id, fileHandler.size)
}

func(fileHandler FileHandler) serialize() []byte {
	var temp []byte
	temp = make([]byte, 17, 17)

	for i:=7; i>=0; i--{
		temp[i] = uint8(fileHandler.id % 256)
		fileHandler.id /= 256
		temp[8+i] = uint8(fileHandler.size % 256)
		fileHandler.size /= 256
	}

	temp[16] = uint8(len(fileHandler.filename))
	temp = append(temp, fileHandler.filename...)
	temp = append(temp, uint8(len(fileHandler.path)))
	temp = append(temp, fileHandler.path...)
	return temp
}

func(fileHandler *FileHandler) deserialize(rawBytes []byte)  {
	fileHandler.id = uint64(rawBytes[0])
	for i:=1; i<8; i++{
		fileHandler.id <<= 8
		fileHandler.id += uint64(rawBytes[i])
	}

	fileHandler.size = uint64(rawBytes[8])
	for i:=9; i<16; i++{
		fileHandler.size <<= 8
		fileHandler.size += uint64(rawBytes[i])
	}

	filenameLen := rawBytes[16]
	fileHandler.filename = string(rawBytes[17:17+filenameLen])
	//pathLen := rawBytes[17+filenameLen]
	fileHandler.path = string(rawBytes[18+filenameLen:])
}

type FileSlice struct {
	fileID  uint64
	chunkID uint64
	size    uint16 /* slice size (max = packetSize = 65535 - 28 - 18)*/
	data    []byte
}

func(fileSlice FileSlice) print() {
	fmt.Printf("FileSlice: fileID = %d, chunkID = %d, size = %d, data = %s\n", fileSlice.fileID, fileSlice.chunkID, fileSlice.size, fileSlice.data)
}


func(fileSlice FileSlice) serialize() []byte {
	var temp []byte
	temp = make([]byte, 16, 16)

	for i:=7; i>=0; i--{
		temp[i] = uint8(fileSlice.fileID % 256)
		fileSlice.fileID /= 256
		temp[8+i] = uint8(fileSlice.chunkID % 256)
		fileSlice.chunkID /= 256
	}

	temp = append(temp, uint8(fileSlice.size >> 8), uint8(fileSlice.size))
	temp = append(temp, fileSlice.data...)
	return temp
}

func(fileSlice *FileSlice) deserialize(rawBytes []byte)  {
	fileSlice.fileID = uint64(rawBytes[0])
	for i:=1; i<8; i++{
		fileSlice.fileID <<= 8
		fileSlice.fileID += uint64(rawBytes[i])
	}

	fileSlice.chunkID = uint64(rawBytes[8])
	for i:=9; i<16; i++{
		fileSlice.chunkID <<= 8
		fileSlice.chunkID += uint64(rawBytes[i])
	}

	fileSlice.size = uint16(rawBytes[16])
	fileSlice.size <<= 8
	fileSlice.size += uint16(rawBytes[17])

	fileSlice.data = rawBytes[18:]
}

func timeDuration(begin time.Time , end time.Time) string {
	diff := end.Sub(begin)
	if diff.Seconds() > 1 {
		return fmt.Sprintf("%.3f Seconds", diff.Seconds())
	} else if diff.Milliseconds() > 1 {
		return fmt.Sprintf("%d.%d Milliseconds", diff.Milliseconds(), diff.Microseconds() - diff.Milliseconds() * 1000)
	} else if diff.Microseconds() > 1 {
		return fmt.Sprintf("%d.%d Microseconds", diff.Microseconds(), diff.Microseconds() - diff.Milliseconds() * 1000)
	}else {
		return fmt.Sprintf("%d.%d Nanoseconds", diff.Nanoseconds(), diff.Microseconds() - diff.Milliseconds() * 1000)
	}
}

func bytesFormatting(size uint64) string{
	if (size > GB) {
		return fmt.Sprintf("%.3f GB (%d Bytes)", float64(size)/GB, size)
	}
	if (size > MB){
		return fmt.Sprintf("%.3f MB (%d Bytes)", float64(size)/MB, size)
	}
	if (size > KB){
		return fmt.Sprintf("%.3f KB (%d Bytes)", float64(size)/KB, size)
	}
	return fmt.Sprintf("%d Bytes", size)
}

func main() {
	const PAYLOAD_SIZE uint16 = 65535 - 28 /* 2 pow 16 - (udp header size) */ - 18 /* fileSlice size without data field*/
	const SIZE_OF_FILE uint64 = 20 * GB / uint64(PAYLOAD_SIZE) * uint64(PAYLOAD_SIZE)
	const COUNT_OF_PACKETS uint64 = SIZE_OF_FILE / uint64(PAYLOAD_SIZE)


	fmt.Println(bytesFormatting(SIZE_OF_FILE))

	var fileHandler FileHandler
	fileHandler.filename = "lior"
	fileHandler.path = "C:/"
	fileHandler.id = 256
	fileHandler.size = SIZE_OF_FILE



	fileHandler2 := FileHandler{"beta8989", "D:/", 1, SIZE_OF_FILE}

	fileHandler.print()
	fileHandler2.print()

	fmt.Printf("SIZE_OF_FILE is %d bytes\n", SIZE_OF_FILE)

	fmt.Printf("fileHandler.size is from type %T\n", fileHandler.size)
	fmt.Printf("fileHandler.size is from type %T\n", fileHandler2.id)

	temp := fileHandler.serialize()
	fmt.Printf("FileHandler serialized type %T\n", temp)
	fmt.Println("FileHandler serialized size = ", len(temp))
	fmt.Println("FileHandler serialized: ", temp)

	fileHandler2.deserialize(temp)
	fileHandler2.print()

	fileSlice := FileSlice{1, 245, PAYLOAD_SIZE, make([]byte, PAYLOAD_SIZE, PAYLOAD_SIZE)}

	var dataTemp []byte
	dataTemp = make([]byte, PAYLOAD_SIZE, PAYLOAD_SIZE)

	var i uint16 = 0
	for ; i< PAYLOAD_SIZE;i++{ /* len(fileSlice.data) = PAYLOAD_SIZE -> uint16 */
		fileSlice.data[i] = 'a'
		dataTemp[i] = 'b'
	}

	fileSlice.print()
	fmt.Printf("fileSlice.data is from type %T, data len = %d, data cap = %d\n", fileSlice.data, len(fileSlice.data), cap(fileSlice.data))

	fmt.Println("FileHandler struct size = ", unsafe.Sizeof(fileHandler2))
	fmt.Println("FileSlice struct size = ", unsafe.Sizeof(fileSlice))

	temp = fileSlice.serialize()
	fmt.Printf("FileSlice serialized type %T\n", temp)
	fmt.Println("FileSlice serialized size = ", len(temp))
	fmt.Println("FileSlice serialized: ", temp)

	fileSlice.deserialize(temp)
	fileSlice.print()

	start := time.Now()

	var j uint64 = 0
	for ; j< COUNT_OF_PACKETS; j++{
		fileSlice.chunkID = j
		fileSlice.data = dataTemp
		fileSlice.serialize()

		// Send Packet
	}
	fmt.Println(timeDuration(start, time.Now()))

	//fileSlice.deserialize(temp)
	fileSlice.print()

	start = time.Now()

	for j = 0; j< COUNT_OF_PACKETS; j++{

		// Receive Packet as RawBytes and keep in temp
		fileSlice.deserialize(temp)

	}

	fileSlice.print()
	fmt.Println(timeDuration(start, time.Now()))
}
