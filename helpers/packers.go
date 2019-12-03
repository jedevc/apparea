package helpers

import (
	"encoding/binary"
	"fmt"
)

func PackInt(payload *[]byte, i uint32) {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, i)
	*payload = append(*payload, bs...)
}

func PackString(payload *[]byte, s string) {
	// append string length
	bs := make([]byte, 4)
	length := uint32(len(s))
	binary.BigEndian.PutUint32(bs, length)
	*payload = append(*payload, bs...)

	// append string
	*payload = append(*payload, []byte(s)...)
}

func UnpackInt(payload *[]byte) (uint32, error) {
	// extract int
	if len(*payload) < 4 {
		return 0, fmt.Errorf("unpack error: no more bytes to read")
	}
	i := binary.BigEndian.Uint32((*payload)[:4])

	// re-adjust buffer
	*payload = (*payload)[4:]

	return i, nil
}

func UnpackString(payload *[]byte) (string, error) {
	// unpack string length
	length, err := UnpackInt(payload)
	if err != nil {
		return "", err
	}

	// extract string
	if length > uint32(len(*payload)) {
		return "", fmt.Errorf("unpack error: invalid length")
	}
	s := string((*payload)[:length])

	// re-adjust buffer
	*payload = (*payload)[length:]

	return s, nil
}
