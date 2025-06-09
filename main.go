package main

import (
	"log"
	"runtime"
	"syscall"
	"unsafe"
)

// Windows API常量和类型定义
const (
	SE_PRIVILEGE_ENABLED = 0x00000002
)

type LUID struct {
	LowPart  uint32
	HighPart int32
}

type LUIDAndAttributes struct {
	Luid       LUID
	Attributes uint32
}

type Tokenprivileges struct {
	PrivilegeCount uint32
	Privileges     [1]LUIDAndAttributes
}

func init() {
	// 在Windows下设置进程特权
	if runtime.GOOS == "windows" {
		err := enableProcessPrivileges()
		if err != nil {
			log.Printf("Warning: Failed to enable process privileges: %v\n", err)
		}
	}
}

func enableProcessPrivileges() error {
	var token syscall.Token
	currentProcess, err := syscall.GetCurrentProcess()
	if err != nil {
		return err
	}

	err = syscall.OpenProcessToken(currentProcess, syscall.TOKEN_ADJUST_PRIVILEGES|syscall.TOKEN_QUERY, &token)
	if err != nil {
		return err
	}
	defer token.Close()

	privs := []string{
		"SeSecurityPrivilege",
		"SeBackupPrivilege",
		"SeRestorePrivilege",
		"SeSystemEnvironmentPrivilege",
		"SeIncreaseWorkingSetPrivilege",
		"SeTimeZonePrivilege",
		"SeCreateSymbolicLinkPrivilege",
	}

	for _, priv := range privs {
		err = EnablePrivilege(token, priv, true)
		if err != nil {
			log.Printf("Warning: Failed to enable %s: %v\n", priv, err)
		}
	}
	return nil
}

func EnablePrivilege(token syscall.Token, privilege string, enable bool) error {
	var luid LUID
	err := lookupPrivilegeValue(nil, privilege, &luid)
	if err != nil {
		return err
	}
	privileges := Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]LUIDAndAttributes{
			{
				Luid:       luid,
				Attributes: SE_PRIVILEGE_ENABLED,
			},
		},
	}
	
	err = adjustTokenPrivileges(token, false, &privileges)
	return err
}

var (
	modadvapi32 = syscall.NewLazyDLL("advapi32.dll")
	procLookupPrivilegeValueW = modadvapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges = modadvapi32.NewProc("AdjustTokenPrivileges")
)

func lookupPrivilegeValue(systemName *uint16, name string, luid *LUID) error {
	lpName, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}

	ret, _, _ := procLookupPrivilegeValueW.Call(
		uintptr(unsafe.Pointer(systemName)),
		uintptr(unsafe.Pointer(lpName)),
		uintptr(unsafe.Pointer(luid)))
	if ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}

func adjustTokenPrivileges(token syscall.Token, disableAllPrivileges bool, newState *Tokenprivileges) error {
	ret, _, _ := procAdjustTokenPrivileges.Call(
		uintptr(token),
		uintptr(boolToUint(disableAllPrivileges)),
		uintptr(unsafe.Pointer(newState)),
		0,
		0,
		0)
	if ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}

func boolToUint(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func main() {
	err := run()
	if err != nil {
		log.Println(err.Error())
	}
}
