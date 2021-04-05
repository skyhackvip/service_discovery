package configs

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

type GlobalConfig struct {
	Nodes      []string `yaml:"nodes"`
	Hostname   string   `yaml:"hostname"`
	Env        string   `yaml:"env"`
	HttpServer string   `yaml:"http_server"`
	Protect    bool     `yaml:"protect"`
}

func LoadConfig(c string) (*GlobalConfig, error) {
	configFile, err := ioutil.ReadFile(c)
	if err != nil {
		return nil, err
	}
	config := new(GlobalConfig)
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
