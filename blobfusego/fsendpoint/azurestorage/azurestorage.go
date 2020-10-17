package azurestorage

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/Azure/azure-storage-blob-go/azblob"
	FSFact "github.com/blobfusego/fswrapper/fscreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
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
var writeFiles = make(map[string]bool)

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

	blobURL := az.containerURL.NewBlockBlobURL(name)
	metadata := azblob.Metadata{}

	o := azblob.UploadToBlockBlobOptions{
		Metadata:    metadata,
		Parallelism: 10,
		BlockSize:   0,
	}

	data := make([]byte, 0)
	_, err := azblob.UploadBufferToBlockBlob(az.ctx, data, blobURL, o)
	if err != nil {
		Logger.LogErr("Falied to write buffer to blob")
		return syscall.ENOENT
	}

	return nil
}

func (az *azurestorageFS) DeleteFile(name string) error {
	Logger.LogDebug("FS : DeleteFile %s", name)

	blobURL := az.containerURL.NewBlobURL(name)
	_, err := blobURL.Delete(az.ctx, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	if err != nil {
		Logger.LogErr("Failed to delete the file %s (%s)", name, err.Error())
		return err
	}

	return nil
}

func (az *azurestorageFS) OpenFile(name string, flag int, mode os.FileMode) error {
	Logger.LogDebug("FS : OpenFile %s", name)

	if true {
		f, err := os.OpenFile(*Config.BlobfuseConfig.TmpPath+"/"+name,
			os.O_RDWR|os.O_APPEND|os.O_CREATE,
			Config.BlobfuseConfig.DefaultPerm)
		if err != nil {
			Logger.LogErr("Failed to open local file")
			return err
		}

		blobURL := az.containerURL.NewBlobURL(name)

		Logger.LogErr("Going for file download %s", name)
		if true {
			err = azblob.DownloadBlobToFile(az.ctx, blobURL, 0, 0, f, azblob.DownloadFromBlobOptions{})
			if err != nil {
				Logger.LogErr("Download to file failed for %s (%s)", name, err.Error())
				return err
			}
			size, _ := f.Seek(0, io.SeekEnd)
			Logger.LogErr("Download complete of %s, %d bytes read", name, size)
		} else {
			resp, err := blobURL.Download(az.ctx, 0, 0, azblob.BlobAccessConditions{}, false)
			if err != nil {
				Logger.LogErr("Download to file failed for %s (%s)", name, err.Error())
				return err
			}
			Logger.LogErr("Download complete %s", name)

			data, err := ioutil.ReadAll(resp.Response().Body)
			if err != nil {
				Logger.LogErr("Failed to read data from resp")
				return err
			}
			Logger.LogErr("All data read %s", name)

			_, err = f.Write(data)
			if err != nil {
				Logger.LogErr("Failed to save data to file")
				return err
			}
			size, _ := f.Seek(0, io.SeekCurrent)
			Logger.LogErr("File Written %s with %d bytes", name, size)
			resp.Body(azblob.RetryReaderOptions{}).Close()
		}
		f.Close()
	}

	return nil
}

func (az *azurestorageFS) CloseFile(name string) (err error) {
	Logger.LogDebug("FS : CloseFile %s", name)

	if writeFiles[name] == true {
		// File was written so upload the file now
		err = az.FlushFile(name)
		writeFiles[name] = false
	}
	os.Remove(*Config.BlobfuseConfig.TmpPath + "/" + name)
	return err
}

func (az *azurestorageFS) ReadFile(name string, offset int64, len int64) (data []byte, err error) {
	Logger.LogDebug("FS : ReadFile %s (%d : %d)", name, offset, len)

	data = make([]byte, len)

	f, err := os.OpenFile(*Config.BlobfuseConfig.TmpPath+"/"+name,
		os.O_RDONLY,
		Config.BlobfuseConfig.DefaultPerm)
	if err == nil {
		if len == 0 {
			// We need to read till the end of the file
			_, _ = f.Seek(offset, io.SeekStart)
			endpos, _ := f.Seek(0, io.SeekEnd)
			len = (endpos - offset)
			data = make([]byte, len)
		}
		n, err := f.ReadAt(data, offset)
		if err != nil && err != io.EOF {
			Logger.LogErr("Failed to read specified bytes form file")
			return data, err
		}
		f.Close()
		data = data[:n]
		return data, nil
	}

	blobURL := az.containerURL.NewBlobURL(name)
	o := azblob.DownloadFromBlobOptions{
		Parallelism: 10,
		BlockSize:   0,
	}

	err = azblob.DownloadBlobToBuffer(az.ctx, blobURL, offset, len, data, o)
	if err != nil {
		Logger.LogErr("Failed to download the file")
	}
	return data, err
}

func (az *azurestorageFS) WriteFile(name string, offset int64, len int64, data []byte) (bytes int, err error) {
	Logger.LogDebug("FS : WriteFile %s (%d : %d)", name, offset, len)

	f, err := os.OpenFile(*Config.BlobfuseConfig.TmpPath+"/"+name,
		os.O_RDWR|os.O_CREATE,
		Config.BlobfuseConfig.DefaultPerm)

	if err == nil {
		n, err := f.WriteAt(data, offset)
		if err != nil && err != io.EOF {
			Logger.LogErr("Failed to read specified bytes form file")
			return 0, err
		}
		f.Close()
		writeFiles[name] = true
		return n, nil
	}

	return 0, err
}

func (az *azurestorageFS) FlushFile(name string) error {
	Logger.LogDebug("FS : FlushFile %s", name)

	f, err := os.OpenFile(*Config.BlobfuseConfig.TmpPath+"/"+name,
		os.O_RDONLY,
		Config.BlobfuseConfig.DefaultPerm)
	err = az.CopyFromFile(name, f)
	f.Close()

	return err
}

func (az *azurestorageFS) TruncateFile(name string, len int64) error {
	Logger.LogDebug("FS : TruncateFile %s", name)

	// Read len bytes from file
	blobURL := az.containerURL.NewBlobURL(name)
	i := azblob.DownloadFromBlobOptions{
		Parallelism: 10,
		BlockSize:   0,
	}

	data := make([]byte, len)
	err := azblob.DownloadBlobToBuffer(az.ctx, blobURL, 0, len, data, i)
	if err != nil {
		Logger.LogErr("Failed to download the file")
	}

	// Overwrite the file with just n bytes to truncate it
	metadata := azblob.Metadata{}
	o := azblob.UploadToBlockBlobOptions{
		Metadata:    metadata,
		Parallelism: 10,
		BlockSize:   0,
	}

	upblobURL := az.containerURL.NewBlockBlobURL(name)
	_, err = azblob.UploadBufferToBlockBlob(az.ctx, data, upblobURL, o)
	if err != nil {
		Logger.LogErr("Falied to write buffer to blob")
		return syscall.ENOENT
	}

	return nil
}

func (az *azurestorageFS) CopyToFile(name string, f *os.File) (err error) {
	Logger.LogDebug("FS : CopyToFile %s", name)

	blobURL := az.containerURL.NewBlobURL(name)

	Logger.LogErr("Going for file download %s", name)
	err = azblob.DownloadBlobToFile(az.ctx, blobURL, 0, 0, f, azblob.DownloadFromBlobOptions{})
	if err != nil {
		Logger.LogErr("Download to file failed for %s (%s)", name, err.Error())
		return err
	}
	size, _ := f.Seek(0, io.SeekCurrent)
	Logger.LogErr("Download complete of %s, %d bytes read", name, size)

	return nil
}

func (az *azurestorageFS) CopyFromFile(name string, f *os.File) (err error) {
	Logger.LogDebug("FS : CopyFromFile %s", name)

	blobURL := az.containerURL.NewBlockBlobURL(name)

	Logger.LogErr("Going for upload of %s", name)
	_, err = azblob.UploadFileToBlockBlob(az.ctx, f, blobURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		Logger.LogErr("Upload from file failed for %s (%s)", name, err.Error())
		return err
	}
	size, _ := f.Seek(0, io.SeekCurrent)
	Logger.LogErr("Upload complete of %s, %d bytes read", name, size)

	return nil
}

func (az *azurestorageFS) ReleaseFile(name string) error {
	Logger.LogDebug("FS : ReleaseFile %s", name)
	return az.FlushFile(name)
}

// Filesystem level operations
func (az *azurestorageFS) GetAttr(name string) (attr FSIntf.BlobAttr, err error) {
	Logger.LogDebug("FS : GetAttr %s", name)
	attr, err = az.getBlobAttr(name)
	if err != nil {
		Logger.LogErr("Failed to get list of blobs (%s)", err.Error)
	}
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
