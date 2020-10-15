package azurestorage

import (
	"fmt"

	"github.com/Azure/azure-storage-blob-go/azblob"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"
)

func setFlags(attr *FSIntf.BlobAttr, blob *azblob.BlobItem) {
	metadata := blob.Metadata
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

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err = az.containerURL.ListBlobsHierarchySegment(az.ctx, marker, "/",
			azblob.ListBlobsSegmentOptions{MaxResults: maxResults,
				Prefix: name[1:],
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
			setFlags(&attr, &blobInfo)
			blobLst = append(blobLst, attr)
			Logger.LogDebug("Added %s to the dir listing", blobInfo.Name)
		}
	}

	return blobLst, err
}
