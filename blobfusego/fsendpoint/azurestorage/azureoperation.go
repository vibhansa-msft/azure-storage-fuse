package azurestorage

import (
	"fmt"
	"syscall"

	"github.com/Azure/azure-storage-blob-go/azblob"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"
)

func setFlags(attr *FSIntf.BlobAttr, metadata azblob.Metadata) {
	for k, v := range metadata {
		fmt.Print(k + "=" + v + "\n")
		if k == "hdi_isfolder" &&
			v == "true" {
			attr.Flags.Set(FSIntf.PropFlagIsDir)
		} else if k == "is_symlink" &&
			v == "true" {
			attr.Flags.Set(FSIntf.PropFlagSymlink)
		}
	}
}

// Refer below link for BlobProperties
// https://github.com/Azure/azure-storage-blob-go/blob/master/azblob/zz_generated_models.go#L4921

func (az *azurestorageFS) getBlobList(name string) (blobLst []FSIntf.BlobAttr, err error) {
	var maxResults int32 = 5000
	var listBlob *azblob.ListBlobsHierarchySegmentResponse

	if name != "" {
		name += "/"
	}

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err = az.containerURL.ListBlobsHierarchySegment(az.ctx, marker, "/",
			azblob.ListBlobsSegmentOptions{MaxResults: maxResults,
				Prefix: name,
				Details: azblob.BlobListingDetails{
					Metadata:  true,
					Deleted:   false,
					Snapshots: false,
				},
			})

		// Store the next marker to next iteration
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			attr := FSIntf.BlobAttr{
				Name:    blobInfo.Name,
				Size:    uint64(*blobInfo.Properties.ContentLength),
				Mode:    Config.BlobfuseConfig.DefaultPerm, // TODO : VB : To be replcaed by blob mode later
				Modtime: blobInfo.Properties.LastModified,
				Flags:   0,
			}
			setFlags(&attr, blobInfo.Metadata)
			blobLst = append(blobLst, attr)
			Logger.LogDebug("Added %s to the dir listing", blobInfo.Name)
		}
	}

	return blobLst, err
}

func (az *azurestorageFS) getBlobAttr(name string) (attr FSIntf.BlobAttr, err error) {
	blobURL := az.containerURL.NewBlockBlobURL(name)
	prop, err := blobURL.GetProperties(az.ctx, azblob.BlobAccessConditions{})

	if err != nil {
		e := StoreErrToErr(err)
		if e == ErrFileNotFound {
			Logger.LogErr("Failed to get properties of object %s (%s)", name, ErrStr[ErrFileNotFound])
			return attr, syscall.ENOENT
		}
	}

	attr = FSIntf.BlobAttr{
		Name:    name,
		Size:    uint64(prop.ContentLength()),
		Mode:    Config.BlobfuseConfig.DefaultPerm, // TODO : VB : To be replcaed by blob mode later
		Modtime: prop.LastModified(),
		Flags:   0,
	}
	setFlags(&attr, prop.NewMetadata())
	return attr, err
}
