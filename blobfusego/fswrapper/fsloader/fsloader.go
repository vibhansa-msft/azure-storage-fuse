package fsloader

import (
	// Just load all the factory object packages so that they register to factory
	_ "../../fsendpoint/loopback"
	_ "../../fsendpoint/dummy"
)
