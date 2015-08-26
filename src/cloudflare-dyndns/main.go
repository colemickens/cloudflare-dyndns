package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/miekg/dns"
    "io/ioutil"
    "log"
    "os"
    "net/http"
    "strings"
)

const (
    envCli string = "env"
    emailCli string = "email"
    apiKeyCli string = "apikey"
    recordsCli string = "records"

    emailEnv string = "CLOUDFLARE_EMAIL"
    apiKeyEnv string = "CLOUDFLARE_APIKEY"
    recordsEnv string = "CLOUDFLARE_RECORDS"
)

var (
    env = flag.Bool(envCli, false, "load from env instead of cli params")
    apiKey = flag.String(apiKeyCli, "", "api key from cloudflare")
    email = flag.String(emailCli, "", "email address for cloudflare")
    records = flag.String(recordsCli, "", "the names of the records to update")

    newIP string
)

// very incomplete, but as much as we should need
type zoneListResponse struct {
    Success bool `json:"success"`
    Errors []string `json:"errors"`
    Messages []string `json:"messages"`
    Result []struct {
        ID string  `json:"id"`
    } `json:"result"`
}

type recordListResponse struct {
    Success bool `json:"success"`
    Errors []string `json:"errors"`
    Messages []string `json:"messages"`
    Result []struct {
        ID string `json:"id"`
        Name string `json:"name"`
        Content string `json:"content"`
        Type string `json:"type"`
    } `json:"result"`
}

type updateRecordRequest struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Content string `json:"content"`
    Type string `json:"type"`
}

var recordsToUpdate []string

func getWanIP() string {
    m := new(dns.Msg)
    m.SetQuestion("o-o.myaddr.l.google.com.", dns.TypeTXT)
    c := new(dns.Client)
    in, _, err := c.Exchange(m, "ns1.google.com:53")
    if err != nil {
        panic(err)
    }

    if t, ok := in.Answer[0].(*dns.TXT); ok {
        return t.Txt[0]
    }

    panic("failed to lookup ip from google")
}

func getZones(client *http.Client) zoneListResponse {
    zoneRequest, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones", nil)
    if err != nil {
        panic(err)
    }

    authHeaders(zoneRequest.Header)

    zoneListResp, err := client.Do(zoneRequest)
    if err != nil {
        panic(err)
    }

    jsonDecoder := json.NewDecoder(zoneListResp.Body)
    defer zoneListResp.Body.Close()

    var zoneList zoneListResponse
    err = jsonDecoder.Decode(&zoneList)
    if err != nil {
        panic(err)
    }

    return zoneList
}

func getRecords(client *http.Client, zoneID string) recordListResponse {
    url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)
    recordRequest, err := http.NewRequest("GET", url, nil)
    if err != nil {
        panic(err)
    }

    authHeaders(recordRequest.Header)

    recordListResp, err := client.Do(recordRequest)
    if err != nil {
        panic(err)
    }

    jsonDecoder := json.NewDecoder(recordListResp.Body)
    defer recordListResp.Body.Close()

    var recordList recordListResponse
    err = jsonDecoder.Decode(&recordList)
    if err != nil {
        panic(err)
    }

    return recordList
}

func updateRecord(client *http.Client, zoneID, recordID, name, typ, content string) error {
    updateURL := fmt.Sprintf(
        "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
        zoneID,
        recordID)

    updateRecordRequestBody := updateRecordRequest{
        ID: recordID,
        Name: name,
        Content: content,
        Type: typ,
    }

    var buffer bytes.Buffer
    jsonEncoder := json.NewEncoder(&buffer)
    err := jsonEncoder.Encode(updateRecordRequestBody)
    if err != nil {
        return err
    }

    updateRequest, err := http.NewRequest("PUT", updateURL, &buffer)
    if err != nil {
        return err
    }

    authHeaders(updateRequest.Header)
    updateRequest.Header.Set("Content-Type", "application/json")

    updateRecordResponse, err := client.Do(updateRequest)
    if err != nil {
        return err
    }

    log.Println(updateRecordResponse.Status)
    if updateRecordResponse.StatusCode != 200 {
        contents, _ := ioutil.ReadAll(updateRecordResponse.Body)
        updateRecordResponse.Body.Close()
        return fmt.Errorf(string(contents))
    }

    return nil
}

func init() {
    flag.Parse()

    if *env {
        *records = os.Getenv(recordsEnv)
        *email = os.Getenv(emailEnv)
        *apiKey = os.Getenv(apiKeyEnv)
    }

    if *records == "" {
        panic(fmt.Sprintf("--%s was empty/missing or %s was empty", recordsCli, recordsEnv))
    }

    if *email == "" {
        panic(fmt.Sprintf("--%s was empty/missing or %s was empty", emailCli, emailEnv))
    }

    if *apiKey== "" {
        panic(fmt.Sprintf("--%s was empty/missing or %s was empty", apiKeyCli, apiKeyEnv))
    }

    newIP = getWanIP()
    recordsToUpdate = strings.Split(*records, ",")
}

func authHeaders(header http.Header) {
    header.Set("X-Auth-Email", *email)
    header.Set("X-Auth-Key", *apiKey)
}

func main() {
    client := &http.Client{}
    zoneList := getZones(client)

    for _, zone := range zoneList.Result {
        recordList := getRecords(client, zone.ID)

        for _, record := range recordList.Result {
            for _, potentialRecord := range recordsToUpdate {
                if record.Name == potentialRecord && record.Type == "A" {
                    if (record.Content == newIP) {
                        log.Printf("skipping %s... already correct\n", record.Name)
                        continue
                    }
                    log.Printf("updating %s...\n", record.Name)

                    err := updateRecord(client, zone.ID, record.ID, record.Name, record.Type, record.Content)
                    if err != nil {
                        panic(err)
                    }

                    log.Printf("%s updated...\n", record.Name)
                }
            }
        }
    }
}

