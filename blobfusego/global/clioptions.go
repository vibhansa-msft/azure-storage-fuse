
package clioptions

import (
	"fmt"
	"flag"
	"os"
)

// GlobalConfig : Global config for the application
type GlobalConfig struct {
	MountPath		*string			// Mandatory 	: Path to the mounted directory
	TmpPath			*string			// Mandatory 	: Path to the tmp directory

	FSName			*string			// Optional		: FS Name (dummy / loopback) 	
	FDName			*string			// Optional		: FS Name (bazil) 	
}

// BlobfuseConfig : Global config for the application
var BlobfuseConfig GlobalConfig

func init() {
	// Add all your command line options parsing here
	BlobfuseConfig.MountPath 	= flag.String("mount-path", 	".", 			"Path for the mount directory")
	BlobfuseConfig.TmpPath 		= flag.String("tmp-path", 		".", 			"Path for the temp directory") 
	BlobfuseConfig.FSName 		= flag.String("fs", 			"loopback",		"File System to be used") 
	BlobfuseConfig.FDName 		= flag.String("fd", 			"bazil", 		"Fuse Driver to be used") 


	flag.Usage = Usage

	flag.Parse()
}


// Usage : Print the usage of the application
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage : blobfuse <options>\n")
	flag.PrintDefaults()
}

// PrintOptionValues : Print the command line arguments
func PrintOptionValues() {
	fmt.Println("Cli option : Mount path : " + *BlobfuseConfig.MountPath)
	fmt.Println("Cli option : Tmp path : " + *BlobfuseConfig.TmpPath)
	fmt.Println("Cli option : FS Name : " + *BlobfuseConfig.FSName)
	fmt.Println("Cli option : FD Name : " + *BlobfuseConfig.FDName)

}




