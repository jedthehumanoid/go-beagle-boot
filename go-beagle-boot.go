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

const ROMID = "0451 6141"
const SPLID = "0525 a4a2"
const UMSID = "0451 d022"

const ipUDP = 17
const rndisSize = 44
const etherSize = 14
const arpSize = 28
const ipSize = 20
const udpSize = 8
const bootpSize = 300
const tftpSize = 4
const fullSize = 386

var debug = false

var serverHwaddr = [6]byte{0x9a, 0x1f, 0x85, 0x1c, 0x3d, 0x0e}
var serverIP = [4]byte{192, 168, 1, 9}
var bbIP = [4]byte{192, 168, 1, 3}

const maxbuf = 450

var ctx *gousb.Context

var binPath string

type configuration struct {
	vid    gousb.ID
	pid    gousb.ID
	config int
	intf   int
	alt    int
	in     int
	out    int
}

var ROMConf = configuration{0x0451, 0x6141, 1, 1, 0, 1, 2}
var SPLConf = configuration{0x0525, 0xa4a2, 2, 1, 0, 1, 1}

func open(conf configuration, file string) bool {
	dev, err := ctx.OpenDeviceWithVIDPID(conf.vid, conf.pid)
	if err == gousb.ErrorAccess {
		fmt.Println("Access denied")
		return false
	}
	if dev == nil {
		return false
	}
	check(err)
	defer dev.Close()

	err = dev.SetAutoDetach(true)
	check(err)

	config, err := dev.Config(conf.config)
	check(err)
	defer config.Close()

	if file == "spl" && (runtime.GOOS == "windows" || runtime.GOOS == "darwin") {
		initRNDIS(dev)
	}
	intf, err := config.Interface(conf.intf, conf.alt)
	check(err)
	defer intf.Close()

	iep, err := intf.InEndpoint(conf.in)
	check(err)

	oep, err := intf.OutEndpoint(conf.out)
	check(err)

	inchan := listen(iep)

	for {
		indata, err := read(inchan, 0)
		if err != nil {
			return false
		}
		request := identifyRequest(indata, len(file))
		var data []byte
		if request == "BOOTP" {
			fmt.Print("bootp")
			data, _ = processBOOTP(indata, file)
			sendUSB(oep, data)
		} else if request == "ARP" {
			fmt.Print(", arp")
			data, _ = processARP(indata)
			sendUSB(oep, data)
		} else if request == "TFTP" {
			fmt.Print(", tftp\n\n")
			data, _ = processTFTP(indata, file)
			sendUSB(oep, data)
		} else if request == "TFTP_Data" {
			fmt.Print(".")
			data, _ = processTFTPData(indata, file)
			sendUSB(oep, data)
			if string(data) == "" {
				fmt.Print("\n")
				return false
			}
		}
	}
	return true
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

	args := []string{"g_mass_storage", "file=/dev/sda", "removable=y"}

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
		device, err := onAttach(ctx)
		if err != nil {
			panic(err)
		}
		if contains(device, ROMID) {
			if *exportEnabled {
				unexport()
			}
			fmt.Println("Found Beaglebone in ROM mode, sending SPL")
			open(ROMConf, "spl")
		} else if contains(device, SPLID) {
			fmt.Println("Found Beaglebone in SPL mode, sending UBOOT")
			time.Sleep(time.Second)
			open(SPLConf, "uboot")
			fmt.Println("\nDone!")
		} else if contains(device, UMSID) {
			fmt.Println("found mass storage")
			waitforMassStorage()

			if *exportEnabled {
				export()
			}
		}
	}
}
