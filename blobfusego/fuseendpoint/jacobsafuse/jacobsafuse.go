package jacobsafuse

import (
	"context"
	"sync"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	FDFact "github.com/blobfusego/fuseendpoint/fusecreator"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseutil"
)

// jacobFS : Master struct for global data held in instance
type jacobFS struct {
	rootFD *jacobNode

	refCount int
	client   FSIntf.FileSystem
	server   fuse.Server
	jfs      *fuse.MountedFileSystem
	mu       sync.RWMutex
	nodeID   uint64
}

var instance *jacobFS
var fdName = string("jacobsa")

var regObj = FDFact.FDManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}

func init() {
	FDFact.RegisterFuseDriver(fdName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FDCREATOR TO WPORK

// CreateObj : Create the dummy FS object for factory
func CreateObj() FDFact.FuseDriver {
	if instance == nil {
		instance = &jacobFS{}
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
func (f *jacobFS) InitFuse() {
	Logger.LogDebug("Init the FD : " + fdName)

	instance.rootFD = NewJacobRoot()
	instance.server = fuseutil.NewFileSystemServer(instance.rootFD)

	Logger.LogDebug(fdName + " Initialized successfully")
}

// Start  : begine the FUSE Listener
func (f *jacobFS) Start() int {
	Logger.LogDebug("Start the FD : " + fdName)
	// Init a mount configuration object.
	mountCfg := &fuse.MountConfig{
		FSName:                  "blobfusego",
		DisableWritebackCaching: true,
		ErrorLogger:             Logger.GetLoggerObj(),
		//Options:                 flags.AdditionalMountOptions,
	}

	// Mount the file system.
	var err error
	instance.jfs, err = fuse.Mount(*Config.BlobfuseConfig.MountPath, instance.server, mountCfg)
	if err != nil {
		Logger.LogErr("FD : Failed to mount %s", err.Error())
		return -1
	}

	instance.jfs.Join(context.Background())
	return 0
}

// DeInitFuse : DeInitialize the fuse driver
func (f *jacobFS) DeInitFuse() {
	Logger.LogDebug("Deinit the FD : " + fdName)
}

// SetClient : Set the next layer that handles the call
func (f *jacobFS) SetClient(cons FSIntf.FileSystem) int {
	instance.client = cons
	return 0
}

// GetName : Get the fuse driver name
func (f *jacobFS) GetName() string {
	return fdName
}

// PrintPipeline : Print the current pipeline
func (f *jacobFS) PrintPipeline() string {
	if instance.client != nil {
		return (fdName + " -> " + instance.client.PrintPipeline())
	}
	return (fdName + " -> X ")
}
