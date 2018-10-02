package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

func listDevices() ([]string, error) {
	var ret []string
	_, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		ret = append(ret, fmt.Sprintf("%s %s", desc.Vendor, desc.Product))
		return false
	})
	return ret, err
}

func onAttach(ctx *gousb.Context) ([]string, error) {
	oldDevices, err := listDevices()
	if err != nil {
		return []string{}, err
	}
	for {
		newDevices, err := listDevices()
		if err != nil {
			return []string{}, err
		}
		attached := difference(newDevices, oldDevices)
		if len(attached) > 0 {
			return attached, nil
		}
		oldDevices = newDevices
		time.Sleep(time.Millisecond * 200)
	}
}

func readUSB(ep *gousb.InEndpoint) ([]byte, error) {

	buf := make([]byte, 10*ep.Desc.MaxPacketSize)
	bytesread, err := ep.Read(buf)
	if err != nil {
		return nil, err
	}
	buf2 := buf[:bytesread]
	if debug {
		fmt.Printf("Receiving: --%d-- % x\n", len(buf2), buf2)
	}
	return buf2, nil
}

func sendUSB(ep *gousb.OutEndpoint, data []byte) {
	byteswritten, err := ep.Write(data)
	check(err)
	if debug {
		fmt.Printf("Sending: --%d/%d-- % x\n", byteswritten, len(data), data)
	}
}
