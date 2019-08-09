package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gofrs/uuid"
)

func main() {
	files, errW := walkDir(".")

	if errW != nil {
		fmt.Println("Error has occured:", errW)
	} else {
		var fFiles []string
		for _, fName := range files {
			if strings.Contains(fName, "jpg") {
				fFiles = append(fFiles, fName)
			}
		}

		m := make(map[string][]byte)

		// Read file contents into memory
		for _, fName := range fFiles {
			fmt.Println("Found file:", fName)
			dat, errR := ReadFile(fName)

			if errR != nil {
				fmt.Println("Error reading file:", fName, "Error:", errR)
			} else {
				fmt.Println("Finished reading bytes for file:", fName)
				m[fName] = dat
			}
		}

		// push file contents from memory to Azure
		for _, fName := range files {
			fmt.Println("Started uploading: ", fName)
			u, errU := UploadBytesToBlob(m[fName])
			if errU != nil {
				fmt.Println("Error during upload: ", errU)
			}

			fmt.Println("Finished uploading to: ", u)
			fmt.Println("==========================================================")
		}
	}
}

func walkDir(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func ReadFile(filePath string) ([]byte, error) {
	dat, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, err
	} else {
		return dat, nil
	}
}

func UploadBytesToBlob(b []byte) (string, error) {
	azrKey, accountName, endPoint, container := GetAccountInfo()
	u, _ := url.Parse(fmt.Sprint(endPoint, container, "/", GetBlobName()))
	credential, errC := azblob.NewSharedKeyCredential(accountName, azrKey)
	if errC != nil {
		return "", errC
	}

	blockBlobUrl := azblob.NewBlockBlobURL(*u, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	ctx := context.Background()
	o := azblob.UploadToBlockBlobOptions{
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{
			ContentType: "image/jpg",
		},
	}

	_, errU := azblob.UploadBufferToBlockBlob(ctx, b, blockBlobUrl, o)
	return blockBlobUrl.String(), errU
}

func GetAccountInfo() (string, string, string, string) {
	azrKey := "your_azure_access_key"
	azrBlobAccountName := "mytechblog"
	azrPrimaryBlobServiceEndpoint := fmt.Sprintf("https://%s.blob.core.windows.net/", azrBlobAccountName)
	azrBlobContainer := "blog-photos"

	return azrKey, azrBlobAccountName, azrPrimaryBlobServiceEndpoint, azrBlobContainer
}

func GetBlobName() string {
	t := time.Now()
	uuid, _ := uuid.NewV4()

	return fmt.Sprintf("%s-%v.jpg", t.Format("20060102"), uuid)
}
