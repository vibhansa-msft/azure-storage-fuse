package bazilfuse

import (
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	FDFact "github.com/blobfusego/fuseendpoint/fusecreator"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type bazilFD struct {
	refCount int
	client   FSIntf.FileSystem
}

var instance *bazilFD
var fdName = string("bazil")

var regObj = FDFact.FDManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}

func init() {
	FDFact.RegisterFuseDriver(fdName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FDCREATOR TO WPORK

// CreateObj : Create the dummy FS object for factory
func CreateObj() FDFact.FuseDriver {
	if instance == nil {
		bazilConn = nil
		BazilFS = nil

		instance = &bazilFD{}
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
func (f *bazilFD) InitFuse() {
	Logger.LogDebug("Init the FD : " + fdName)

	var err error
	bazilConn, err = fuse.Mount(
		*Config.BlobfuseConfig.MountPath,
		fuse.FSName("blobfuse"),
		fuse.Subtype("azure"),
	)

	if err != nil {
		Logger.LogErr("Failed to mount : %v", err)

		if err := bazilConn.MountError; err != nil {
			Logger.LogErr("Failed to mount MntErr: %v", bazilConn.MountError)
		}
		panic("Failed to mount")
	}

	BazilFS = NewFS()

	Logger.LogDebug(fdName + " Initialized successfully")
}

// Start  : begine the FUSE Listener
func (f *bazilFD) Start() int {
	Logger.LogDebug("Start the FD : " + fdName)

	if BazilFS == nil {
		Logger.LogErr("FD : Failed to start the fuse driver : fs is null")
		return -1
	}

	if bazilConn == nil {
		Logger.LogErr("FD : Failed to start the fuse driver : connection is null")
		return -1
	}

	if err := fs.Serve(bazilConn, BazilFS); err != nil {
		Logger.LogErr("FD : Failed to start the fuse driver : %v", err)
		return -1
	}

	<-bazilConn.Ready
	return 0
}

// DeInitFuse : DeInitialize the fuse driver
func (f *bazilFD) DeInitFuse() {
	Logger.LogDebug("Deinit the FD : " + fdName)
	bazilConn.Close()
}

// SetClient : Set the next layer that handles the call
func (f *bazilFD) SetClient(cons FSIntf.FileSystem) int {
	instance.client = cons
	BazilFS.client = cons
	return 0
}

// GetName : Get the fuse driver name
func (f *bazilFD) GetName() string {
	return fdName
}

// PrintPipeline : Print the current pipeline
func (f *bazilFD) PrintPipeline() string {
	if instance.client != nil {
		return (fdName + " -> " + instance.client.PrintPipeline())
	}
	return (fdName + " -> X ")
}
