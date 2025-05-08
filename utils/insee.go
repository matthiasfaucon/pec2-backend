package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type InseeToken struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   time.Time
	mu          sync.Mutex
}

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
	inseeToken     *InseeToken
	inseeSiretURL  = "https://api.insee.fr/entreprises/sirene/V3.11/siret/"
	inseeTokenURL  = "https://api.insee.fr/token"
	tokenExpiresIn = 7 * 24 * time.Hour // 7 days
)

func init() {
	inseeToken = &InseeToken{}
}

func (t *InseeToken) isExpired() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.AccessToken == "" || time.Now().After(t.ExpiresAt)
}

func generateInseeToken() error {
	consumerKey := os.Getenv("INSEE_CONSUMER_KEY")
	consumerSecret := os.Getenv("INSEE_CONSUMER_SECRET")

	if consumerKey == "" || consumerSecret == "" {
		return fmt.Errorf("INSEE_CONSUMER_KEY and INSEE_CONSUMER_SECRET are required in environment variables")
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", consumerKey, consumerSecret)))

	req, err := http.NewRequest("POST", inseeTokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with INSEE API: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	inseeToken.mu.Lock()
	defer inseeToken.mu.Unlock()
	inseeToken.AccessToken = tokenResp.AccessToken
	inseeToken.ExpiresAt = time.Now().Add(tokenExpiresIn)

	return nil
}

func ensureValidToken() error {
	if inseeToken.isExpired() {
		if err := generateInseeToken(); err != nil {
			return fmt.Errorf("error generating token: %v", err)
		}
	}
	return nil
}

func VerifySiret(siret string) (bool, error) {
	// Validate SIRET format (14 digits)
	matched, err := regexp.MatchString(`^\d{14}$`, siret)
	if err != nil {
		return false, fmt.Errorf("error validating SIRET format: %v", err)
	}
	if !matched {
		return false, nil
	}

	if err := ensureValidToken(); err != nil {
		return false, err
	}

	req, err := http.NewRequest("GET", inseeSiretURL+siret, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}

	inseeToken.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+inseeToken.AccessToken)
	inseeToken.mu.Unlock()
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
