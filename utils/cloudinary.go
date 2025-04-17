package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
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
		return fmt.Errorf("Cloudinary configuration is missing")
	}

	cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return fmt.Errorf("error initializing Cloudinary: %v", err)
	}

	return nil
}

func boolPointer(b bool) *bool {
	return &b
}

// UploadProfilePicture télécharge une image de profil vers Cloudinary
func UploadProfilePicture(file *multipart.FileHeader) (string, error) {
	if cld == nil {
		if err := InitCloudinary(); err != nil {
			return "", err
		}
	}

	// Ouvrir le fichier
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer src.Close()

	// Créer un contexte avec un timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Télécharger le fichier vers Cloudinary
	uploadParams := uploader.UploadParams{
		Folder:         "profile_pictures",
		PublicID:       fmt.Sprintf("profile_%d", time.Now().Unix()), // Générer un ID unique
		UseFilename:    boolPointer(true),
		UniqueFilename: boolPointer(true),
		Overwrite:      boolPointer(true),
	}

	uploadResult, err := cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return "", fmt.Errorf("error uploading to Cloudinary: %v", err)
	}

	return uploadResult.SecureURL, nil
}
