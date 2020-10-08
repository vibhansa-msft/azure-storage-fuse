
package main

import (
	"fmt"
	"os"

	_			"./fswrapper/fsinterface"
	_ 			"./fswrapper/fsloader"
	_ 			"./fuseendpoint/fuseloader"

	FSFact		"./fswrapper/fscreator"
	FDFact 		"./fuseendpoint/fusecreator"

	Config 		"./global"
)

// Usage and global config are part of 'global' package

func main() {	
	Config.PrintOptionValues()

	fs, _ := FSFact.GetFileSystem(*Config.BlobfuseConfig.FSName)
	if fs == nil {
		fmt.Println(" >> FS : " + *Config.BlobfuseConfig.FSName + " does not exists in the system")
		os.Exit(1)
	}

	fmt.Println(fs.GetName())
	fmt.Println(fs.GetCount())

	fd, _ := FDFact.GetFuseDriver(*Config.BlobfuseConfig.FDName)
	if fd == nil {
		fmt.Println(" >> FD : " + *Config.BlobfuseConfig.FDName + " does not exists in the system")
		os.Exit(1)
	}
	
	fd.SetConsumer(fs)
	fmt.Println(fd.GetName())
}

