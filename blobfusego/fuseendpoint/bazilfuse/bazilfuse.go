package bazilfuse

import (
	FDFact "github.com/blobfusego/fuseendpoint/fusecreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Logger "github.com/blobfusego/global/logger"
	Config "github.com/blobfusego/global"
	
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type bazilFD struct{
	refCount 		int
	consumer		FSIntf.FileSystem
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
		conn = nil 
		fileSys = nil

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
	conn, err := fuse.Mount(
				*config.BlobfuseConfig.MountPath,
				fuse.FSName("blobfuse"),
				fuse.Subtype("blobfuse-go"),
				fuse.LocalVolume(),
				fuse.VolumeName(*Config.BlobfuseConfig.Container),
			)
	if err != nil {
		if err := conn.MountError; err != nil {
			panic(err)
		}
		Logger.LogErr("Failed to mount")
		panic("Failed to mount")
	}

	cfg := &fs.Config()
	fileSys := NewFS()
	
	<-conn.Ready
	if err := conn.MountError; err != nil {
		Logger.LogErr(err)
	}

	Logger.LogDebug(fdName + " Initialized successfully")
}


// Start  : begine the FUSE Listener
func (f *bazilFD) Start() int {
	Logger.LogDebug("Start the FD : " + fdName)
	if err := fs.Serve(filesys); err != nil {
		Logger.LogErr(err)
	}
}


// DeInitFuse : DeInitialize the fuse driver
func (f *bazilFD) DeInitFuse() {
	Logger.LogDebug("Deinit the FD : " + fdName)
	conn.Close()
}

// SetConsumer : Set the next layer that handles the call
func (f *bazilFD) SetConsumer(cons FSIntf.FileSystem) int {
	instance.consumer = cons
	return 0;
}

// GetName : Get the fuse driver name
func (f *bazilFD) GetName() string {
	return fdName
}

// PrintPipeline : Print the current pipeline
func (f *bazilFD) PrintPipeline() string {
	if instance.consumer != nil {
		return (fdName + " -> " + instance.consumer.PrintPipeline())
	} 
	return (fdName + " -> X ")
}

