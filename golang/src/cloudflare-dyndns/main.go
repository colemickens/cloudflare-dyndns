package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

var key = flag.String("key", "abc", "api key from cloudflare")

var email = flag.String("email", "someone@example.com", "email address for cloudflare")

var records = flag.String("records", "mickens.us,\\*.mickens.us", "the names of the records to update")

var newIP = flag.String("newIP", "77.88.99.01", "the new ip address for the replacement")

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

func init() {
    flag.Parse()
    recordsToUpdate = strings.Split(*records, ",")
    log.Println("Records to update", recordsToUpdate)
    log.Println("New ip", *newIP)
}

func authHeaders(header http.Header) {
    header.Set("X-Auth-Email", *email)
    header.Set("X-Auth-Key", *key)
}

func main() {
    client := &http.Client{}

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
    /*
    contents, _ := ioutil.ReadAll(zoneListResp.Body)
    log.Println(string(contents))
    */
    var zoneList zoneListResponse
    err = jsonDecoder.Decode(&zoneList)
    if err != nil {
        panic(err)
    }

    log.Println(zoneList)

    for _, zone := range zoneList.Result {
        url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zone.ID)
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

        for _, currentRecord := range recordList.Result {
            log.Printf("checking %s %s %s", currentRecord.Name, currentRecord.Type, currentRecord.Content)
            for _, recordToUpdate := range recordsToUpdate {
                if recordToUpdate == currentRecord.Name && currentRecord.Type == "A" {
                    log.Printf("%s updating...\n", currentRecord.Name)

                    updateURL := fmt.Sprintf(
                        "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
                        zone.ID,
                        currentRecord.ID)

                    updateRecordRequestBody := updateRecordRequest{
                        ID: currentRecord.ID,
                        Name: currentRecord.Name,
                        Content: *newIP,
                        Type: currentRecord.Type,
                    }

                    var buffer bytes.Buffer
                    jsonEncoder := json.NewEncoder(&buffer)
                    err = jsonEncoder.Encode(updateRecordRequestBody)
                    if err != nil {
                        panic(err)
                    }

                    updateRequest, err := http.NewRequest("PUT", updateURL, &buffer)
                    if err != nil {
                        panic(err)
                    }

                    authHeaders(updateRequest.Header)
                    updateRequest.Header.Set("Content-Type", "application/json")

                    updateRecordResponse, err := client.Do(updateRequest)
                    if err != nil {
                        panic(err)
                    }

                    log.Println(updateRecordResponse.Status)
                    if updateRecordResponse.StatusCode != 200 {
                        contents, _ := ioutil.ReadAll(updateRecordResponse.Body)
                        updateRecordResponse.Body.Close()
                        log.Println(string(contents))
                    }

                    _ = updateRecordResponse

                    log.Printf("%s updated...\n", currentRecord.Name)
                }
            }
        }
    }
}
