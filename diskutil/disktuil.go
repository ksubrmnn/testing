// +build linux darwin

package diskutil

import (
	"fmt"
)

func GetDiskNumber(disk string, number *int64) error {
	return fmt.Errorf("Unimplemented")
}

func GetDiskHasPage83Id(disk string, id string) error {
	return fmt.Errorf("Unimplemented")
}
