package utils

import (
	"context"
	"fmt"
	"io"
	"log"
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

	log.Printf("Cloudinary configuration: Cloud Name: %s, API Key: %s, API Secret: %s...",
		cloudName, apiKey, apiSecret[:5]+"...")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return fmt.Errorf("les variables d'environnement Cloudinary ne sont pas définies")
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
		return fmt.Errorf("erreur lors de la vérification de la connexion à Cloudinary: %v", err)
	}

	log.Println("Cloudinary initialisé avec succès et connexion vérifiée")
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
	log.Printf("Début de l'upload de l'image: %s, taille: %d", file.Filename, file.Size)

	// Vérifier le type d'image
	if !isValidImageType(file.Filename) {
		return "", fmt.Errorf("format d'image non supporté. Utilisez JPG, PNG, GIF, WEBP, BMP ou SVG")
	}

	// Vérifier la taille de l'image (10MB max)
	if file.Size > 10*1024*1024 {
		return "", fmt.Errorf("taille d'image trop grande. Maximum 10MB autorisé")
	}

	if cld == nil {
		log.Println("Cloudinary n'est pas initialisé, tentative d'initialisation...")
		if err := InitCloudinary(); err != nil {
			log.Printf("Échec de l'initialisation de Cloudinary: %v", err)
			return "", err
		}
	}

	// Ouvrir le fichier
	src, err := file.Open()
	if err != nil {
		log.Printf("Erreur lors de l'ouverture du fichier: %v", err)
		return "", fmt.Errorf("erreur lors de l'ouverture du fichier: %v", err)
	}
	defer src.Close()

	// Lire les premiers octets pour vérifier la signature du fichier
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		log.Printf("Erreur lors de la lecture du fichier: %v", err)
		return "", fmt.Errorf("erreur lors de la lecture du fichier: %v", err)
	}

	// Réinitialiser le curseur du fichier
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("Erreur lors de la réinitialisation du curseur du fichier: %v", err)
		return "", fmt.Errorf("erreur lors de la réinitialisation du curseur du fichier: %v", err)
	}

	// Créer un contexte avec un timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Générer un ID unique
	timestamp := time.Now().Unix()
	publicID := fmt.Sprintf("profile_%d", timestamp)

	// Télécharger le fichier vers Cloudinary
	log.Printf("Préparation du téléchargement vers Cloudinary... (Cloud Name: %s)", os.Getenv("CLOUDINARY_CLOUD_NAME"))
	uploadParams := uploader.UploadParams{
		Folder:         "profile_pictures",
		PublicID:       publicID,
		UseFilename:    boolPointer(true),
		UniqueFilename: boolPointer(true),
		Overwrite:      boolPointer(true),
		ResourceType:   "auto", // Utilisez 'auto' au lieu de 'image' pour plus de flexibilité
	}

	log.Printf("Téléchargement vers Cloudinary en cours: dossier=%s, publicID=%s",
		uploadParams.Folder, uploadParams.PublicID)

	uploadResult, err := cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		log.Printf("Erreur lors du téléchargement vers Cloudinary: %v", err)
		return "", fmt.Errorf("erreur lors du téléchargement vers Cloudinary: %v", err)
	}

	if uploadResult.SecureURL == "" {
		log.Printf("URL sécurisée vide dans la réponse de Cloudinary. AssetID: %s, PublicID: %s",
			uploadResult.AssetID, uploadResult.PublicID)

		// Si l'URL est vide mais qu'on a un PublicID, on peut essayer de construire l'URL
		if uploadResult.PublicID != "" {
			cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
			constructedURL := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s",
				cloudName, uploadResult.PublicID)
			log.Printf("Construction manuelle de l'URL: %s", constructedURL)
			return constructedURL, nil
		}

		return "", fmt.Errorf("URL sécurisée vide dans la réponse de Cloudinary")
	}

	log.Printf("Image téléchargée avec succès, URL: %s", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}
