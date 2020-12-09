package cliparser

import (
	"flag"
	"fmt"
	"os"
	"time"

	Logger "github.com/blobfusego/global/logger"
)

// GlobalConfig : Global config for the application
type GlobalConfig struct {
	MountPath *string // Mandatory 	: Path to the mounted directory
	TmpPath   *string // Mandatory 	: Path to the tmp directory

	FSName *string // Optional		: FS Name (dummy / loopback)
	FDName *string // Optional		: FS Name (bazil)

	LogLevel      *string // Optional		: Logging Level
	LogFile       *string // Optional		: Log file name
	LogFileSizeMB *int    // Optional		: Size of each log file at max
	LogFileCount  *int    // Optional		: Number of logs files to be used for rotation

	StoreAccountName   *string // Mandatory : Storage account name
	StoreAccountKey    *string // Optional : Storage account key for
	StoreAuthType      *string // Mandatory : Auth type chosen by customer
	StoreContainerName *string // Mandatory    : Container name to be mounted
	StorageAccountADLS *bool   // Optional : Whether storage account is ADLS or not

	BlockSizeInMB     *int        // Optional : Size of each block in MB
	ParallelismFactor *int        // Optional : Number of parallel upload/download threads in SDK
	AttrTimeOut       *int        // Optional		: Atttibute timeout for fuse caching
	DefaultPerm       os.FileMode // Default permissions for each blob mounted
	MountTime         time.Time
}

// BlobfuseConfig : Global config for the application
var BlobfuseConfig GlobalConfig

func init() {
	// Basic config
	BlobfuseConfig.MountPath = flag.String("mount-path", ".", "Path for the mount directory")
	BlobfuseConfig.TmpPath = flag.String("tmp-path", ".", "Path for the temp directory")

	BlobfuseConfig.FSName = flag.String("fs", "loopback", "File System to be used")
	BlobfuseConfig.FDName = flag.String("fd", "bazil", "Fuse Driver to be used")

	// Logging related config
	BlobfuseConfig.LogLevel = flag.String("log-level", "WARN", "Logging level")
	BlobfuseConfig.LogFile = flag.String("log-file", "blobfuse.log", "Name of the log file")
	BlobfuseConfig.LogFileSizeMB = flag.Int("log-file-size", 100, "Size of each log file in MB")
	BlobfuseConfig.LogFileCount = flag.Int("log-file-count", 10, "Total number of log files")

	// FD related config
	BlobfuseConfig.AttrTimeOut = flag.Int("attr-timeout", 120, "Timeout for attribute caching in fuse")

	// Azure storage account config
	BlobfuseConfig.StoreAccountName = flag.String("account", "", "Azure Storage account name")
	BlobfuseConfig.StoreAccountKey = flag.String("accountkey", "", "Azure Storage account key")
	BlobfuseConfig.StoreAuthType = flag.String("authtype", "", "Azure Storage auth type")
	BlobfuseConfig.StorageAccountADLS = flag.Bool("adls", false, "Storage account if ADLS or not")
	BlobfuseConfig.StoreContainerName = flag.String("container", "tmp", "Name of the container")

	// Azure GoSDK related param
	BlobfuseConfig.BlockSizeInMB = flag.Int("block-size-in-mb", 0, "Size of each block in the storage account")
	BlobfuseConfig.ParallelismFactor = flag.Int("parallelism", 0, "Number of parallel threads per uload or download")

	flag.Usage = Usage

	flag.Parse()
	StartLogger()

	BlobfuseConfig.DefaultPerm = 0777
	BlobfuseConfig.MountTime = time.Now()

	overrideWithEnvOptions()
	validateConfig()

}

// overrideWithEnvOptions : Override the config from file with that in the env variable
func overrideWithEnvOptions() {
	var str string

	str = os.Getenv("AZURE_STORAGE_ACCOUNT")
	if str != "" {
		*BlobfuseConfig.StoreAccountName = str
	}

	str = os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if str != "" {
		*BlobfuseConfig.StoreAccountKey = str
	}
}

// IsAuthTypeAccKey : Check whether given auth type is key or not
func IsAuthTypeAccKey() bool {
	if *BlobfuseConfig.StoreAuthType == "key" {
		return true
	}
	return false
}

// validateConfig : Check whether provided config is valid or not
func validateConfig() {
	if *BlobfuseConfig.FSName == "azurestorage" {
		// Azure Storage is chosen as the FS
		Logger.LogDebug("Validating config for the azure storage")
		if *BlobfuseConfig.StoreAccountName == "" {
			panic("Storage Account Name is missing")
		}
		if *BlobfuseConfig.StoreContainerName == "" {
			panic("Storage Container Name is missing")
		}

		if IsAuthTypeAccKey() {
			Logger.LogDebug("Chosen AuthType : Key")
			if *BlobfuseConfig.StoreAccountKey == "" {
				panic("Storage Account Key is missing")
			}
		} else {
			panic("Invalid AuthType provided")
		}
	}
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

	Logger.LogInfo("Cli option : FS Name : " + *BlobfuseConfig.FSName)
	Logger.LogInfo("Cli option : FD Name : " + *BlobfuseConfig.FDName)

	Logger.LogInfo("Cli option : Account : " + *BlobfuseConfig.StoreAccountName)
	Logger.LogInfo("Cli option : Container : " + *BlobfuseConfig.StoreContainerName)
	Logger.LogInfo("Cli option : AuthType : " + *BlobfuseConfig.StoreAuthType)
	Logger.LogInfo("Cli option : ADLS : %d", *BlobfuseConfig.StorageAccountADLS)

}

// StartLogger : Init and start the logging infra
func StartLogger() {
	var logcfg Logger.LogConfig
	logcfg.LogLevel = *BlobfuseConfig.LogLevel
	logcfg.LogFile = *BlobfuseConfig.LogFile
	logcfg.LogSizeMB = *BlobfuseConfig.LogFileSizeMB
	logcfg.LogFileCount = *BlobfuseConfig.LogFileCount

	Logger.StartLogger(logcfg)
}

// StopLogger : Stop the logging infra
func StopLogger() {
	Logger.LogCrit("System shutting down")
	Logger.StopLogger()
}
