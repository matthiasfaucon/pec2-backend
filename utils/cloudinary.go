package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

// InitCloudinary initialise la connexion à Cloudinary
func InitCloudinary() error {
	var err error

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return fmt.Errorf("the cloudinary environment variables are not defined")
	}

	cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return fmt.Errorf("erreur lors de l'initialisation de Cloudinary: %v", err)
	}

	// Vérifier la connexion à Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = cld.Admin.Ping(ctx)
	if err != nil {
		return fmt.Errorf("error checking the connection to Cloudinary: %v", err)
	}

	return nil
}

func boolPointer(b bool) *bool {
	return &b
}

// Vérifie si l'extension du fichier est supportée
func isValidImageType(filename string) bool {
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	lowerFilename := strings.ToLower(filename)

	for _, ext := range validExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	return false
}

// UploadProfilePicture télécharge une image de profil vers Cloudinary
func UploadProfilePicture(file *multipart.FileHeader) (string, error) {
	if !isValidImageType(file.Filename) {
		return "", fmt.Errorf("unsupported image format. Use JPG, PNG, GIF, WEBP, BMP or SVG")
	}

	if file.Size > 10*1024*1024 {
		return "", fmt.Errorf("image size too large. Maximum 10MB allowed")
	}

	if cld == nil {
		if err := InitCloudinary(); err != nil {
			return "", err
		}
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening the file: %v", err)
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading the file: %v", err)
	}

	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("error resetting the file cursor: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timestamp := time.Now().Unix()
	publicID := fmt.Sprintf("profile_%d", timestamp)

	uploadParams := uploader.UploadParams{
		Folder:         "profile_pictures",
		PublicID:       publicID,
		UseFilename:    boolPointer(true),
		UniqueFilename: boolPointer(true),
		Overwrite:      boolPointer(true),
		ResourceType:   "auto",
	}

	uploadResult, err := cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return "", fmt.Errorf("erreur lors du téléchargement vers Cloudinary: %v", err)
	}

	if uploadResult.SecureURL == "" {
		if uploadResult.PublicID != "" {
			cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
			constructedURL := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s",
				cloudName, uploadResult.PublicID)
			return constructedURL, nil
		}

			return "", fmt.Errorf("URL sécurisée vide dans la réponse de Cloudinary")
	}

	return uploadResult.SecureURL, nil
}
