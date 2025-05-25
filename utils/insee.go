package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

type InseeApiResponse struct {
	Header struct {
		Statut  int    `json:"statut"`
		Message string `json:"message"`
	} `json:"header"`
	Etablissement *struct {
		Siret                        string `json:"siret"`
		StatutDiffusionEtablissement string `json:"statutDiffusionEtablissement"`
		EtablissementSiege           bool   `json:"etablissementSiege"`
	} `json:"etablissement"`
}

var (
	inseeApiKeyEnv = "INSEE_API_KEY"
	inseeSiretURL  = "https://api.insee.fr/api-sirene/3.11/siret/"
	// inseeTokenURL  = "https://api.insee.fr/token"
	// tokenExpiresIn = 7 * 24 * time.Hour // 7 days
)

func VerifySiret(siret string) (bool, error) {
	// Validate SIRET format (14 digits)
	matched, err := regexp.MatchString(`^\d{14}$`, siret)
	if err != nil {
		return false, fmt.Errorf("error validating SIRET format: %v", err)
	}
	if !matched {
		return false, nil
	}

	apiKey := os.Getenv(inseeApiKeyEnv)
	if apiKey == "" {
		return false, fmt.Errorf("INSEE_API_KEY is required in environment variables")
	}

	url := inseeSiretURL + siret
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("X-INSEE-Api-Key-Integration", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("INSEE API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var apiResp InseeApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return false, fmt.Errorf("error decoding response: %v", err)
	}

	return apiResp.Header.Statut == 200 &&
		apiResp.Etablissement != nil &&
		apiResp.Etablissement.Siret == siret &&
		apiResp.Etablissement.StatutDiffusionEtablissement == "O", nil
}
