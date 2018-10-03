package main

import (
	"errors"
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
	buf = buf[:bytesread]
	return buf, nil
}

type response struct {
	data []byte
	err  error
}

func listen(ep *gousb.InEndpoint) chan response {
	buf := make([]byte, 10*ep.Desc.MaxPacketSize)
	c := make(chan response)

	go func() {
		for {
			bytesread, err := ep.Read(buf)
			if err != nil {
				c <- response{[]byte{}, err}
			}
			buf = buf[:bytesread]
			c <- response{buf, nil}
		}
	}()

	return c
}

func read(c chan response, d time.Duration) ([]byte, error) {
	if d == 0 {
		resp := <-c
		return resp.data, resp.err
	}

	select {
	case resp := <-c:
		return resp.data, resp.err
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
