package conf
import(
	"bufio"
	"log"
	"os"
)


func GetAgentEmail() []string {
	var basePath string
	basePath = getUserHomeDir()
	emailList := "/.triage/agentemail"
	emailFilePath := basePath + emailList
	dataPtr, err := os.Open(emailFilePath)
	if err != nil {
		log.Fatal("Error reading email file:", err)
	}
	var arrEmail []string
	fileScanner:=bufio.NewScanner(dataPtr)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		arrEmail=append(arrEmail,fileScanner.Text())
	}
return arrEmail
}


func getUserHomeDir() string {
	userHomePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error parsing json files in userHomedir : ", err)
	}
	return userHomePath
}