package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
)

const (
	Version        = "0.0.1.1"
	ConfigFileName = "config.json"
)

type Config struct {
	Version         string `json:"-"`
	Host            string `json:"host"`
	Port            string `json:"port"`
	LogFileName     string `json:"log_file_name"`
	CookieIDName    string `json:"cookie_id_name"`
	CookieMaxLength int    `json:"cookie_max_length"`
}

func LoadConfig() (*Config, error) {
	cfg := new(Config)
	cfg.Version = Version

	jsonFile, err := ioutil.ReadFile(ConfigFileName)
	if err != nil {
		return getDefaultSettings(), err
	}

	err = json.Unmarshal(jsonFile, &cfg)
	if err != nil {
		return getDefaultSettings(), err
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *Config) Save() error {
	jsonObject, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(ConfigFileName, jsonObject, 0644)
	if err != nil {
		return err
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

func getDefaultSettings() *Config {
	return &Config{
		Version:         Version,
		LogFileName:     "q50w.log",
		Host:            "127.0.0.1",
		Port:            "8080",
		CookieIDName:    "watch_id",
		CookieMaxLength: 15,
	}
}
