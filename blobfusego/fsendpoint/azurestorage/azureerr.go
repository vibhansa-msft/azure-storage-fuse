package azurestorage

import "github.com/Azure/azure-storage-blob-go/azblob"

// For detailed error list refert ServiceCodeType at below link
// https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#ListBlobsSegmentOptions

// Convert store error to common errors
func StoreErrToErr(err error) uint16 {
	if serr, ok := err.(azblob.StorageError); ok {
		switch serr.ServiceCode() {
		case azblob.ServiceCodeBlobAlreadyExists:
			return ErrFileAlreadyExists
		case azblob.ServiceCodeBlobNotFound:
			return ErrFileNotFound
		default:
			return ErrUnknown
		}
	}
	return ErrNoErr
}

const (
	ErrNoErr uint16 = iota
	ErrUnknown
	ErrFileNotFound
	ErrFileAlreadyExists
)

// ErrStr : Store error to string mapping
var ErrStr = map[uint16]string{
	ErrNoErr:             "No Error found",
	ErrUnknown:           "Unknown store error",
	ErrFileNotFound:      "Blob not found",
	ErrFileAlreadyExists: "Blob already exists",
}
