# Go KIP Uploader by `redClaw`

A simple Go script to automate uploading `.txt` files to a MinIO/S3 bucket.

## Features

- Scans the `kip_files` directory for `.txt` files.
- Uploads one file every 5 minutes to the configured MinIO bucket.
- Moves successfully uploaded files to the `kip_files/success` directory.
- Moves failed uploads to the `kip_files/failed` directory.
- Uses AWS S3/MinIO compatible APIs for file transfer.

## Prerequisites

- Go installed on your machine
- MinIO or AWS S3 credentials
- `.env` file configured properly

## Setup and Configuration

1. **Navigate to the project directory:**
   Ensure you are in the project root directory where `main.go` is located.

2. **Set up the `.env` file:**
   Create a `.env` file in the root of the project with the following structure. **Note:** Keep your actual `.env` values secret and never commit them to version control! (The `.env` file should be ignored in `.gitignore`).

   ```env
   MINIO_ENDPOINT="s3.minio.example.com:9000"
   MINIO_DOWNLOAD_FILE_STORE="files/"
   MINIO_ACCESS_KEY="your-access-key-here"
   MINIO_SECRET_KEY="your-secret-key-here"
   MINIO_BUCKET_NAME="your-bucket-name"
   MINIO_USE_SSL="true"
   MINIO_DURATION_PATH="24h"
   MINIO_IS_TESTING="false"
   MINIO_CRMBE_INTERACTION_PATH="crmbe_interaction/"
   ```

3. **Install Dependencies:**
   Run the following command to download the required Go modules (like `minio-go` and `godotenv`):
   ```bash
   go mod tidy
   ```

## Usage

To start the uploader script, simply run:

```bash
go run main.go
```

The script will:
- Automatically ensure the `kip_files`, `kip_files/success`, and `kip_files/failed` directories exist.
- Run in a continuous loop, checking for new `.txt` files every 5 minutes.
- Upload found files using the credentials in your `.env` file to the specified `MINIO_BUCKET_NAME` under the `MINIO_CRMBE_INTERACTION_PATH` prefix.

## Important Notes

- **Security:** NEVER commit your actual `.env` file with real access keys to version control. Keep the values secret.
- **File Types:** Currently, the script is designed to strictly look for and process `.txt` files.
- **Interval:** The polling interval is hardcoded to wait 5 minutes before attempting the next upload.