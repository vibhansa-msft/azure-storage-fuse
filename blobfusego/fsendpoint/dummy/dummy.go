package dummy

import (
	"os"

	FSFact "github.com/blobfusego/fswrapper/fscreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Logger "github.com/blobfusego/global/logger"
)

type dummyFS struct {
	refCount int
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
// Init/DeInit the filesystem
func (f *dummyFS) InitFS() int {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) DeInitFs() int {
	panic("not implemented") // TODO: Implement
}

// Set the next component in pipeline for this system
func (f *dummyFS) SetClient(cons FSIntf.FileSystem) int {
	panic("not implemented") // TODO: Implement
}

// Get the file system name
func (f *dummyFS) GetName() string {
	panic("not implemented") // TODO: Implement
}

// Get the reference count
func (f *dummyFS) GetCount() int {
	panic("not implemented") // TODO: Implement
}

// Print the pipeline
func (f *dummyFS) PrintPipeline() string {
	panic("not implemented") // TODO: Implement
}

// Get the file system stats
func (f *dummyFS) StatFS() error {
	panic("not implemented") // TODO: Implement
}

// Directory level operations
func (f *dummyFS) CreateDir(_ string, _ os.FileMode) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) DeleteDir(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) OpenDir(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) CloseDir(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReadDir(_ string) ([]FSIntf.BlobAttr, error) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) RenameDir(_ string, _ string) error {
	panic("not implemented") // TODO: Implement
}

// File level operations
func (f *dummyFS) CreateFile(_ string, _ os.FileMode) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) DeleteFile(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) OpenFile(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) CloseFile(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReadFile(_ string, _ int64, _ int64) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) WriteFile(_ string, _ int64, _ int64, _ []byte) (int, error) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) TruncateFile(_ string, _ int64) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) FlushFile(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReleaseFile(_ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) UnlinkFile(_ string) error {
	panic("not implemented") // TODO: Implement
}

// Symlink operations
func (f *dummyFS) CreateLink(_ string, _ string) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) ReadLink(_ string) (string, error) {
	panic("not implemented") // TODO: Implement
}

// Filesystem level operations
func (f *dummyFS) GetAttr(_ string) (FSIntf.BlobAttr, error) {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) SetAttr(_ string, _ FSIntf.BlobAttr) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) Chmod(_ string, _ os.FileMode) error {
	panic("not implemented") // TODO: Implement
}

func (f *dummyFS) Chown(_ string, _ string) error {
	panic("not implemented") // TODO: Implement
}
