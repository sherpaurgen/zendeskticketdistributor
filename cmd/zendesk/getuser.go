package zendesk

import (
	b64 "encoding/base64"
	"encoding/json"
	"log"
	"os"
)

type Zds struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}
//zd api basic auth -usercred is base64 encoded string
func GetZdConfig() string {
	var basePath string
	basePath = GetUserHomeDir()
	zdconfigFile := "/.triage/zd.json"
	zdjsonFilePath := basePath + zdconfigFile
	zddata, err := os.ReadFile(zdjsonFilePath)
	if err != nil {
		log.Fatal("Error reading zd config file ~/.triage/zd.json, Please check formatting eg:\n{ \"user\":\"bob@example.com\", \"pass\":\"s3cretPass\" } ", err)
	}
    var zdcred Zds

	err=json.Unmarshal(zddata, &zdcred)
	if err != nil {
		log.Fatal("Problem unmarshalling ~/.triage/zd.json file,Please check formatting eg:\n{ \"user\":\"bob@example.com\", \"pass\":\"s3cretPass\" } ",err)
	}
	return b64.StdEncoding.EncodeToString([]byte(zdcred.User+":"+zdcred.Pass))
}

func GetUserHomeDir() string {
	userHomePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error parsing json files in userHomedir : ", err)
	}
	return userHomePath
}

func GetAgentList() Agents {
	agentListFile:="/.triage/agents.json"
	var basePath string
	basePath = GetUserHomeDir()
	agentjsonFilePath := basePath + agentListFile
	agentdata, err := os.ReadFile(agentjsonFilePath)
	if err != nil {
		log.Fatalf("Error reading agent config file ~/.triage/agents.json, expected format:\n { \"agentlist\":[\"kdb@example.com\",\"mhz@example.com\",\"bsl@example.com\"]}\n\n", err)
	}
	agentnames:= Agents{Agentlist: []string{},Domain: "" }
	err=json.Unmarshal(agentdata, &agentnames)
	if err != nil {
		log.Fatalf("Problem unmarshalling ~/.triage/agents.json file")
	}
	log.Println("Reading agentlist json:",agentnames.Agentlist)
	return agentnames
}
type Agents struct {
	Agentlist []string `json:"agentlist"`
	Domain string `json:"subdomain"`
}