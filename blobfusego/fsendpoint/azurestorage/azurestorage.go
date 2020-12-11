package azurestorage

import (
	"context"
	"io"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	FSFact "github.com/blobfusego/fswrapper/fscreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"
	Stats "github.com/blobfusego/global/perfmonitor"
)

// Example for azblob usage : https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#pkg-examples
// For methods help refer : https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#ContainerURL

type azurestorageFS struct {
	refCount     int
	epURL        *url.URL
	azPipeline   pipeline.Pipeline
	serviceURL   azblob.ServiceURL
	containerURL azblob.ContainerURL
}

var instance *azurestorageFS
var fsName = string("azurestorage")

var writeFileMapLock = sync.RWMutex{}
var openFileMapLock = sync.RWMutex{}
var writeFiles = make(map[string]bool)
var openFiles = make(map[string]*os.File)

func SetWriteFile(name string, flag bool) {
	writeFileMapLock.Lock()
	defer writeFileMapLock.Unlock()
	writeFiles[name] = flag
}

func GetWriteFile(name string) (bool, bool) {
	writeFileMapLock.RLock()
	defer writeFileMapLock.RUnlock()

	e, f := writeFiles[name]
	return e, f
}

func SetOpenFile(name string, f *os.File) {
	openFileMapLock.Lock()
	defer openFileMapLock.Unlock()
	openFiles[name] = f
}

func GetOpenFile(name string) (*os.File, bool) {
	openFileMapLock.RLock()
	defer openFileMapLock.RUnlock()

	e, f := openFiles[name]
	return e, f
}

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

////// FILE OPERATIONS

// File level operations
func (az *azurestorageFS) CreateFile(name string, _ os.FileMode) error {
	Logger.LogDebug("FS : CreateFile %s", name)

	f, err := os.Create(*Config.BlobfuseConfig.TmpPath + "/" + name)
	if err != nil {
		Logger.LogErr("Failed to create file %s", name)
		return err
	}
	f.Close()

	return az.CopyFromFile(name, nil)
}

func (az *azurestorageFS) DeleteFile(name string) error {
	Logger.LogDebug("FS : DeleteFile %s", name)

	blobURL := az.containerURL.NewBlobURL(name)
	_, err := blobURL.Delete(context.Background(), azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	if err != nil {
		Logger.LogErr("Failed to delete the file %s (%s)", name, err.Error())
		return err
	}

	return nil
}

func (az *azurestorageFS) OpenFile(name string, flag int, mode os.FileMode) error {
	Logger.LogDebug("FS : OpenFile %s", name)
	Stats.OpenRequest(fsName)

	SetOpenFile(name, nil)
	return az.CopyToFile(name, nil)
}

func (az *azurestorageFS) CloseFile(name string) (err error) {
	Logger.LogDebug("FS : CloseFile %s", name)

	if flag, f := GetWriteFile(name); f && flag {
		// File was written so upload the file now
		err = az.FlushFile(name)
		SetWriteFile(name, false)
	}

	if e, f := GetOpenFile(name); f && e != nil {
		e.Close()
		SetOpenFile(name, nil)
	}
	return err
}

func (az *azurestorageFS) ReadFile(name string, offset int64, size int64) (data []byte, err error) {
	//Logger.LogDebug("FS : ReadFile %s (%d : %d)", name, offset, size)
	Stats.ReadRequest(fsName)
	data = make([]byte, size)
	var f *os.File
	var found bool

	if f, found = GetOpenFile(name); !found || f == nil {
		f, err = os.OpenFile(*Config.BlobfuseConfig.TmpPath+"/"+name,
			os.O_RDONLY,
			Config.BlobfuseConfig.DefaultPerm)
		SetOpenFile(name, f)
	}

	if size == 0 {
		// We need to read till the end of the file
		endpos, _ := f.Seek(0, io.SeekEnd)
		size = (endpos - offset)
		data = make([]byte, size)
	}

	n, err := f.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		Logger.LogErr("Failed to read specified bytes form file")
		Stats.ReadRequestFail(fsName)
		return data, err
	}

	data = data[:n]
	Stats.ReadBytes(fsName, uint64(n))
	return data, nil

}

func (az *azurestorageFS) WriteFile(name string, offset int64, size int64, data []byte) (bytes int, err error) {
	//Logger.LogDebug("FS : WriteFile %s (%d : %d)", name, offset, size)
	var f *os.File
	var found bool
	if f, found = GetOpenFile(name); !found || f == nil {
		f, err = os.OpenFile(*Config.BlobfuseConfig.TmpPath+"/"+name,
			os.O_RDWR|os.O_CREATE,
			Config.BlobfuseConfig.DefaultPerm)
		SetOpenFile(name, f)
	}

	if err == nil {
		n, err := f.WriteAt(data, offset)
		if err != nil && err != io.EOF {
			Logger.LogErr("Failed to read specified bytes form file")
			return 0, err
		}
		SetWriteFile(name, true)
		f.Sync()

		return n, nil
	}

	return 0, err
}

func (az *azurestorageFS) FlushFile(name string) (err error) {
	Logger.LogDebug("FS : FlushFile %s", name)

	if write, found := GetWriteFile(name); found && write {
		return az.CopyFromFile(name, nil)
		/*
			f, err := os.Open(*Config.BlobfuseConfig.TmpPath + "/" + name)
			err = az.CopyFromFile(name, f)
			if err != nil {
				Logger.LogErr("Failed to upload the file %s (%s)", name, err.Error())
			}
			f.Close()
		*/
	}

	return err
}

func (az *azurestorageFS) TruncateFile(name string, size int64) error {
	Logger.LogDebug("FS : TruncateFile %s", name)
	return az.CopyFromFile(name, nil)
}

func (az *azurestorageFS) CopyToFile(name string, fi *os.File) (err error) {
	Logger.LogDebug("FS : CopyToFile %s", name)

	f, err := os.Create(*Config.BlobfuseConfig.TmpPath + "/" + name)
	defer f.Close()
	blobURL := az.containerURL.NewBlockBlobURL(name)

	downopt := azblob.DownloadFromBlobOptions{}
	if (*Config.BlobfuseConfig.BlockSizeInMB) != 0 {
		downopt.BlockSize = (int64(*Config.BlobfuseConfig.BlockSizeInMB) * 1024 * 1024)
		downopt.Parallelism = uint16(*Config.BlobfuseConfig.ParallelismFactor)
	}

	Logger.LogErr("Going for file download %s", name)

	time1 := time.Now()
	err = azblob.DownloadBlobToFile(context.Background(), blobURL.BlobURL, 0, 0, f, azblob.DownloadFromBlobOptions{
		BlockSize:   8 * 1024 * 1024,
		Parallelism: 64,
	})
	if err != nil {
		Logger.LogErr("Download to file failed for %s (%s)", name, err.Error())
		return err
	}
	time2 := time.Now()
	size, _ := f.Seek(0, io.SeekEnd)
	Logger.LogErr("Download complete of %s, %d bytes read", name, size)

	diff := time2.Sub(time1).Seconds()
	Logger.LogErr("** Download %s done in %d seconds", name, diff)

	return nil
}

func (az *azurestorageFS) CopyFromFile(name string, f *os.File) (err error) {
	Logger.LogDebug("FS : CopyFromFile %s", name)

	if f == nil {
		f, err = os.Open(*Config.BlobfuseConfig.TmpPath + "/" + name)
		if err != nil {
			Logger.LogErr("Fail to open file %s", name)
			return err
		}
	}
	defer f.Close()

	blobURL := az.containerURL.NewBlockBlobURL(path.Base(name))
	upopt := azblob.UploadToBlockBlobOptions{}
	if (*Config.BlobfuseConfig.BlockSizeInMB) != 0 {
		upopt.BlockSize = (int64(*Config.BlobfuseConfig.BlockSizeInMB) * 1024 * 1024)
		upopt.Parallelism = uint16(*Config.BlobfuseConfig.ParallelismFactor)
	}

	Logger.LogErr("Going for upload of %s", name)
	time1 := time.Now()
	_, err = azblob.UploadFileToBlockBlob(context.Background(), f, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   8 * 1024 * 1024,
		Parallelism: 64,
	})

	if err != nil {
		Logger.LogErr("Upload from file failed for %s (%s)", name, err.Error())
		return err
	}
	time2 := time.Now()
	size, _ := f.Seek(0, io.SeekEnd)
	Logger.LogErr("Upload complete of %s, %d bytes read", name, size)

	diff := time2.Sub(time1).Seconds()
	Logger.LogErr("** Upload %s done in %d seconds", name, diff)

	return nil
}

func (az *azurestorageFS) ReleaseFile(name string) error {
	Logger.LogDebug("FS : ReleaseFile %s", name)
	//return az.FlushFile(name)
	return nil
}

// Filesystem level operations
func (az *azurestorageFS) GetAttr(name string) (attr FSIntf.BlobAttr, err error) {
	//Logger.LogDebug("FS : GetAttr %s", name)
	attr, err = az.getBlobAttr(name)
	//if err != nil {
	//	Logger.LogErr("Failed to get list of blobs (%s)", err.Error)
	//}
	return attr, err
}

func (az *azurestorageFS) SetAttr(name string, _ FSIntf.BlobAttr) error {
	Logger.LogDebug("FS : SetAttr %s", name)
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

func (az *azurestorageFS) Chmod(name string, _ os.FileMode) error {
	Logger.LogDebug("FS : Chmod %s", name)
	return nil
}

func (az *azurestorageFS) Chown(name string, _ string) error {
	Logger.LogDebug("FS : Chown %s", name)
	return nil
}
