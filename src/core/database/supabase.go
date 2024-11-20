package database

import (
	"errors"
	"os"

	storage_go "github.com/supabase-community/storage-go"
)

// SupabaseStorage initializes the storage client and bucket name
func SupabaseStorage() (*storage_go.Client, string, error) {
	projectReferenceID := os.Getenv("SUPABASE_URL")
	projectSecretAPIKey := os.Getenv("SUPABASE_KEY")
	bucketName := os.Getenv("BUCKET_NAME")

	if projectReferenceID == "" || projectSecretAPIKey == "" || bucketName == "" {
		return nil, "", errors.New("missing SUPABASE_URL, SUPABASE_KEY, or BUCKET_NAME in environment variables")
	}

	storageClient := storage_go.NewClient(projectReferenceID+"/storage/v1/s3", projectSecretAPIKey, nil)
	return storageClient, bucketName, nil
}
