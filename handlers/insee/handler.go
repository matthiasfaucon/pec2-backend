package insee

import (
	"encoding/json"
	"net/http"
	"os"
	"pec2-backend/utils"

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
		utils.LogError(nil, "SIRET invalid in GetEntrepriseInfo")
		c.JSON(http.StatusBadRequest, gin.H{"error": "SIRET invalid"})
		return
	}

	apiKey := os.Getenv("INSEE_API_KEY")
	if apiKey == "" {
		utils.LogError(nil, "INSEE_API_KEY missing in GetEntrepriseInfo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INSEE_API_KEY missing"})
		return
	}

	url := "https://api.insee.fr/api-sirene/3.11/siret/" + siret
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.LogError(err, "Error creating request in GetEntrepriseInfo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}
	req.Header.Set("X-INSEE-Api-Key-Integration", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogError(err, "Error calling INSEE in GetEntrepriseInfo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calling INSEE"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		utils.LogError(nil, "Siret not found in GetEntrepriseInfo")
		c.JSON(http.StatusNotFound, gin.H{"error": "Siret not found"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		utils.LogError(nil, "INSEE error (status != 200) in GetEntrepriseInfo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INSEE error", "status": resp.StatusCode})
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
		utils.LogError(err, "Error decoding INSEE in GetEntrepriseInfo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding INSEE"})
		return
	}

	etab := apiResp.Etablissement
	adresse := etab.AdresseEtablissement.NumeroVoieEtablissement + " " + etab.AdresseEtablissement.TypeVoieEtablissement + " " + etab.AdresseEtablissement.LibelleVoieEtablissement

	info := EntrepriseInfo{
		Siret:        etab.Siret,
		Company_name: etab.UniteLegale.DenominationUniteLegale,
		Company_type: etab.UniteLegale.CategorieEntreprise,
		Address:      adresse,
		Postal_code:  etab.AdresseEtablissement.CodePostalEtablissement,
		City:         etab.AdresseEtablissement.LibelleCommuneEtablissement,
	}
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "INSEE informations retrieved successfully in GetEntrepriseInfo")
	c.JSON(http.StatusOK, info)
}
