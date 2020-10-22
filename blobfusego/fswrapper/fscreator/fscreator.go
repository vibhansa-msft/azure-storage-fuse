package fscreator

import (
	"sync"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Logger "github.com/blobfusego/global/logger"
	Stats "github.com/blobfusego/global/perfmonitor"
)

// CreateObjFunc : Generic functoin that factory call to create object of FS
type CreateObjFunc func() FSIntf.FileSystem

// ReleaseObjFunc : Generic function that factory calls to delete object of FS
type ReleaseObjFunc func()

// FSManager : Method to be used by all implementations to register
type FSManager struct {
	CreateObjFunc
	ReleaseObjFunc
}

var (
	creatorLock sync.RWMutex
	fsList      = make(map[string]FSManager)
)

// RegisterFileSystem : Registration method for all the implementations to factory
func RegisterFileSystem(fsName string, fs FSManager) {
	Logger.LogDebug("Registering : " + fsName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if _, exist := fsList[fsName]; exist {
		panic("FS " + fsName + " already registered")
	}

	fsList[fsName] = fs
}

// GetFileSystem : Factory method to get the object based on name
func GetFileSystem(fsName string) (FSIntf.FileSystem, bool) {
	Logger.LogDebug("Generating object of : " + fsName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if fs, exist := fsList[fsName]; exist {
		Stats.Create(fsName)
		return fs.CreateObjFunc(), true
	}

	return nil, false
}

// ReleaseFileSystem : Factory method to release the object
func ReleaseFileSystem(fs FSIntf.FileSystem) bool {
	Logger.LogDebug("Releasing object of : " + fs.GetName())

	creatorLock.Lock()
	defer creatorLock.Unlock()

	name := fs.GetName()
	if fs, exist := fsList[name]; exist {
		fs.ReleaseObjFunc()
		Stats.Delete(name)
		return true
	}

	return false
}
