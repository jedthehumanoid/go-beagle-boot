package main

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

func listDevices() []string {
	var ret []string
	_, _ = ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		ret = append(ret, fmt.Sprintf("%s %s", desc.Vendor, desc.Product))
		return false
	})
	return ret
}

func onAttach(ctx *gousb.Context) string {
	oldDevices := listDevices()
	for {
		newDevices := listDevices()
		attached := difference(newDevices, oldDevices)
		if len(attached) > 0 {
			return attached[0]
		}
		oldDevices = newDevices
		time.Sleep(time.Millisecond * 100)
	}
}

func readUSB(ep *gousb.InEndpoint) []byte {

	buf := make([]byte, 10*ep.Desc.MaxPacketSize)
	bytesread, err := ep.Read(buf)
	check(err)
	buf2 := buf[:bytesread]
	if debug {
		fmt.Printf("Receiving: --%d-- % x\n", len(buf2), buf2)
	}
	return buf2
}

func sendUSB(ep *gousb.OutEndpoint, data []byte) {
	byteswritten, err := ep.Write(data)
	check(err)
	if debug {
		fmt.Printf("Sending: --%d/%d-- % x\n", byteswritten, len(data), data)
	}
}
