package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/google/gousb"
)

const romvid gousb.ID = 0x0451
const rompid gousb.ID = 0x6141

const splvid gousb.ID = 0x525
const splpid gousb.ID = 0xa4a2

const ipUDP = 17
const rndisSize = 44
const etherSize = 14
const arpSize = 28
const ipSize = 20
const udpSize = 8
const bootpSize = 300
const tftpSize = 4
const fullSize = 386
const initrndis = false

var debug = false

var serverHwaddr = [6]byte{0x9a, 0x1f, 0x85, 0x1c, 0x3d, 0x0e}
var serverIP = [4]byte{192, 168, 1, 9}
var bbIP = [4]byte{192, 168, 1, 3}

const maxbuf = 450

var ctx *gousb.Context

var binPath string

func sendSPL() bool {

	dev, err := ctx.OpenDeviceWithVIDPID(romvid, rompid)
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()

	dev.SetAutoDetach(true)

	config, err := dev.Config(1)
	check(err)
	defer config.Close()

	if initrndis == true || runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		initRNDIS(dev)
	}
	intf, err := config.Interface(1, 0)
	check(err)
	defer intf.Close()

	iep, err := intf.InEndpoint(1)
	check(err)

	oep, err := intf.OutEndpoint(2)
	check(err)

	transfer(iep, oep, "spl")
	return true
}

func sendUBOOT() bool {

	dev, err := ctx.OpenDeviceWithVIDPID(splvid, splpid)
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()

	dev.SetAutoDetach(true)

	config, err := dev.Config(2)
	check(err)
	defer config.Close()

	intf, err := config.Interface(1, 0)
	check(err)
	defer intf.Close()

	iep, err := intf.InEndpoint(1)
	check(err)

	oep, err := intf.OutEndpoint(1)
	check(err)

	transfer(iep, oep, "uboot")
	return true
}

func transfer(in *gousb.InEndpoint, out *gousb.OutEndpoint, filename string) {
	for {
		in, err := readUSB(in)
		if err != nil {
			return
		}

		request := identifyRequest(in, len(filename))
		var data []byte

		if request == "BOOTP" {
			fmt.Print("bootp")
			data, _ = processBOOTP(in, filename)
		} else if request == "ARP" {
			fmt.Print(", arp")
			data, _ = processARP(in)
		} else if request == "TFTP" {
			fmt.Print(", tftp\n\n")
			data, _ = processTFTP(in, filename)
		} else if request == "TFTP_Data" {
			fmt.Print(".")
			data, _ = processTFTPData(in, filename)
			if string(data) == "" {
				fmt.Print("\n")
				return
			}
		}
		if string(data) != "" {
			sendUSB(out, data)
		}
	}
}

func initRNDIS(dev *gousb.Device) {
	var rtsend uint8 = gousb.ControlOut | gousb.ControlClass | gousb.ControlInterface
	var rtreceive uint8 = gousb.ControlIn | gousb.ControlClass | gousb.ControlInterface

	fmt.Println("Initiating RNDIS...")

	rndis := rndisInitMsg{2, 24, 1, 1, 1, 64}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, rndis)
	check(err)

	_, err = dev.Control(rtsend, 0, 0, 0, buf.Bytes())
	check(err)

	rec := make([]byte, 1025)

	i, err := dev.Control(rtreceive, 0x01, 0, 0, rec)
	check(err)
	rec = rec[:i]

	rndisset := rndisSetMsg{5, 28, 23, 0x1010E, 4, 20, 0, 0x2d}
	buf = new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rndisset)
	check(err)

	_, err = dev.Control(rtsend, 0, 0, 0, buf.Bytes())
	check(err)

	i, err = dev.Control(rtreceive, 0x01, 0, 0, rec)
	check(err)
	rec = rec[:i]
}

func export() {
	modprobe, err := exec.LookPath("modprobe")
	if err != nil {
		panic(err)
	}

	args := []string{"g_mass_storage", "file=/dev/sda"}

	err = exec.Command(modprobe, args...).Run()
	if err != nil {
		panic(err)
	}
}

func unexport() {
	modprobe, err := exec.LookPath("rmmod")
	if err != nil {
		panic(err)
	}

	args := []string{"g_mass_storage"}

	exec.Command(modprobe, args...).Run()
}

func waitforMassStorage() {
	for {
		_, err := os.Stat("/dev/sda")
		if err == nil {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func main() {

	exportEnabled := flag.Bool("export", false, "Export as mass storage")
	flag.StringVar(&binPath, "bin", "bin", "Path to binary images")
	flag.Parse()

	ctx = gousb.NewContext()
	defer ctx.Close()
	fmt.Println("Connect Beaglebone (with boot-button pressed)")

	for true {
		device := onAttach(ctx)
		if device == "0451 6141" {
			if *exportEnabled {
				unexport()
			}
			fmt.Println("Found Beaglebone in ROM mode, sending SPL")
			sendSPL()
		} else if device == "0525 a4a2" {
			fmt.Println("Found Beaglebone in SPL mode, sending UBOOT")
			time.Sleep(time.Second)
			sendUBOOT()
			fmt.Println("\nDone!")
		} else if device == "0451 d022" {
			fmt.Println("found mass storage")
			waitforMassStorage()

			if *exportEnabled {
				export()
			}

			ctx.Close()
			os.Exit(0)
		}
	}
}
