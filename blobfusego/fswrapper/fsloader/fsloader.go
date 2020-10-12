package fsloader

import (
	// Just load all the factory object packages so that they register to factory
	_ "github.com/blobfusego/fsendpoint/loopback"
	_ "github.com/blobfusego/fsendpoint/dummy"
)
