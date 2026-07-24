package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gopkg.in/twindagger/httpsig.v1"
)

const (
	JmsServerURL    = "http://192.168.4.13"
	AccessKeyID     = "96dd1976-7c61-483e-ab4a-d597eb0aa626"
	AccessKeySecret = "Cg0zPrtOfCDyIPNmI95ubbGcU1yezDC4o2qR"
)

type SigAuth struct {
	KeyID    string
	SecretID string
}

func (auth *SigAuth) Sign(r *http.Request) error {
	headers := []string{"(request-target)", "date"}
	signer, err := httpsig.NewRequestSigner(auth.KeyID, auth.SecretID, "hmac-sha256")
	if err != nil {
		return err
	}
	return signer.SignRequest(r, headers, nil)
}

func GetUserInfo(jmsurl string, auth *SigAuth) {
	url := jmsurl + "/api/v1/users/users/"
	gmtFmt := "Mon, 02 Jan 2006 15:04:05 GMT"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("create request failed: %v", err)
	}

	req.Header.Set("Date", time.Now().Format(gmtFmt))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-JMS-ORG", "00000000-0000-0000-0000-000000000002")

	if err := auth.Sign(req); err != nil {
		log.Fatalf("sign request failed: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("read body failed: %v", err)
	}

	fmt.Printf("status code: %d\n", resp.StatusCode)
	fmt.Println("response body:")
	fmt.Println(string(body))
}

func main() {
	auth := SigAuth{
		KeyID:    AccessKeyID,
		SecretID: AccessKeySecret,
	}
	GetUserInfo(JmsServerURL, &auth)
}
