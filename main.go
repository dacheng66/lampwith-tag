package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

const (
	modeNormal  = 3
	modeBreathe = 4
	modeStrobe  = 5
	modeSingle  = 6
	modeMarquee = 7
)

// LampWithClient lampwith client
type LampWithClient struct {
	Client   modbus.Client
	Quantity int

	cStopMarquee chan bool

	ControlMode       int
	ControlPercentage int
	ControlPosition   int
	ControlColor      string
}

func main() {
	portPtr := flag.String("p", "", "port name")

	flag.Parse()

	if *portPtr == "" {
		log.Fatalln("-h for help")
	}

	port := *portPtr

	// new handler
	handler := modbus.NewRTUClientHandler(port)
	handler.BaudRate = 19200
	handler.Timeout = time.Second
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	// handler.Logger = log.New(os.Stdout, "rtu: ", log.LstdFlags)

	err := handler.Connect()
	if err != nil {
		fmt.Printf("connect port [%s] failed, errmsg[%v]\n", port, err)
		return
	}

	fmt.Printf("connect port [%s] success\n", port)

	client := modbus.NewClient(handler)
	inputReader := bufio.NewReader(os.Stdin)

	var q int
	fmt.Printf("please enter the quantity of lampwith's led(default 30): ")
	_, err = fmt.Scanln(&q)
	if err != nil || q <= 0 {
		q = 30
		fmt.Printf("illegal input, use default number [30], control start\n\n")
	} else {
		fmt.Printf("input [%d], control start\n\n", q)
	}

	var lc LampWithClient
	lc.Client = client
	lc.Quantity = q
	lc.cStopMarquee = make(chan bool)

	lc.ControlMode = modeNormal
	lc.ControlPercentage = 100
	lc.ControlPosition = 1
	lc.ControlColor = "3,4,5"

	lc.showHelp()

	lastcommand := ""

LOOP:
	for {
		fmt.Printf("> ")
		input, _, err := inputReader.ReadLine()
		if err != nil {
			fmt.Printf("error input: %v\n", err)
		}

		si := string(input)

		// trim space
		si = strings.Replace(si, " ", "", -1)

		if lastcommand == "10" || lastcommand == "11" || lastcommand == "12" || (lastcommand == "exec" && lc.ControlMode == modeMarquee) {
			lc.cStopMarquee <- true
			time.Sleep(time.Millisecond * 50)
		}

		// percent=[?]			控制的灯珠比例。ex: percent=20 代表控制前20%的灯
		// position=[?]			控制灯珠的位置（单颗灯控制使用）。ex: position=5 代表控制第5颗灯的颜色
		// rgb=[r,g,b]			控制灯的颜色和亮度。ex: rgb=255,0,0 代表设置灯的颜色为红色

		if strings.HasPrefix(si, "percent=") {
			sn := strings.Trim(si, "percent=")
			n, err := strconv.Atoi(sn)
			if err != nil {
				fmt.Printf("illegal input: %s\n", si)
				continue
			}

			if n <= 0 || n > 100 {
				fmt.Printf("percentage should between 1 and 100\n")
				continue
			}

			lc.ControlPercentage = n
			fmt.Printf("set percentage: %d\ntype 'option' for setting or 'exec' for effect.\n", n)
			lastcommand = si
			continue
		} else if strings.HasPrefix(si, "position=") {
			sn := strings.Trim(si, "position=")
			n, err := strconv.Atoi(sn)
			if err != nil {
				fmt.Printf("illegal input: %s\n", si)
				continue
			}

			if n <= 0 || n >= lc.Quantity {
				fmt.Printf("percentage should between 1 and %d\n", lc.Quantity)
				continue
			}

			lc.ControlPosition = n
			fmt.Printf("set position: %d\ntype 'option' for setting or 'exec' for effect.\n", n)
			lastcommand = si
			continue
		} else if strings.HasPrefix(si, "rgb=") {
			sc := strings.Trim(si, "rgb=")
			scs := strings.Split(sc, ",")
			if len(scs) != 3 {
				fmt.Printf("illegal input: %s\n", si)
				continue
			}

			r, _ := strconv.Atoi(scs[0])
			g, _ := strconv.Atoi(scs[1])
			b, _ := strconv.Atoi(scs[2])

			if r < 0 || g < 0 || b < 0 || r > 255 || g > 255 || b > 255 {
				fmt.Printf("color value should be 0 and 255\n")
				continue
			}

			if r == 0 && g == 0 && b == 0 {
				fmt.Printf("color value set zero(0) !!!!!!\n")
			}

			lc.ControlColor = strconv.Itoa(r) + "," + strconv.Itoa(g) + "," + strconv.Itoa(b)
			fmt.Printf("set color: r,g,b=%s\ntype 'option' for setting or 'exec' for effect.\n", lc.ControlColor)
			lastcommand = si
			continue
		}

		switch si {
		case "":
		case "0":
			bval := []byte{0x03, 0x64, 0x00, 0x00, 0x00, 0x00}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "1":
			bval := []byte{0x03, 0x64, 0x00, 0x25, 0x00, 0x00}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "2":
			bval := []byte{0x03, 0x64, 0x00, 0x00, 0x25, 0x00}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "3":
			bval := []byte{0x03, 0x64, 0x25, 0x00, 0x00, 0x00}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "4":
			bval := []byte{0x04, 0x64, 0x00, 0x25, 0x00, 0x05}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "5":
			bval := []byte{0x04, 0x64, 0x00, 0x00, 0x25, 0x05}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "6":
			bval := []byte{0x04, 0x64, 0x25, 0x00, 0x00, 0x05}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "7":
			bval := []byte{0x05, 0x64, 0x00, 0x25, 0x00, 0x64}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "8":
			bval := []byte{0x05, 0x64, 0x00, 0x00, 0x25, 0x64}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "9":
			bval := []byte{0x05, 0x64, 0x25, 0x00, 0x00, 0x64}
			if err := lc.control(bval); err != nil {
				fmt.Printf("contorl err: %v\n", err)
				continue
			}
		case "10":
			go lc.marquee("r")
		case "11":
			go lc.marquee("g")
		case "12":
			go lc.marquee("b")
		case "sma":
			lc.ControlMode = modeNormal
		case "smb":
			lc.ControlMode = modeBreathe
		case "smc":
			lc.ControlMode = modeStrobe
		case "smd":
			lc.ControlMode = modeSingle
		case "sme":
			lc.ControlMode = modeMarquee
		case "option":
			lc.showCurrentOptions()
		case "exec":
			lc.exec()
		case "h":
			lc.showHelp()
		case "q":
			if lastcommand == "10" || lastcommand == "11" || lastcommand == "12" || (lastcommand == "exec" && lc.ControlMode == modeMarquee) {
				lc.cStopMarquee <- true
			}

			bval := []byte{0x03, 0x64, 0x00, 0x00, 0x00, 0x00}
			lc.control(bval)

			break LOOP
		default:
			fmt.Printf("h for help, q for quit\n")
		}

		lastcommand = si
	}
}

func (lc *LampWithClient) control(val []byte) error {
	address := uint16(9)
	quantity := uint16(3)

	_, err := lc.Client.WriteMultipleRegisters(address, quantity, val)
	if err != nil {
		return err
	}

	return nil
}

func (lc *LampWithClient) marquee(color string) {
	value := []byte{}

	switch color {
	case "r":
		value = []byte{0x06, 0x00, 0x00, 0x25, 0x00, 0x00}
	case "g":
		value = []byte{0x06, 0x00, 0x25, 0x00, 0x00, 0x00}
	case "b":
		value = []byte{0x06, 0x00, 0x00, 0x00, 0x25, 0x00}
	default:
		// parse current color
		r, g, b := lc.parseColor()
		value = []byte{0x06, 0x00, g, r, b, 0x00}
	}

OUTLOOP:
	for {
		for i := 1; i < lc.Quantity; i++ {
			value[1] = byte(i)
			lc.control(value)

			select {
			case <-time.After(time.Millisecond * 500):
				// turn off light only
				tmp := []byte{0x06, byte(i), 0x00, 0x00, 0x00, 0x00}
				lc.control(tmp)
			case <-lc.cStopMarquee:
				// turn off light and break loop
				tmp := []byte{0x06, byte(i), 0x00, 0x00, 0x00, 0x00}
				lc.control(tmp)

				break OUTLOOP
			}
		}
	}
}

func (lc *LampWithClient) parseColor() (byte, byte, byte) {
	c := lc.ControlColor
	s := strings.Split(c, ",")
	if len(s) != 3 {
		fmt.Println("illegel color format, use 25,0,0")
		return 25, 0, 0
	}

	r, _ := strconv.Atoi(s[0])
	g, _ := strconv.Atoi(s[1])
	b, _ := strconv.Atoi(s[2])

	return byte(r), byte(g), byte(b)
}

func (lc *LampWithClient) exec() error {
	r, g, b := lc.parseColor()

	switch lc.ControlMode {
	case modeNormal:
		quantity := (lc.ControlPercentage * lc.Quantity) / 100
		value := []byte{0x03, byte(quantity), g, r, b, 0x00}
		lc.control(value)
	case modeBreathe:
		quantity := (lc.ControlPercentage * lc.Quantity) / 100
		value := []byte{0x04, byte(quantity), g, r, b, 0x05}
		lc.control(value)
	case modeStrobe:
		quantity := (lc.ControlPercentage * lc.Quantity) / 100
		value := []byte{0x05, byte(quantity), g, r, b, 0x64}
		lc.control(value)
	case modeSingle:
		value := []byte{0x06, byte(lc.ControlPosition), g, r, b, 0x00}
		lc.control(value)
	case modeMarquee:
		go lc.marquee("")
	}

	return nil
}

func (lc *LampWithClient) showCurrentOptions() {
	mode := ""
	switch lc.ControlMode {
	case modeNormal:
		mode = "常亮"
	case modeBreathe:
		mode = "呼吸"
	case modeStrobe:
		mode = "频闪"
	case modeSingle:
		mode = "单颗灯控制"
	case modeMarquee:
		mode = "跑马灯"
	default:
		mode = "未知"
	}

	fmt.Printf(`Current options:
	mode: %s
	percent: %d
	position: %d
	color: r,g,b=%s

`, mode, lc.ControlPercentage, lc.ControlPosition, lc.ControlColor)
}

func (lc *LampWithClient) showHelp() {
	fmt.Print(`Usage:
	please type below number for preset mode (h for help, q for quit):

	number					 description
	  0 					所有灯灭
	  1 					所有灯常亮：红
	  2 					所有灯常亮：蓝
	  3 					所有灯常亮：绿
	  4 					所有灯呼吸：红
	  5 					所有灯呼吸：蓝
	  6 					所有灯呼吸：绿
	  7 					所有灯频闪：红
	  8 					所有灯频闪：蓝
	  9 					所有灯频闪：绿
	  10 					跑马灯：红
	  11 					跑马灯：蓝
	  12 					跑马灯：绿

	please type below command for special setting:
	
	command					 description
	  sma					设置为常亮模式
	  smb					设置为呼吸模式
	  smc					设置为频闪模式					
	  smd					设置为单颗灯控制模式
	  sme 					设置为跑马灯模式

	percent=[?]				控制的灯珠比例。ex: percent=20 代表控制前20%的灯
	position=[?]			控制灯珠的位置（单颗灯控制使用）。ex: position=5 代表控制第5颗灯的颜色
	rgb=[r,g,b]				控制灯的颜色和亮度。ex: rgb=255,0,0 代表设置灯的颜色为红色

	 option					显示当前配置
	  exec					使用当前配置执行控制

`)
}
