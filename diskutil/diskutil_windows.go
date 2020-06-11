// +build windows

package diskutil

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32DLL = syscall.NewLazyDLL("kernel32.dll")
)

const (
	IOCTL_STORAGE_GET_DEVICE_NUMBER = 0x2D1080
	IOCTL_STORAGE_QUERY_PROPERTY    = 0x2D1400
)

func GetDiskNumber(disk string, number *int64) error {
	h, err := syscall.Open(disk, syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	return DiskNumber(h, number)
}
func GetDiskHasPage83Id(disk string, id string) error {
	h, err := syscall.Open(disk, syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	found := false
	return DiskHasPage83Id(h, id, uint16(len(id)), &found)
}

// can be done with powershell
func DiskNumber(disk syscall.Handle, number *int64) error {
	var bytes uint32
	devNum := StorageDeviceNumber{}
	buflen := uint32(unsafe.Sizeof(devNum.DeviceType)) + uint32(unsafe.Sizeof(devNum.DeviceNumber)) + uint32(unsafe.Sizeof(devNum.PartitionNumber))

	err := syscall.DeviceIoControl(disk, IOCTL_STORAGE_GET_DEVICE_NUMBER, nil, 0, (*byte)(unsafe.Pointer(&devNum)), buflen, &bytes, nil)

	fmt.Printf("devNum: %v \n", devNum)

	if err == nil {
		*number = int64(devNum.DeviceNumber)
	}
	return err
}

func DiskHasPage83Id(disk syscall.Handle, matchID string, matchLen uint16, found *bool) error {
	query := StoragePropertyQuery{}
	//devIDDesc := StorageDeviceIDDescriptor{}

	bufferSize := uint32(4 * 1024)
	buffer := make([]byte, 4*1024)
	var size uint32
	var n uint32
	var m uint16

	*found = false
	query.QueryType = PropertyStandardQuery
	query.PropertyID = StorageDeviceIDProperty

	querySize := uint32(unsafe.Sizeof(query.PropertyID)) + uint32(unsafe.Sizeof(query.QueryType)) + uint32(unsafe.Sizeof(query.Byte))
	querySize = uint32(unsafe.Sizeof(query))
	err := syscall.DeviceIoControl(disk, IOCTL_STORAGE_QUERY_PROPERTY, (*byte)(unsafe.Pointer(&query)), querySize, (*byte)(unsafe.Pointer(&buffer[0])), bufferSize, &size, nil)
	if err != nil {
		return fmt.Errorf("IOCTL_STORAGE_QUERY_PROPERTY failed: %v", err)
	}

	fmt.Print("IOCTL successful \n")

	devIDDesc := (*StorageDeviceIDDescriptor)(unsafe.Pointer(&buffer[0]))
	fmt.Printf("StorageDeviceIDDescriptor: %v \n", devIDDesc)

	pID := (*StorageIdentifier)(unsafe.Pointer(&devIDDesc.Identifiers[0]))

	page83ID := []byte{}
	byteSize := unsafe.Sizeof(byte(0))
	fmt.Printf("Number of Identifiers: %d \n ", devIDDesc.NumberOfIdentifiers)
	for n = 0; n < devIDDesc.NumberOfIdentifiers; n++ {
		fmt.Printf("StorageIdentifier %d: %v \n", n, *pID)
		fmt.Printf("pID.CodeSet: %d \n", pID.CodeSet)
		fmt.Printf("pID.Association: %d \n ", pID.Association)
		if pID.CodeSet == StorageIDCodeSetASCII && pID.Association == StorageIDAssocDevice {
			if matchLen != pID.IdentifierSize {
				fmt.Printf("MatchLen does not equal ID size \n")
			}
			for m = 0; m < matchLen; m++ {
				page83ID = append(page83ID, *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&pID.Identifier[0])) + byteSize*uintptr(m))))
			}
			return fmt.Errorf("Page83ID: %s", string(page83ID))
			*found = true
			return nil
		}
		pID = (*StorageIdentifier)(unsafe.Pointer(uintptr(unsafe.Pointer(pID)) + byteSize*uintptr(pID.NextOffset)))
	}
	return nil
}

type StorageDeviceNumber struct {
	DeviceType      DeviceType
	DeviceNumber    uint32
	PartitionNumber uint32
}
type DeviceType uint32

/*
typedef struct _STORAGE_PROPERTY_QUERY {
	STORAGE_PROPERTY_ID PropertyId;
	STORAGE_QUERY_TYPE  QueryType;
	BYTE                AdditionalParameters[1];
  } STORAGE_PROPERTY_QUERY, *PSTORAGE_PROPERTY_QUERY;
*/

type StoragePropertyID uint32

const (
	StorageDeviceProperty                  StoragePropertyID = 0
	StorageAdapterProperty                                   = 1
	StorageDeviceIDProperty                                  = 2
	StorageDeviceUniqueIDProperty                            = 3
	StorageDeviceWriteCacheProperty                          = 4
	StorageMiniportProperty                                  = 5
	StorageAccessAlignmentProperty                           = 6
	StorageDeviceSeekPenaltyProperty                         = 7
	StorageDeviceTrimProperty                                = 8
	StorageDeviceWriteAggregationProperty                    = 9
	StorageDeviceDeviceTelemetryProperty                     = 10
	StorageDeviceLBProvisioningProperty                      = 11
	StorageDevicePowerProperty                               = 12
	StorageDeviceCopyOffloadProperty                         = 13
	StorageDeviceResiliencyProperty                          = 14
	StorageDeviceMediumProductType                           = 15
	StorageAdapterRpmbProperty                               = 16
	StorageAdapterCryptoProperty                             = 17
	StorageDeviceIoCapabilityProperty                        = 18
	StorageAdapterProtocolSpecificProperty                   = 19
	StorageDeviceProtocolSpecificProperty                    = 20
	StorageAdapterTemperatureProperty                        = 21
	StorageDeviceTemperatureProperty                         = 22
	StorageAdapterPhysicalTopologyProperty                   = 23
	StorageDevicePhysicalTopologyProperty                    = 24
	StorageDeviceAttributesProperty                          = 25
	StorageDeviceManagementStatus                            = 26
	StorageAdapterSerialNumberProperty                       = 27
	StorageDeviceLocationProperty                            = 28
	StorageDeviceNumaProperty                                = 29
	StorageDeviceZonedDeviceProperty                         = 30
	StorageDeviceUnsafeShutdownCount                         = 31
	StorageDeviceEnduranceProperty                           = 32
)

type StorageQueryType uint32

const (
	PropertyStandardQuery StorageQueryType = iota
	PropertyExistsQuery
	PropertyMaskQuery
	PropertyQueryMaxDefined
)

type StoragePropertyQuery struct {
	PropertyID StoragePropertyID
	QueryType  StorageQueryType
	Byte       []AdditionalParameters
}

type AdditionalParameters byte

/*
typedef struct _STORAGE_DEVICE_DESCRIPTOR {
  ULONG            Version;
  ULONG            Size;
  UCHAR            DeviceType;
  UCHAR            DeviceTypeModifier;
  BOOLEAN          RemovableMedia;
  BOOLEAN          CommandQueueing;
  ULONG            VendorIdOffset;
  ULONG            ProductIdOffset;
  ULONG            ProductRevisionOffset;
  ULONG            SerialNumberOffset;
  STORAGE_BUS_TYPE BusType;
  ULONG            RawPropertiesLength;
  UCHAR            RawDeviceProperties[1];
} STORAGE_DEVICE_DESCRIPTOR, *PSTORAGE_DEVICE_DESCRIPTOR;

*/

type StorageDeviceIDDescriptor struct {
	Version             uint32
	Size                uint32
	NumberOfIdentifiers uint32
	Identifiers         [1]byte
}

type StorageIdentifierCodeSet uint32

const (
	StorageIDCodeSetReserved StorageIdentifierCodeSet = 0
	StorageIDCodeSetBinary                            = 1
	StorageIDCodeSetASCII                             = 2
	StorageIDCodeSetUtf8                              = 3
)

type StorageIdentifierType uint32

const (
	StorageIdTypeVendorSpecific           StorageIdentifierType = 0
	StorageIDTypeVendorID                                       = 1
	StorageIDTypeEUI64                                          = 2
	StorageIDTypeFCPHName                                       = 3
	StorageIDTypePortRelative                                   = 4
	StorageIDTypeTargetPortGroup                                = 5
	StorageIDTypeLogicalUnitGroup                               = 6
	StorageIDTypeMD5LogicalUnitIdentifier                       = 7
	StorageIDTypeScsiNameString                                 = 8
)

/*typedef struct _STORAGE_IDENTIFIER {
	STORAGE_IDENTIFIER_CODE_SET CodeSet;
	STORAGE_IDENTIFIER_TYPE     Type;
	USHORT                      IdentifierSize;
	USHORT                      NextOffset;
	STORAGE_ASSOCIATION_TYPE    Association;
	UCHAR                       Identifier[1];
  } STORAGE_IDENTIFIER, *PSTORAGE_IDENTIFIER;
*/
type StorageAssociationType uint32

const (
	StorageIDAssocDevice StorageAssociationType = 0
	StorageIDAssocPort                          = 1
	StorageIDAssocTarget                        = 2
)

type StorageIdentifier struct {
	CodeSet        StorageIdentifierCodeSet
	Type           StorageIdentifierType
	IdentifierSize uint16
	NextOffset     uint16
	Association    StorageAssociationType
	Identifier     [1]byte
}
