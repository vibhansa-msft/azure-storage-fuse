package bazilfuse

import (
	FDFact "../fusecreator"
	FSIntf "../../fswrapper/fsinterface"
	Logger "../../global/logger"
)

type bazilFD struct{
	refCount 		int
	consumer		FSIntf.FileSystem
}

var instance *bazilFD
var fdName = string("bazil")

var regObj = FDFact.FDManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}
func init() {
    FDFact.RegisterFuseDriver(fdName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FDCREATOR TO WPORK

// CreateObj : Create the dummy FS object for factory
func CreateObj() FDFact.FuseDriver {
    if instance == nil {
		instance = &bazilFD{}
		instance.refCount = 0
		Logger.LogDebug("Created first instances of " + fdName)
    }
    instance.refCount++
    return instance
}

// ReleaseObj : Delete the dummy FS object
func ReleaseObj() {
    instance.refCount--
    if instance.refCount == 0 {
		instance = nil
		Logger.LogDebug("Released all instances of " + fdName)
    }
}

////////////////////////////////////////

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
	instance.consumer = cons
	return 0;
}

// Get the file system name
func (f *bazilFD) GetName() string {
	if instance.consumer != nil {
		Logger.LogDebug("Calling FS from FD : " + instance.consumer.GetName())
	}
	return fdName
}



