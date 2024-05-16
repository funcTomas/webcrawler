package module

import (
	"fmt"
	"net"
	"strconv"
	"webcrawler/errors"
)

type mAddr struct {
	network string
	address string
}

func (ma *mAddr) Network() string {
	return ma.network
}

func (ma *mAddr) String() string {
	return ma.address
}

func NewAddr(network string, ip string, port uint64) (net.Addr, error) {
	if network != "http" && network != "https" {
		errMsg := fmt.Sprintf("illegal network for module address: %s", network)
		return nil, errors.NewIllegalParameterError(errMsg)
	}
	if parsedIP := net.ParseIP(ip); parsedIP == nil {
		errMsg := fmt.Sprintf("illegal ip for module address: %s", ip)
		return nil, errors.NewIllegalParameterError(errMsg)
	}
	return &mAddr{
		network: network,
		address: ip + ":" + strconv.Itoa(int(port)),
	}, nil
}
