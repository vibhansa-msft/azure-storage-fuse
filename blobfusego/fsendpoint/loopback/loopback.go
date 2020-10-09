
package loopback

import (
	"fmt"
    FSIntf "../../fswrapper/fsinterface"
	FSFact "../../fswrapper/fscreator"
	Logger "../../global/logger"
)

type loopbackFS struct{
	refCount 		int
	consumer		FSIntf.FileSystem
}

var instance *loopbackFS
var fsName = string("loopback")

var regObj = FSFact.FSManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}
func init() {
    FSFact.RegisterFileSystem(fsName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FSCREATOR TO WPORK

// CreateObj : Create the loopback FS object for factory
func CreateObj() FSIntf.FileSystem {
    if instance == nil {
		instance = &loopbackFS{}
		instance.refCount = 0
		Logger.LogDebug("Created first instances of " + fsName)
    }
    instance.refCount++
    return instance
}

// ReleaseObj : Delete the loopback FS object
func ReleaseObj() {
    instance.refCount--
    if instance.refCount == 0 {
		instance = nil
		Logger.LogDebug("Released all instances of " + fsName)
    }
}

////////////////////////////////////////

func (f *loopbackFS) InitFS() int {
    return 0
}

func (f *loopbackFS) DeInitFs() int {
    return 0
}

// Get the file system name
func (f *loopbackFS) GetName() string {
	return fsName
}

// Get the ref count
func (f *loopbackFS) GetCount() int {
	return instance.refCount
}

// Set the next component in pipeline for this system
func (f *loopbackFS) SetConsumer(cons FSIntf.FileSystem) int {
	fmt.Println("Set consumer in " + fsName)
	instance.consumer = cons
	return 0;
}

// Get the file system stats
func (f *loopbackFS) StatFS() int {
	fmt.Println(fsName + "Calling next level")
	return instance.consumer.StatFS()
}

// Directory level operations
func (f *loopbackFS) CreateDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) DeleteDir(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) OpenDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) CloseDir(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReadDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) RenameDir(path string, name string) int {
	panic("not implemented") // TODO: Implement
}

// File level operations
func (f *loopbackFS) CreateFile(path string, mode int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) DeleteFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) OpenFile(path string, mode int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) CloseFile(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReadFile(path string, offset int, length int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) WriteFile(path string, offset int, length int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) FlushFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReleaseFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) UnlinkFile(path string) int {
	panic("not implemented") // TODO: Implement
}

// Symlink operations
func (f *loopbackFS) CreateLink(path string, dst string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReadLink(path string, link string) int {
	panic("not implemented") // TODO: Implement
}

// Filesystem level operations
func (f *loopbackFS) GetAttr(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) SetAttr(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) Chmod(path string, mod int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) Chown(path string, owner string) int {
	panic("not implemented") // TODO: Implement
}

