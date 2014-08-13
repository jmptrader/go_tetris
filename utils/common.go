package utils

import (
	"encoding/binary"
	"fmt"
	"io"
)

// tcpBuffer set to 1024, always send 1024 bytes to as3
// should test later
const tcpBuffer = 1 << 10

const errTooLargeDatagram = "length of the data pack is %v, too large"

func SendDataOverTcp(w io.Writer, data []byte) (err error) {
	n := len(data)
	if n > tcpBuffer {
		return fmt.Errorf(errTooLargeDatagram, n)
	}
	buf := make([]byte, tcpBuffer)
	binary.BigEndian.PutUint32(buf, uint32(n))
	copy(buf[4:], data)
	_, err = w.Write(buf)
	return err
}

func ReadDataOverTcp(r io.Reader) ([]byte, error) {
	buf := make([]byte, tcpBuffer)
	n, err := io.ReadAtLeast(r, buf[:], 4)
	if err != nil {
		return nil, err
	}
	length := int(binary.BigEndian.Uint32(buf))
	if length > tcpBuffer-4 {
		return nil, fmt.Errorf("length %d is larger than 508\nbytes convert to string is %s\n", length, buf)
	}
	size := length - n + 4
	if size > 0 {
		_, err = io.ReadAtLeast(r, buf[n:], size)
	}
	return buf[4 : length+4], err
}
