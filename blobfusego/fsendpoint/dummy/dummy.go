package dummy


import (
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	FSFact "github.com/blobfusego/fswrapper/fscreator"
	Logger "github.com/blobfusego/global/logger"
)

type dummyFS struct{
	refCount 		int
}

var instance *dummyFS
var fsName = string("dummy")

var regObj = FSFact.FSManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}
func init() {
    FSFact.RegisterFileSystem(fsName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FSCREATOR TO WPORK

// CreateObj : Create the dummy FS object for factory
func CreateObj() FSIntf.FileSystem {
    if instance == nil {
		instance = &dummyFS{}
		instance.refCount = 0
		Logger.LogDebug("Created first instances of " + fsName)
    }
    instance.refCount++
    return instance
}

// ReleaseObj : Delete the dummy FS object
func ReleaseObj() {
    instance.refCount--
    if instance.refCount == 0 {
		Logger.LogDebug("Released all instances of " + fsName)
		instance = nil
    }
}

////////////////////////////////////////



func (f *dummyFS) InitFS() int {
    return 0
}

func (f *dummyFS) DeInitFs() int {
    return 0
}

// Get the file system name
func (f *dummyFS) GetName() string {
	return fsName
}

// Get the ref count
func (f *dummyFS) GetCount() int {
	return instance.refCount
}

// Set the next component in pipeline for this system
func (f *dummyFS) SetConsumer(cons FSIntf.FileSystem) int {
	return 0;
}

// Get the file system stats
func (f *dummyFS) StatFS() int {
	Logger.LogDebug(fsName + "called at the last level")
	return 0
}

// PrintPipeline : Print the current pipeline
func (f *dummyFS) PrintPipeline() string {
	return (fsName + " -> X ")
}

// Directory level operations
func (f *dummyFS) CreateDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) DeleteDir(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) OpenDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) CloseDir(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReadDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) RenameDir(path string, name string) int {
	panic("not implemented") // TODO: Implement
}

// File level operations
func (f *dummyFS) CreateFile(path string, mode int) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) DeleteFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) OpenFile(path string, mode int) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) CloseFile(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReadFile(path string, offset int, length int) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) WriteFile(path string, offset int, length int) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) FlushFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReleaseFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) UnlinkFile(path string) int {
	panic("not implemented") // TODO: Implement
}

// Symlink operations
func (f *dummyFS) CreateLink(path string, dst string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReadLink(path string, link string) int {
	panic("not implemented") // TODO: Implement
}

// Filesystem level operations
func (f *dummyFS) GetAttr(path string, attr *FSIntf.BlobAttr) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) SetAttr(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) Chmod(path string, mod int) int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) Chown(path string, owner string) int {
	panic("not implemented") // TODO: Implement
}

