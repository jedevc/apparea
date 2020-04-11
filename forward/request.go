package forward

import (
	"fmt"
	"strconv"

	"github.com/jedevc/AppArea/helpers"
)

type ForwardRequest struct {
	Host string
	Port uint32
}

func ParseForwardRequest(payload []byte) (ForwardRequest, error) {
	req := ForwardRequest{}

	host, err := helpers.UnpackString(&payload)
	if err != nil {
		return req, err
	}
	req.Host = host

	port, err := helpers.UnpackInt(&payload)
	if err != nil {
		return req, err
	}
	req.Port = port

	if len(payload) != 0 {
		return req, fmt.Errorf("Forward request parse error: Unknown excess data")
	}

	return req, nil
}

func (fr ForwardRequest) Address() string {
	return fr.Host + ":" + strconv.FormatUint(uint64(fr.Port), 10)
}
