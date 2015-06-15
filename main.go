package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
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
	fmt.Println("Service file to check " + 	*configFileName)

	filename, _ := filepath.Abs(*configFileName)
	hostAndPorts := ParseConfig(filename)
	for _, hostAndPort := range hostAndPorts {
		status, message := AssertVersion(hostAndPort, *version)
		if !status {
			fmt.Println("FAIL:" + message)
			os.Exit(1)
		}else{
			fmt.Println(message)
			os.Exit(0)
		}
	}
}

func ParseConfig(filename string) []HostAndPort {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var hostAndPorts []HostAndPort

	err = yaml.Unmarshal(yamlFile, &hostAndPorts)
	if err != nil {
		panic(err)
	}
	return hostAndPorts
}

func AssertVersion(hostAndPort HostAndPort, version string) (bool, string) {
	resp, err := http.Get("http://" + hostAndPort.Hostname + ":" + strconv.Itoa(hostAndPort.Port) + "/build-info")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	var deployedServiceInfo BuildInfoResponse
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &deployedServiceInfo)
	if err != nil {
		panic(err)
	}

	if deployedServiceInfo.BuildInfo.Version == version {
		return true, "Success"
	} else {
		return false, "Expected: " + version + ", but found: " + deployedServiceInfo.BuildInfo.Version
	}

}
