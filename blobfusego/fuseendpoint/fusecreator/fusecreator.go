package fusecreator

import (
	"sync"
	FSIntf "../../fswrapper/fsinterface"
)

// FuseDriver : Wrapper which eveyr fuse driver needs to implement
type FuseDriver interface {

	// InitFuse : Initialize the fuse driver
	InitFuse()
	
	// DeInitFuse : DeInitialize the fuse driver
	DeInitFuse()

	// SetConsumer : Set the next layer that handles the call
	SetConsumer(cons FSIntf.FileSystem) int
	
	// Get the file system name
	GetName() string
}



// CreateObjFunc : Generic functoin that factory call to create object of FS
type CreateObjFunc  	func()(FuseDriver)
// ReleaseObjFunc : Generic function that factory calls to delete object of FS
type ReleaseObjFunc		func()()

// FDManager : Method to be used by all implementations to register 
type FDManager struct{
	CreateObjFunc
	ReleaseObjFunc
}

// Method to create object based on fuse driver name
var (
	creatorLock 	sync.RWMutex
    fuseList		= make(map[string]FDManager)
)


// RegisterFuseDriver : Register every fuse driver to system using this
func RegisterFuseDriver(fdName string, fd FDManager) {	
	//fmt.Println("Registering : " + fdName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if _, exist := fuseList[fdName]; exist {
        panic("FD " + fdName + " already registered")
    }

    fuseList[fdName] = fd
}


// GetFuseDriver : Factory method to get the object based on name
func GetFuseDriver(fdName string) (FuseDriver, bool) {
	//fmt.Println("Generating object of : " + fsName)

	creatorLock.Lock()
	defer creatorLock.Unlock()
	
	if fd, exist := fuseList[fdName]; exist {
		return fd.CreateObjFunc(), true
	} else {
		return nil, false
	}
}

// ReleaseFuseDriver : Factory method to release the object
func ReleaseFuseDriver(fd FuseDriver) bool{
	//fmt.Println("Generating object of : " + fsName)

	creatorLock.Lock()
	defer creatorLock.Unlock()
	
	if fd, exist := fuseList[fd.GetName()]; exist {
		fd.ReleaseObjFunc()
		return true
	} 
	
	return false
}