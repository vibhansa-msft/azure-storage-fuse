package gofuse

import (
	"time"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	FDFact "github.com/blobfusego/fuseendpoint/fusecreator"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type gofuseFD struct {
	refCount int

	client FSIntf.FileSystem
	server *fuse.Server

	rootFD fs.InodeEmbedder
	nodeID uint64
}

var instance *gofuseFD
var fdName = string("gofuse")

var regObj = FDFact.FDManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}

func init() {
	FDFact.RegisterFuseDriver(fdName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FDCREATOR TO WPORK

// CreateObj : Create the dummy FS object for factory
func CreateObj() FDFact.FuseDriver {
	if instance == nil {
		instance = &gofuseFD{}
		instance.refCount = 0
		Logger.LogDebug("Created first instances of " + fdName)
	}
	instance.refCount++
	return instance
}

// ReleaseObj : Delete the dummy FS object
func ReleaseObj() {
	instance.refCount--
	if instance.refCount == 0 {
		instance = nil
		Logger.LogDebug("Released all instances of " + fdName)
	}
}

////////////////////////////////////////

// InitFuse : Initialize the fuse driver
func (f *gofuseFD) InitFuse() {
	Logger.LogDebug("Init the FD : " + fdName)
	var err error
	instance.rootFD, err = NewGofuseRoot(*Config.BlobfuseConfig.TmpPath)

	sec := time.Second * 120
	opts := &fs.Options{
		// These options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		AttrTimeout:  &sec,
		EntryTimeout: &sec,
	}
	opts.MountOptions.Options = append(opts.MountOptions.Options, "default_permissions")
	opts.MountOptions.Options = append(opts.MountOptions.Options, "fsname=blobfusego")
	opts.MountOptions.Name = "blobfusego"
	rawFS := fs.NewNodeFS(instance.rootFD, opts)
	instance.server, err = fuse.NewServer(rawFS, *Config.BlobfuseConfig.MountPath, &opts.MountOptions)

	if err != nil {
		Logger.LogErr("FD : Failed to create new server")
	}

	Logger.LogDebug(fdName + " Initialized successfully")
}

// Start  : begine the FUSE Listener
func (f *gofuseFD) Start() int {
	Logger.LogDebug("Start the FD : " + fdName)
	instance.server.Serve()
	return 0
}

// DeInitFuse : DeInitialize the fuse driver
func (f *gofuseFD) DeInitFuse() {
	Logger.LogDebug("Deinit the FD : " + fdName)
	instance.server.Wait()
}

// SetClient : Set the next layer that handles the call
func (f *gofuseFD) SetClient(cons FSIntf.FileSystem) int {
	instance.client = cons
	return 0
}

// GetName : Get the fuse driver name
func (f *gofuseFD) GetName() string {
	return fdName
}

// PrintPipeline : Print the current pipeline
func (f *gofuseFD) PrintPipeline() string {
	if instance.client != nil {
		return (fdName + " -> " + instance.client.PrintPipeline())
	}
	return (fdName + " -> X ")
}
