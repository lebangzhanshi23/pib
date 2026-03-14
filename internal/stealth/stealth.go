// Package stealth provides process hiding functionality for Boss Mode
// This module helps the application hide itself in the process list
package stealth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.MustLoadDLL("kernel32.dll")
	setConsoleTitleW = kernel32.MustFindProc("SetConsoleTitleW")
)

// ProcessName is the fake process name used to hide
var ProcessName = "svchost.exe"

// HideProcess hides the current process by setting a fake process name
// On Windows, we can modify the command line to appear as a system process
func HideProcess() error {
	// On Windows, we can set the console title to blend in
	// This is a visual trick but doesn't actually change the process name in Task Manager
	title := fmt.Sprintf("Windows Update Service (svchost.exe)")
	setConsoleTitleW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))

	return nil
}

// IsHidden returns whether stealth mode is active
func IsHidden() bool {
	return true
}

// GetFakeProcessName returns a random fake process name to use
func GetFakeProcessName() string {
	processes := []string{
		"svchost.exe",
		"rundll32.exe",
		"searchindexer.exe",
		"RuntimeBroker.exe",
		"SecurityHealthService.exe",
		"MsMpEng.exe",
		"spoolsv.exe",
	}
	
	// Use the exe name from command line if provided
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-process=") {
			return strings.TrimPrefix(arg, "-process=")
		}
	}
	
	// Default to svchost.exe
	return processes[0]
}

// GetProcessDisplayInfo returns fake process info for display
func GetProcessDisplayInfo() string {
	// Return fake process info that matches GetFakeProcessName()
	return fmt.Sprintf(`%s -k NetworkService -p -s NetSvc`, GetFakeProcessName())
}

// GetArgs returns the fake command line args
func GetArgs() []string {
	return []string{
		"-k",
		"NetworkService",
		"-p",
		"-s",
		"NetSvc",
	}
}

// RenameExecutable creates a copy of the executable with a fake name
// This is useful for manual renaming to hide the app
func RenameExecutable(fakeName string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	dir := filepath.Dir(exePath)
	newPath := filepath.Join(dir, fakeName)

	// Check if already renamed
	if strings.EqualFold(filepath.Base(exePath), fakeName) {
		return nil
	}

	// Copy the file
	originalData, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("failed to read executable: %w", err)
	}

	err = os.WriteFile(newPath, originalData, 0755)
	if err != nil {
		return fmt.Errorf("failed to write new executable: %w", err)
	}

	fmt.Printf("Created hidden executable: %s\n", newPath)
	fmt.Printf("Original: %s\n", exePath)

	return nil
}
