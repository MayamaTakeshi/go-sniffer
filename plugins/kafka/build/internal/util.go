package internal

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

var KafkaMessageMaxBytes = 1 * 1024 * 1024 // 1MB
var KafkaTopicPartitionMaxCount int32 = 4000

func GetNowStr(isClient bool) string {
	var msg string
	layout := "01/02 15:04:05.000000"
	msg += time.Now().Format(layout)
	if isClient {
		msg += "| cli -> ser |"
	} else {
		msg += "| ser -> cli |"
	}
	return msg
}

func IsEof(r io.Reader) bool {
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return true
	}
	return false
}

func ReadOnce() {

}

func ReadByte(r io.Reader) (n byte) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt16(r io.Reader) (n int16) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt32(r io.Reader) (n int32) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadUint32(r io.Reader) (n uint32) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt64(r io.Reader) (n int64, err error) {
	err = binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadString(r io.Reader) (string, int) {

	l := int(ReadInt16(r))

	// -1 => null
	if l <= 0 || l > KafkaMessageMaxBytes {
		return " ", 1
	}

	str := make([]byte, l)
	if _, err := io.ReadFull(r, str); err != nil {
		// panic(err)
		fmt.Println(err.Error())
	}

	return string(str), l
}

//
//func TryReadInt16(r io.Reader) (n int16, err error) {
//
//	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
//		if n == -1 {
//			return 1,nil
//		}
//		panic(err)
//	}
//}

func ReadBytes(r io.Reader) []byte {

	l := int(ReadInt32(r))
	result := make([]byte, 0)

	if l <= 0 || l > KafkaMessageMaxBytes {
		return result
	}

	var b = make([]byte, l)
	for i := 0; i < l; i++ {

		_, err := r.Read(b)

		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		result = append(result, b...)
	}

	return result
}
