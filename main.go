package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it. Using default env variables or those from system.")
	}

	// MinIO Configuration from env
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9000" // fallback example
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "kip-files"
	}
	useSsl := os.Getenv("MINIO_USE_SSL") == "true"

	sourceDir := "kip_files"
	successDir := sourceDir + "/success"
	failedDir := sourceDir + "/failed"

	// Ensure directories exist
	for _, dir := range []string{sourceDir, successDir, failedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	log.Println("Starting kip uploader script...")

	for {
		processNextFile(endpoint, accessKey, secretKey, bucketName, useSsl, sourceDir, successDir, failedDir)

		log.Println("Waiting for 5 minutes before checking again...")
		time.Sleep(5 * time.Minute)
	}
}

func processNextFile(endpoint, accessKey, secretKey, bucketName string, useSsl bool, sourceDir, successDir, failedDir string) {
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Printf("Error reading source directory: %v\n", err)
		return
	}

	var targetFile os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		log.Println("No .txt files found to process.")
		return
	}

	fileName := targetFile.Name()
	sourcePath := filepath.Join(sourceDir, fileName)
	log.Printf("Processing file: %s\n", fileName)

	ctx := context.Background()

	// Upload file
	err = AWSS3ObjectUpload(ctx, endpoint, accessKey, secretKey, bucketName, fileName, sourcePath, useSsl)

	if err != nil {
		log.Printf("Failed to upload %s: %v\n", fileName, err)
		failedPath := filepath.Join(failedDir, fileName)
		if moveErr := os.Rename(sourcePath, failedPath); moveErr != nil {
			log.Printf("Failed to move file to failed dir: %v\n", moveErr)
		} else {
			log.Printf("Moved %s to failed directory\n", fileName)
		}
	} else {
		log.Printf("Successfully uploaded %s\n", fileName)
		successPath := filepath.Join(successDir, fileName)
		if moveErr := os.Rename(sourcePath, successPath); moveErr != nil {
			log.Printf("Failed to move file to success dir: %v\n", moveErr)
		} else {
			log.Printf("Moved %s to success directory\n", fileName)
		}
	}
}

func AWSS3ObjectDownload(ctx context.Context, endpoint, accessKey, secretKey, bucketName, objectName string, useSsl bool, filename string, duration time.Duration) (*string, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSsl,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		return nil, errors.New("download failed: " + err.Error() + "001")
	}

	_, errGetData := minioClient.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if errGetData != nil {
		return nil, errors.New("download failed: " + errGetData.Error() + "002")
	}

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment;filename="+filename)
	presignedURL, errUrl := minioClient.PresignedGetObject(ctx, bucketName, objectName, duration, reqParams)

	if errUrl != nil {
		return nil, errors.New("download failed: " + errUrl.Error() + "003")
	}

	var u = presignedURL.Scheme + "://" + presignedURL.Host + presignedURL.Path + "?" + presignedURL.RawQuery

	return &u, nil
}

func AWSS3ObjectUpload(ctx context.Context, endpoint, accessKey, secretKey, bucketName, objectName, filePath string, useSsl bool) error {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSsl,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		return errors.New("upload failed: " + err.Error())
	}

	exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
	if errBucketExists == nil && !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return errors.New("failed to create bucket: " + err.Error())
		}
	}

	objectName = os.Getenv("MINIO_CRMBE_INTERACTION_PATH") + objectName

	_, errUpload := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if errUpload != nil {
		return errors.New("upload failed: " + errUpload.Error())
	}

	return nil
}
