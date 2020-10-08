package bazilfuse

import (
	"fmt"
	FDFact "../fusecreator"
	FSIntf "../../fswrapper/fsinterface"
)

type bazilFD struct{
	refCount 		int
	consumer		FSIntf.FileSystem
}

var instance *bazilFD
var fdName = string("bazil")

func init() {
    FDFact.RegisterFuseDriver(fdName, CreateObj)
}

// CreateObj : Create the dummy FS object for factory
func CreateObj() FDFact.FuseDriver {
    if instance == nil {
		instance = &bazilFD{}
		instance.refCount = 0
    }
    instance.refCount++
    return instance
}

// ReleaseObj : Delete the dummy FS object
func ReleaseObj() {
    instance.refCount--
    if instance.refCount == 0 {
		instance = nil
    }
}

// InitFuse : Initialize the fuse driver
func (f *bazilFD) InitFuse() {
	panic("not implemented") // TODO: Implement
}

// DeInitFuse : DeInitialize the fuse driver
func (f *bazilFD) DeInitFuse() {
	panic("not implemented") // TODO: Implement
}

// SetConsumer : Set the next layer that handles the call
func (f *bazilFD) SetConsumer(cons FSIntf.FileSystem) int {
	fmt.Println("Set consumer in " + fdName)
	instance.consumer = cons
	return 0;
}

// Get the file system name
func (f *bazilFD) GetName() string {
	fmt.Println("Calling FS from FD : " + instance.consumer.GetName())
	return fdName
}



