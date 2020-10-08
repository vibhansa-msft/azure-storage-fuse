
package fscreator

import (
	"sync"
	FSIntf "../fsinterface"
)

var (
	creatorLock 	sync.RWMutex
    fsList		= make(map[string]FSIntf.CreateObj)
)

// RegisterFS : Registration method for all the implementations to factory
func RegisterFS(fsName string, fs FSIntf.CreateObj) {	
	//fmt.Println("Registering : " + fsName)

	creatorLock.Lock()
	defer creatorLock.Unlock()

	if fs == nil {
		panic ("Can not register empty FileSystem : " + fsName)
	}

	if _, exist := fsList[fsName]; exist {
        panic("FS " + fsName + " already registered")
    }

    fsList[fsName] = fs
}


// GetFileSystem : Factory method to get the object based on name
func GetFileSystem(fsName string) (FSIntf.FileSystem, bool) {
	//fmt.Println("Generating object of : " + fsName)

	creatorLock.Lock()
	defer creatorLock.Unlock()
	
	if fs, exist := fsList[fsName]; exist {
		return fs(), true
	} else {
		return nil, false
	}
}