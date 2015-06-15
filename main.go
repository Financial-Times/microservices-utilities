package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"encoding/json"
	"strconv"

	"gopkg.in/yaml.v2"
)

type HostAndPort struct {
	Hostname string `yaml:"host"`
	Port int `yaml:"port"`
}

type BuildInfo struct {
	GroupId string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version string `json:"version"`
}


func main() {
	version := flag.String("versionToCheck", "", "Expected version number to check for in the /build-info endpoint")

	filename, _ := filepath.Abs("./services.yml")
	hostAndPorts := ParseConfig(filename)
	for _, hostAndPort := range hostAndPorts {
		AssertVersion(hostAndPort, *version)
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
	if err != nil{
		panic(err)
	}


	defer resp.Body.Close()
	var deployedServiceInfo BuildInfo
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &deployedServiceInfo)
	if err != nil {
		panic(err)
	}

	if deployedServiceInfo.Version == version {
		return true, "Success"
	}else{
		return true, "Expected: " + version + ", but found: " + deployedServiceInfo.Version
	}

}
