package main

import (
	"errors"
	"fmt"
	"github.com/google/gousb"
	"time"
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
	buf = buf[:bytesread]
	return buf, nil
}

func listen(ep *gousb.InEndpoint) chan []byte {
	buf := make([]byte, 10*ep.Desc.MaxPacketSize)
	c := make(chan []byte)

	go func() {
		for {
			bytesread, err := ep.Read(buf)
			if err != nil {
				fmt.Printf("listen error: %s\n", err)
			}
			buf = buf[:bytesread]
			c <- buf
		}
	}()

	return c
}

func readTimeout(c chan []byte, d time.Duration) ([]byte, error) {
	select {
	case data := <-c:
		return data, nil
	case <-time.After(d):
		return []byte{}, errors.New("Timeout")
	}
}

func sendUSB(ep *gousb.OutEndpoint, data []byte) {
	byteswritten, err := ep.Write(data)
	check(err)
	if debug {
		fmt.Printf("Sending: --%d/%d-- % x\n", byteswritten, len(data), data)
	}
}
