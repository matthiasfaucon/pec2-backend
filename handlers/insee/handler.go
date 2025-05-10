package insee

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type EntrepriseInfo struct {
	Siret        string `json:"siret"`
	Company_name string `json:"company_name"`
	Company_type string `json:"company_type"`
	Address      string `json:"address"`
	Postal_code  string `json:"postal_code"`
	City         string `json:"city"`
}

// GetEntrepriseInfo godoc
// @Summary Get Entreprise Info
// @Description Get Entreprise Info
// @Tags insee
// @Param siret path string true "Siret"
// @Success 200 {object} EntrepriseInfo
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /insee/{siret} [get]
func GetEntrepriseInfo(c *gin.Context) {
	siret := c.Param("siret")
	if len(siret) != 14 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SIRET invalide"})
		return
	}

	apiKey := os.Getenv("INSEE_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INSEE_API_KEY manquant"})
		return
	}

	url := "https://api.insee.fr/api-sirene/3.11/siret/" + siret
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création requête"})
		return
	}
	req.Header.Set("X-INSEE-Api-Key-Integration", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur appel INSEE"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "SIRET non trouvé"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur INSEE", "status": resp.StatusCode})
		return
	}

	var apiResp struct {
		Etablissement struct {
			Siret       string `json:"siret"`
			UniteLegale struct {
				DenominationUniteLegale string `json:"denominationUniteLegale"`
				CategorieEntreprise     string `json:"categorieEntreprise"`
			} `json:"uniteLegale"`
			AdresseEtablissement struct {
				NumeroVoieEtablissement        string `json:"numeroVoieEtablissement"`
				TypeVoieEtablissement          string `json:"typeVoieEtablissement"`
				LibelleVoieEtablissement       string `json:"libelleVoieEtablissement"`
				CodePostalEtablissement        string `json:"codePostalEtablissement"`
				LibelleCommuneEtablissement    string `json:"libelleCommuneEtablissement"`
				ComplementAdresseEtablissement string `json:"complementAdresseEtablissement"`
			} `json:"adresseEtablissement"`
		} `json:"etablissement"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur décodage INSEE"})
		return
	}

	etab := apiResp.Etablissement
	adresse := etab.AdresseEtablissement.ComplementAdresseEtablissement
	if adresse != "" {
		adresse += ", "
	}
	adresse += etab.AdresseEtablissement.NumeroVoieEtablissement + " " + etab.AdresseEtablissement.TypeVoieEtablissement + " " + etab.AdresseEtablissement.LibelleVoieEtablissement

	info := EntrepriseInfo{
		Siret:        etab.Siret,
		Company_name: etab.UniteLegale.DenominationUniteLegale,
		Company_type: etab.UniteLegale.CategorieEntreprise,
		Address:      adresse,
		Postal_code:  etab.AdresseEtablissement.CodePostalEtablissement,
		City:         etab.AdresseEtablissement.LibelleCommuneEtablissement,
	}
	c.JSON(http.StatusOK, info)
}
