package azurestorage

import (
	"context"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
	FSFact "github.com/blobfusego/fswrapper/fscreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Logger "github.com/blobfusego/global/logger"
)

// Example for azblob usage : https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#pkg-examples
// For methods help refer : https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#ContainerURL

type azurestorageFS struct {
	refCount     int
	serviceURL   azblob.ServiceURL
	containerURL azblob.ContainerURL
	ctx          context.Context
}

var instance *azurestorageFS
var fsName = string("azurestorage")

var regObj = FSFact.FSManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}

func init() {
	FSFact.RegisterFileSystem(fsName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FSCREATOR TO WPORK

// CreateObj : Create the loopback FS object for factory
func CreateObj() FSIntf.FileSystem {
	if instance == nil {
		instance = &azurestorageFS{}
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

// Expect these from global config
// BlobfuseConfig.StoreAccountName, BlobfuseConfig.Container, BlobfuseConfig.StoreAuthType, BlobfuseConfig.StoreAccountKey

////////////////////////////////////////
// Init/DeInit the filesystem
func (az *azurestorageFS) InitFS() int {
	Logger.LogDebug("FS : %s InitFS", fsName)
	if err := az.validateAccount(); err != nil {
		Logger.LogErr("Unable to validate account key")
		return -1
	}

	return 0
}

func (az *azurestorageFS) DeInitFs() int {
	Logger.LogDebug("FS : %s DeInitFs", fsName)
	return 0
}

// Set the next component in pipeline for this system
func (az *azurestorageFS) SetClient(cons FSIntf.FileSystem) int {
	panic("not implemented") // TODO: Implement
}

// Get the file system name
func (az *azurestorageFS) GetName() string {
	return fsName
}

// Get the reference count
func (az *azurestorageFS) GetCount() int {
	return az.refCount
}

// Print the pipeline
func (az *azurestorageFS) PrintPipeline() string {
	return (fsName + " -> X ")
}

// Get the file system stats
func (az *azurestorageFS) StatFS() error {
	return nil
}

// Directory level operations
func (az *azurestorageFS) CreateDir(name string, _ os.FileMode) error {
	Logger.LogDebug("FS : CreateDir %s", name)
	return nil
}

func (az *azurestorageFS) DeleteDir(name string) error {
	Logger.LogDebug("FS : DeleteDir %s", name)
	return nil
}

func (az *azurestorageFS) OpenDir(name string) error {
	Logger.LogDebug("FS : OpenDir %s", name)
	return nil
}

func (az *azurestorageFS) CloseDir(name string) error {
	Logger.LogDebug("FS : CloseDir %s", name)
	return nil
}

// ReadDir : Get the list of elements in this directory
func (az *azurestorageFS) ReadDir(name string) (lst []FSIntf.BlobAttr, err error) {
	Logger.LogDebug("FS : ReadDir %s", name)
	lst, err = az.getBlobList(name)
	if err != nil {
		Logger.LogErr("Failed to get list of blobs (%s)", err.Error)
	}
	Logger.LogDebug("Total %d element retreieved)", len(lst))
	return lst, err
}

func (az *azurestorageFS) RenameDir(on string, nn string) error {
	Logger.LogDebug("FS : RenameDir %s -> %s", on, nn)
	return nil
}

// File level operations
func (az *azurestorageFS) CreateFile(name string, _ os.FileMode) error {
	Logger.LogDebug("FS : CreateFile %s", name)
	return nil
}

func (az *azurestorageFS) DeleteFile(name string) error {
	Logger.LogDebug("FS : DeleteFile %s", name)
	return nil
}

func (az *azurestorageFS) OpenFile(name string, _ int, _ os.FileMode) error {
	Logger.LogDebug("FS : OpenFile %s", name)
	return nil
}

func (az *azurestorageFS) CloseFile(name string) error {
	Logger.LogDebug("FS : CloseFile %s", name)
	return nil
}

func (az *azurestorageFS) ReadFile(name string, _ int64, _ int64) (data []byte, err error) {
	Logger.LogDebug("FS : ReadFile %s", name)
	return data, err
}

func (az *azurestorageFS) WriteFile(name string, _ int64, _ int64, _ []byte) (bytes int, err error) {
	Logger.LogDebug("FS : WriteFile %s", name)
	return bytes, err
}

func (az *azurestorageFS) TruncateFile(name string, _ int64) error {
	Logger.LogDebug("FS : TruncateFile %s", name)
	return nil
}

func (az *azurestorageFS) FlushFile(name string) error {
	Logger.LogDebug("FS : FlushFile %s", name)
	return nil
}

func (az *azurestorageFS) ReleaseFile(name string) error {
	Logger.LogDebug("FS : ReleaseFile %s", name)
	return nil
}

func (az *azurestorageFS) UnlinkFile(name string) error {
	Logger.LogDebug("FS : UnlinkFile %s", name)
	return nil
}

// Symlink operations
func (az *azurestorageFS) CreateLink(name string, _ string) error {
	Logger.LogDebug("FS : CreateLink %s", name)
	return nil
}

func (az *azurestorageFS) ReadLink(name string) (data string, err error) {
	Logger.LogDebug("FS : ReadLink %s", name)
	return data, err
}

// Filesystem level operations
func (az *azurestorageFS) GetAttr(name string) (attr FSIntf.BlobAttr, err error) {
	Logger.LogDebug("FS : GetAttr %s", name)
	return attr, err
}

func (az *azurestorageFS) SetAttr(name string, _ FSIntf.BlobAttr) error {
	Logger.LogDebug("FS : SetAttr %s", name)
	return nil
}

func (az *azurestorageFS) Chmod(name string, _ os.FileMode) error {
	Logger.LogDebug("FS : Chmod %s", name)
	return nil
}

func (az *azurestorageFS) Chown(name string, _ string) error {
	Logger.LogDebug("FS : Chown %s", name)
	return nil
}
