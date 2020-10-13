
package main

import (
	"fmt"
	"os"

	// As Config initialize the logger this shall always be the first import
	Config 		"github.com/blobfusego/global"

	_			"github.com/blobfusego/fswrapper/fsinterface"
	_ 			"github.com/blobfusego/fswrapper/fsloader"
	_ 			"github.com/blobfusego/fuseendpoint/fuseloader"

	FSFact		"github.com/blobfusego/fswrapper/fscreator"
	FDFact 		"github.com/blobfusego/fuseendpoint/fusecreator"
	Logger		"github.com/blobfusego/global/logger"
)

// Usage and global config are part of 'global' package
// Sample CLI : 
// go run blobfuse.go -mount-path="~/blob_mnt" -tmp-path="/mnt/blobfusetmp" -fs=loopback -fd=bazil -log-level=LOG_DEBUG -log-file=blobfuse.log

func main() {	
	Config.PrintOptionValues()

	Logger.LogInfo("Starting to create pipeline")

	fs, _ := FSFact.GetFileSystem(*Config.BlobfuseConfig.FSName)
	if fs == nil {
		fmt.Println(" >> FS : " + *Config.BlobfuseConfig.FSName + " does not exists in the system")
		os.Exit(1)
	}

	fd, _ := FDFact.GetFuseDriver(*Config.BlobfuseConfig.FDName)
	if fd == nil {
		fmt.Println(" >> FD : " + *Config.BlobfuseConfig.FDName + " does not exists in the system")
		os.Exit(1)
	}
	
	fd.SetConsumer(fs)
	fmt.Println("FD Name : " + fd.GetName())
	Logger.LogInfo("PIPELINE : " + fd.PrintPipeline())

	fd.InitFuse()
	if fd.Start() != 0 {
		Logger.LogErr("Failed to start Fuse Driver")
	}

	Logger.LogInfo("Starting to destroy pipeline")
	fd.DeInitFuse()
	
	fd.SetConsumer(nil)
	FDFact.ReleaseFuseDriver(fd)
	FSFact.ReleaseFileSystem(fs)
	Config.StopLogger()
}

