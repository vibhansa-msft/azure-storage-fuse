package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"
	"io"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

func main() {
	filepath := os.Args[1]
	accountName := "vikasfuseblob"
	sas := "?sv=2019-12-12&ss=b&srt=sco&sp=rwlacx&se=2021-09-29T14:43:37Z&st=2020-09-29T06:43:37Z&spr=https,http&sig=Mr1TUk3m%2B6l0YmphFsJ6%2BROFr%2BrNzoypsti1gFWsXzk%3D"
	containerName := "testcntgo"

	c := azblob.NewAnonymousCredential()
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	cURL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s%s", accountName, containerName, sas))
	containerURL := azblob.NewContainerURL(*cURL, p)

	for i := 1; i < 4; i++ {
		// Generate the url
		blobname := fmt.Sprintf("%s%d", filepath, i)
		filename := fmt.Sprintf("%s%s", "/mnt/ramdisk/", blobname)
		blobURL := containerURL.NewBlockBlobURL(path.Base(blobname))

		fmt.Println("----------------------------------------------------------------------")
		fmt.Println("Next test file ", filename)
		// Download the file
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		fmt.Println("download : ", filename)
		time1 := time.Now()
		err = azblob.DownloadBlobToFile(
			context.Background(),
			blobURL.BlobURL,
			0, 0,
			file,
			azblob.DownloadFromBlobOptions{
				BlockSize:   8 * 1024 * 1024,
				Parallelism: 64,
			})
		if err != nil {
			fmt.Println(err.Error())
		}
		time2 := time.Now()
		size, _ := file.Seek(0, io.SeekEnd)

		fmt.Println("download done : ", filename, " size : ", size)

		diff := time2.Sub(time1).Seconds()
		fmt.Println("Time taken to Download ", filename, "is ", diff, " Seconds")
		file.Close()

		fmt.Println("----------------------------------------------------------------------")
		// Upload the file
		file, err = os.Open(filename)
		if err != nil {
			panic(err)
		}
		fmt.Println("upload : ", filename)

		time1 = time.Now()
		_, err = azblob.UploadFileToBlockBlob(
			context.Background(),
			file,
			blobURL,
			azblob.UploadToBlockBlobOptions{
				BlockSize:   8 * 1024 * 1024,
				Parallelism: 64,
			})
		if err != nil {
			fmt.Println(err.Error())
		}

		time2 = time.Now()
		fmt.Println("upload done : ", filename)

		diff = time2.Sub(time1).Seconds()
		fmt.Println("Time taken to Upload ", filename, "is ", diff, " Seconds")
		file.Close()
		
		_ = os.Remove(filename)
	}

}
