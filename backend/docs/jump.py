// Golang 示例
package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "gopkg.in/twindagger/httpsig.v1"
)

const (
    JmsServerURL = "https://demo.jumpserver.org"
    AccessKeyID = "cf22ac89-e7f5-4073-8420-bebf7ae7061a"
    AccessKeySecret = "OytznFV9MYLr9QIIY0A33Y861XT0YsVYsXCm"
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
    req.Header.Add("Date", time.Now().Format(gmtFmt))
    req.Header.Add("Accept", "application/json")
    req.Header.Add("X-JMS-ORG", "00000000-0000-0000-0000-000000000002")
    if err != nil {
        log.Fatal(err)
    }
    if err := auth.Sign(req); err != nil {
        log.Fatal(err)
    }
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    json.MarshalIndent(body, "", "    ")
    fmt.Println(string(body))
}

func main() {
    auth := SigAuth{
        KeyID:    AccessKeyID,
        SecretID: AccessKeySecret,
    }
    GetUserInfo(JmsServerURL, &auth)
}
