package main

import (
	"fmt"

	"github.com/ksubrmnn/testing/diskutil"
)

func main() {
	var diskNumber int64
	diskNumber = -1
	diskString := "\\\\?\\scsi#disk&ven_google&prod_persistentdisk#4&21cb0360&0&000200#{53f56307-b6bf-11d0-94f2-00a0c91efb8b}"
	err := diskutil.GetDiskNumber(diskString, &diskNumber)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	fmt.Printf("DiskNumber is : %v \n", diskNumber)

	err = diskutil.GetDiskHasPage83Id(diskString, "kalya-test-vm")

	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	fmt.Printf("End of program")
}
