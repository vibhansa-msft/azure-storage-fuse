package fusecreator

import (
	"sync"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Logger "github.com/blobfusego/global/logger"
)

// FuseDriver : Wrapper which eveyr fuse driver needs to implement
type FuseDriver interface {

	// InitFuse : Initialize the fuse driver
	InitFuse()

	// DeInitFuse : DeInitialize the fuse driver
	DeInitFuse()

	// SetClient : Set the next layer that handles the call
	SetClient(cons FSIntf.FileSystem) int

	// Get the file system name
	GetName() string

	// Print the pipeline
	PrintPipeline() string

	// Start the listener
	Start() int
}

// CreateObjFunc : Generic functoin that factory call to create object of FS
type CreateObjFunc func() FuseDriver

// ReleaseObjFunc : Generic function that factory calls to delete object of FS
type ReleaseObjFunc func()

// FDManager : Method to be used by all implementations to register
type FDManager struct {
	CreateObjFunc
	ReleaseObjFunc
}

// Method to create object based on fuse driver name
var (
	creatorLock sync.RWMutex
	fuseList    = make(map[string]FDManager)
)

// RegisterFuseDriver : Register every fuse driver to system using this
func RegisterFuseDriver(fdName string, fd FDManager) {
	Logger.LogDebug("Registering : " + fdName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if _, exist := fuseList[fdName]; exist {
		panic("FD " + fdName + " already registered")
	}

	fuseList[fdName] = fd
}

// GetFuseDriver : Factory method to get the object based on name
func GetFuseDriver(fdName string) (FuseDriver, bool) {
	Logger.LogDebug("Generating object of : " + fdName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if fd, exist := fuseList[fdName]; exist {
		return fd.CreateObjFunc(), true
	}

	return nil, false
}

// ReleaseFuseDriver : Factory method to release the object
func ReleaseFuseDriver(fd FuseDriver) bool {
	Logger.LogDebug("Releasing object of : " + fd.GetName())

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if fd, exist := fuseList[fd.GetName()]; exist {
		fd.ReleaseObjFunc()
		return true
	}

	return false
}
