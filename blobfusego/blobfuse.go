
package main

import (
	"fmt"
	"os"

	// As Config initialize the logger this shall always be the first import
	Config 		"./global"

	_			"./fswrapper/fsinterface"
	_ 			"./fswrapper/fsloader"
	_ 			"./fuseendpoint/fuseloader"

	FSFact		"./fswrapper/fscreator"
	FDFact 		"./fuseendpoint/fusecreator"
	Logger		"./global/logger"
)

// Usage and global config are part of 'global' package
// Sample CLI : go run blobfuse.go -mount-path="~/blob_mnt" -tmp-path="/mnt/blobfusetmp" -fs=loopback -fd=bazil -log-level=DEBUG -log-file=blobfuse.log

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


	fd.SetConsumer(nil)

	Logger.LogInfo("Starting to destroy pipeline")

	FDFact.ReleaseFuseDriver(fd)
	FSFact.ReleaseFileSystem(fs)

	Config.StopLogger()
}

