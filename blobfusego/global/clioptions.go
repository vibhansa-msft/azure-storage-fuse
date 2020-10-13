
package cliparser

import (
	"fmt"
	"flag"
	"os"
	Logger 	"github.com/blobfusego/global/logger"
)

// GlobalConfig : Global config for the application
type GlobalConfig struct {
	MountPath		*string			// Mandatory 	: Path to the mounted directory
	TmpPath			*string			// Mandatory 	: Path to the tmp directory

	Container		*string			// Mandatory    : Container name to be mounted

	FSName			*string			// Optional		: FS Name (dummy / loopback) 	
	FDName			*string			// Optional		: FS Name (bazil) 
	
	LogLevel		*string			// Optional		: Logging Level
	LogFile			*string			// Optional		: Log file name
	LogFileSizeMB	*int			// Optional		: Size of each log file at max
	LogFileCount	*int			// Optional		: Number of logs files to be used for rotation


	DefaultPerm		os.FileMode		// Default permissions for each blob mounted 
}

// BlobfuseConfig : Global config for the application
var BlobfuseConfig GlobalConfig

func init() {
	// Basic config
	BlobfuseConfig.MountPath 	= flag.String("mount-path", 	".", 			"Path for the mount directory")
	BlobfuseConfig.TmpPath 		= flag.String("tmp-path", 		".", 			"Path for the temp directory") 
	BlobfuseConfig.Container	= flag.String("container", 		"tmp", 			"Name of the container") 

	BlobfuseConfig.FSName 		= flag.String("fs", 			"loopback",		"File System to be used") 
	BlobfuseConfig.FDName 		= flag.String("fd", 			"bazil", 		"Fuse Driver to be used") 

	// Logging related config
	BlobfuseConfig.LogLevel 			= flag.String("log-level", 		"WARN", "Logging level")
	BlobfuseConfig.LogFile 				= flag.String("log-file", 		"blobfuse.log", "Name of the log file")
	BlobfuseConfig.LogFileSizeMB 		= flag.Int("log-file-size", 	100, "Size of each log file in MB")
	BlobfuseConfig.LogFileCount 		= flag.Int("log-file-count", 	10, "Total number of log files")

	flag.Usage = Usage

	flag.Parse()
	StartLogger()

	BlobfuseConfig.DefaultPerm = 0777
}


// Usage : Print the usage of the application
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage : blobfuse <options>\n")
	flag.PrintDefaults()
}

// PrintOptionValues : Print the command line arguments
func PrintOptionValues() {
	Logger.LogInfo("Cli option : Mount path : " + *BlobfuseConfig.MountPath)
	Logger.LogInfo("Cli option : Tmp path : " + *BlobfuseConfig.TmpPath)
	Logger.LogInfo("Cli option : Container : " + *BlobfuseConfig.Container)
	Logger.LogInfo("Cli option : FS Name : " + *BlobfuseConfig.FSName)
	Logger.LogInfo("Cli option : FD Name : " + *BlobfuseConfig.FDName)
}


// StartLogger : Init and start the logging infra
func StartLogger(){	
	var logcfg Logger.LogConfig
	logcfg.LogLevel 		= *BlobfuseConfig.LogLevel
	logcfg.LogFile 			= *BlobfuseConfig.LogFile
	logcfg.LogSizeMB 		= *BlobfuseConfig.LogFileSizeMB
	logcfg.LogFileCount		= *BlobfuseConfig.LogFileCount

	Logger.StartLogger(logcfg)
}

// StopLogger : Stop the logging infra
func StopLogger() {
	Logger.LogCrit("System shutting down")
	Logger.StopLogger()
}



