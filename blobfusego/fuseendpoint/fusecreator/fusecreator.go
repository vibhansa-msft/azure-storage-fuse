package fusecreator

import (
	"sync"
	FSIntf "../../fswrapper/fsinterface"
)

// CreateObj : Generic functoin that eeryone needs to implement for the factory
type CreateObj  	func()(FuseDriver)

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


// Method to create object based on fuse driver name
var (
	creatorLock 	sync.RWMutex
    fuseList		= make(map[string]CreateObj)
)


// RegisterFuseDriver : Register every fuse driver to system using this
func RegisterFuseDriver(fdName string, fd CreateObj) {	
	//fmt.Println("Registering : " + fdName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if fd == nil {
		panic ("Can not register empty FileSystem : " + fdName)
	}

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
		return fd(), true
	} else {
		return nil, false
	}
}