package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml/yaml.v2"
)

type HostAndPort struct {
	Hostname string `yaml:"host"`
	Port     int    `yaml:"port"`
}

type BuildInfo struct {
	GroupId    string `json:"artifact.groupId"`
	ArtifactId string `json:"artifact.id"`
	Version    string `json:"artifact.version"`
}

type BuildInfoResponse struct {
	BuildInfo BuildInfo `json:"buildInfo"`
}

var version = flag.String("version", "", "Expected version number to check for in the /build-info endpoint")
var configFileName = flag.String("config", "", "Expected a yml file with the host and port configuration for the services to check")

func main() {
	flag.Parse()
	fmt.Println("Version to check " + *version)
	fmt.Println("Service file to check " + *configFileName)

	filename, err := filepath.Abs(*configFileName)
	if err != nil {
		fmt.Printf("Unable to read the file: %s. Error is %v\n", *configFileName, err)
		panic(err)
	}
	hostAndPorts := ParseConfig(filename)
	for _, hostAndPort := range hostAndPorts {
		status, message := AssertVersion(hostAndPort, *version)
		if !status {
			fmt.Println("FAIL:" + message)
			os.Exit(1)
		} else {
			fmt.Println(message)
			os.Exit(0)
		}
	}
}

func ParseConfig(filename string) []HostAndPort {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("error reading yaml file: %v\n", err)
		panic(err)
	}

	var hostAndPorts []HostAndPort

	err = yaml.Unmarshal(yamlFile, &hostAndPorts)
	if err != nil {
		fmt.Printf("error parsing yaml file: %v\n", err)
		panic(err)
	}
	return hostAndPorts
}

func AssertVersion(hostAndPort HostAndPort, version string) (bool, string) {
	url := "http://" + hostAndPort.Hostname + ":" + strconv.Itoa(hostAndPort.Port) + "/build-info"
	fmt.Printf("Cheking %v for version %v\n", url, version)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Unable to get to the /build-info endpoint. Error is : %v\n", err)
		panic(err)
	}

	defer resp.Body.Close()
	var deployedServiceInfo BuildInfoResponse
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Unable to read the Response. Error is : %v\n", err)
		panic(err)
	}

	err = json.Unmarshal(data, &deployedServiceInfo)
	if err != nil {
		fmt.Printf("Unable to Unmarshall json at /build-info endpoint. Error is : %v\n", err)
		panic(err)
	}

	if deployedServiceInfo.BuildInfo.Version == version {
		fmt.Printf("Version Check succeeded for %v with version %v\n", url, version)
		return true, "Success"
	} else {
		return false, "Expected: " + version + ", but found: " + deployedServiceInfo.BuildInfo.Version
	}

}
