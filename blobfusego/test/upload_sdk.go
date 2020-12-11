package main
import (
        "context"
        "fmt"
        "net/url"
        "os"
		"time"
        "path"
        "github.com/Azure/azure-storage-blob-go/azblob"
)
func main() {
        filepath := os.Args[1]
        accountName := "vikasfuseblob"
        sas := "?sv=2019-12-12&ss=b&srt=sco&sp=rwlacx&se=2021-09-29T14:43:37Z&st=2020-09-29T06:43:37Z&spr=https,http&sig=Mr1TUk3m%2B6l0YmphFsJ6%2BROFr%2BrNzoypsti1gFWsXzk%3D"
        containerName := "testcntgo"
        file, err := os.Open(filepath)
        if err != nil {
                panic(err)
        }
        c := azblob.NewAnonymousCredential()
        p := azblob.NewPipeline(c, azblob.PipelineOptions{})
        cURL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s%s", accountName, containerName, sas))
        containerURL := azblob.NewContainerURL(*cURL, p)

        blobURL := containerURL.NewBlockBlobURL(path.Base(filepath))
		time1 := time.Now()
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
		time2 := time.Now()
		diff := time2.Sub(time1).Seconds()
		fmt.Println("Time taken to Upload ", diff)


		file.Close()
		_ = os.Remove(filepath)
		
		file, err = os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0777)
        if err != nil {
                panic(err)
        }
		time1 = time.Now()
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
		time2 = time.Now()
		diff = time2.Sub(time1).Seconds()
		fmt.Println("Time taken to Download ", diff)
		file.Close()

		

}
