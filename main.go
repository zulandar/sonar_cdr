package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

type DidInfo struct {
	Data[] DidInfoData `json:"data"`
	Paginator Paginator `json:"paginator"`
}

type Paginator struct {
	TotalCount int64 `json:"total_count"`
	TotalPages int64 `json:"total_pages"`
	CurrentPage int64 `json:"current_page"`
}

type DidInfoData struct {
	Id int64 `json:"id"`
	Did string `json:"did"`
}

type CallRecord struct {
	VoiceProviderId int64 `json:"voice_provider_id"`
	Records[] CallRecordData `json:"records"`
}

type CallRecordData struct {
	Origination string `json:"origination"`
	Destination string `json:"destination"`
	CallStart string `json:"call_start"`
	Duration int64 `json:"duration"`
}

type CDRResponse struct {
	Data CDRResponseData `json:"data"`
}

type CDRResponseData struct {
	Success bool `json:"success"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Failed to load .env file")
		panic(err)
	}

	sonarUrl := fmt.Sprintf("https://%s:%s@%s/api/v1", os.Getenv("USERNAME"), os.Getenv("PASSWORD"), os.Getenv("SONAR_URL"))
	filename := os.Getenv("CDR_FILE")

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		panic(err)
	}

	response, err := http.Get(sonarUrl + "/system/voice/dids")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject DidInfo
	var completeDidList DidInfo
	json.Unmarshal(responseData, &responseObject)

	if responseObject.Paginator.TotalPages > 1 {
		for i := int64(1); i < responseObject.Paginator.TotalPages + 1; i++ {
			response, err := http.Get(fmt.Sprintf(sonarUrl + "/system/voice/dids?page=%d", i))
			if err != nil {
				fmt.Print(err.Error())
				os.Exit(1)
			}

			responseData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}

			json.Unmarshal(responseData, &responseObject)

			for i := 0; i < len(responseObject.Data); i++ {
				completeDidList.Data = append(completeDidList.Data, responseObject.Data[i])
			}
		}
	}

	// CSV processing
	var csvFile[] CallRecordData
	for _, line := range lines {

		if containedInArray(completeDidList.Data, line[1]) {
			line[1] = fmt.Sprintf("%d%s", 1, line[1])

			duration, err := strconv.ParseInt(line[7], 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}

			data := CallRecordData{
				Origination: line[1],
				Destination: line[2],
				CallStart: line[4],
				Duration: duration,
			}
			csvFile = append(csvFile, data)
		}

	}

	var finalRecord CallRecord
	finalRecord.VoiceProviderId = 1
	finalRecord.Records = csvFile

	b, err := json.Marshal(finalRecord)
	if err != nil {
		fmt.Println(err)
		return
	}

	response, err = http.Post(sonarUrl + "/system/voice/cdr_rating", "application/json", bytes.NewBuffer(b))

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	}

	data, _ := ioutil.ReadAll(response.Body)

	var CDRResponse CDRResponse

	json.Unmarshal(data, &CDRResponse)

	if CDRResponse.Data.Success != true {
		fmt.Println("Received failure")
		os.Exit(1)
	}

	if os.Getenv("ROTATE_FREESWITCH") == "true" {
		freeswitchRotateMaster()
	}

}

func freeswitchRotateMaster() {
	cmd := exec.Command("/bin/fs_cli -x 'cdr_csv rotate'")
	fmt.Println("Rotating CDR logs")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Command finished with error: %v\n", err)
	}
}

func containedInArray(a []DidInfoData, x string) bool {
	for _, n := range a {
		if x == n.Did {
			return true
		}
	}
	return false
}


