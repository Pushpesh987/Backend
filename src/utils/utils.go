package utils

import (
	"Backend/src/core/database"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	storage_go "github.com/supabase-community/storage-go"
)

// UploadToSupabaseStorage uploads a file to Supabase storage and returns the file's path, public URL, and content type.
func UploadToSupabaseStorage(file *multipart.FileHeader, path string) (string, string, string, error) {
	// Initialize Supabase storage client
	storageClient, bucketName, err := database.SupabaseStorage()
	if err != nil {
		return "", "", "", err
	}

	// Open the file for reading
	fileBody, err := file.Open()
	if err != nil {
		return "", "", "", err
	}
	defer fileBody.Close()

	// Read the file contents
	fileBytes, err := io.ReadAll(fileBody)
	if err != nil {
		return "", "", "", err
	}

	// Reset the file pointer to the beginning
	_, err = fileBody.Seek(0, io.SeekStart)
	if err != nil {
		return "", "", "", err
	}

	// Detect content type based on file contents
	contentType := http.DetectContentType(fileBytes)

	// Define the folder path for storage
	folderPath := fmt.Sprintf("%s", path)

	// Upload the file to Supabase storage
	_, err = storageClient.UploadFile(bucketName, folderPath, fileBody, storage_go.FileOptions{ContentType: &contentType})
	if err != nil {
		return "", "", "", err
	}

	// Get the public URL for the uploaded file
	response := storageClient.GetPublicUrl(bucketName, folderPath)
	fileUrl := response.SignedURL

	return folderPath, fileUrl, contentType, nil
}

// UpdateToSupabaseStorage updates an existing file in Supabase storage.
func UpdateToSupabaseStorage(file *multipart.FileHeader, path string) (string, string, string, error) {
	// Initialize Supabase storage client
	storageClient, bucketName, err := database.SupabaseStorage()
	if err != nil {
		return "", "", "", err
	}

	// Open the file for reading
	fileBody, err := file.Open()
	if err != nil {
		return "", "", "", err
	}
	defer fileBody.Close()

	// Read the file contents
	fileBytes, err := io.ReadAll(fileBody)
	if err != nil {
		return "", "", "", err
	}

	// Reset the file pointer to the beginning
	_, err = fileBody.Seek(0, io.SeekStart)
	if err != nil {
		return "", "", "", err
	}

	// Detect content type based on file contents
	contentType := http.DetectContentType(fileBytes)

	// Define the folder path for storage
	folderPath := fmt.Sprintf("%s", path)

	// Update the existing file in Supabase storage
	_, err = storageClient.UpdateFile(bucketName, folderPath, fileBody, storage_go.FileOptions{ContentType: &contentType})
	if err != nil {
		return "", "", "", err
	}

	// Get the public URL for the updated file
	response := storageClient.GetPublicUrl(bucketName, folderPath)
	fileUrl := response.SignedURL

	return folderPath, fileUrl, contentType, nil
}

// DeleteFromSupabaseStorage deletes a file from Supabase storage given the file path.
func DeleteFromSupabaseStorage(path string) error {
	// Initialize Supabase storage client
	storageClient, bucketName, err := database.SupabaseStorage()
	if err != nil {
		return err
	}

	// Delete the file from Supabase storage
	_, err = storageClient.RemoveFile(bucketName, []string{path})
	if err != nil {
		return err
	}

	return nil
}

// RemoveDuplicates removes duplicate values from a slice of strings.
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, val := range slice {
		if _, ok := seen[val]; !ok {
			seen[val] = true
			result = append(result, val)
		}
	}

	return result
}
