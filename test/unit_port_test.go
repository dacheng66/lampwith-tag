package test

import (
	"testing"
		"unsafe"
	"reflect"
	"syscall"
	"strings"
	"fmt"
)
var (
	nEnumPortsW *syscall.Proc
)

/* definition in C
typedef struct _PORT_INFO_2 {
  LPTSTR pPortName;
  LPTSTR pMonitorName;
  LPTSTR pDescription;
  DWORD  fPortType;
  DWORD  Reserved;
} PORT_INFO_2, *PPORT_INFO_2;
*/

// This is x64
// If you need to test against a C version, be sure to make it x64
type PortInfo2 struct {
	pPortName    *uint16
	pMonitorName *uint16
	pDescription *uint16
	fPortType    uint32
	reserved     uint32
}

func init() {
	wp := syscall.MustLoadDLL("Winspool.drv")
	nEnumPortsW = wp.MustFindProc("EnumPortsW")

	//var pi PortInfo2
	//fmt.Printf("Size of obj is: %v\n", unsafe.Sizeof(pi))
}

func toStr(str *uint16) string {
	strSlice := (*[1 << 30]uint16)(unsafe.Pointer(str))[:]
	return syscall.UTF16ToString(strSlice)
}

func TestPort(t *testing.T)  {
	var cbNeeded, cReturned uint32
	nEnumPortsW.Call(
		0,                 // null
		2,                 // level
		0,                 // return buffer.
		uintptr(cbNeeded), // 0.
		uintptr(unsafe.Pointer(&cbNeeded)),
		uintptr(unsafe.Pointer(&cReturned)),
	)
	//nEnumPorts(NULL, 2, (LPBYTE)pPort, pcbNeeded, &pcbNeeded, &pcReturned);
	// fmt.Printf("cb: %v\n", cbNeeded)
	// fmt.Printf("cReturned: %v\n", cReturned)

	buffer := make([]byte, cbNeeded)
	nEnumPortsW.Call(
		0,
		2,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(cbNeeded), //now we have
		uintptr(unsafe.Pointer(&cbNeeded)),
		uintptr(unsafe.Pointer(&cReturned)),
	)

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&buffer[0])),
		Len:  int(cReturned),
		Cap:  int(cReturned),
	}
	pInfos := *(*[]PortInfo2)(unsafe.Pointer(&hdr))
	pInfos = pInfos[:] // you can even copy all.
	for _, t := range pInfos {
		// And the length of array is 1<<30. Pretty sure unsafe.
		// First make the pointer an array (with length of inifnite)
		// Then make the array into an slice ( cast with `[:]' )
    //
    //   fmt.Println(toStr(t.pPortName))
		if(strings.Contains(toStr(t.pPortName),"COM")){
			/*fmt.Printf("%v - %v - %v\n", toStr(t.pPortName),
				toStr(t.pMonitorName),
				toStr(t.pDescription),
			)*/
			fmt.Print(toStr(t.pPortName))
		}
	}
}