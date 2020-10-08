
package fsinterface

// FileSystem : Master interface for the file system
// All implementations shall register to the factory and use it get the respective object
type FileSystem interface {

	// Init/DeInit the filesystem
	InitFS() int
	DeInitFs() int

	// Set the next component in pipeline for this system
	SetConsumer(cons FileSystem) int
	
	// Get the file system name
	GetName() string

	// Get the reference count
	GetCount() int

	// Get the file system stats
	StatFS() int

	// Directory level operations
	CreateDir	(path string) int
	DeleteDir	(path string)

	OpenDir		(path string) int
	CloseDir	(path string)

	ReadDir		(path string) int
	RenameDir	(path string, name string) int


	// File level operations
	CreateFile	(path string, mode int) int
	DeleteFile	(path string) int

	OpenFile	(path string, mode int) int
	CloseFile	(path string)

	ReadFile	(path string, offset int, length int) int
	WriteFile	(path string, offset int, length int) int

	FlushFile	(path string) int
	ReleaseFile	(path string) int
	UnlinkFile	(path string) int

	// Symlink operations
	CreateLink	(path string, dst string) int
	ReadLink	(path string, link string) int

	// Filesystem level operations
	GetAttr		(path string) int
	SetAttr		(path string) int

	Chmod		(path string, mod int) int
	Chown		(path string, owner string) int
}

