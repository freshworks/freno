package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/outbrain/golib/log"
)

type ConfigSettings struct {
	ListenPort           int      `json:",omitempty"`
	DataCenter           string   `json:",omitempty"`
	Environment          string   `json:",omitempty"`
	Domain               string   `json:",omitempty"`
	ShareDomain          string   `json:",omitempty"`
	RaftBind             string   `json:",omitempty"`
	RaftDataDir          string   `json:",omitempty"`
	DefaultRaftPort      int      `json:",omitempty"`
	RaftNodes            []string `json:",omitempty"`
	BackendMySQLHost     string   `json:",omitempty"`
	BackendMySQLPort     int      `json:",omitempty"`
	BackendMySQLSchema   string   `json:",omitempty"`
	BackendMySQLUser     string   `json:",omitempty"`
	BackendMySQLPassword string   `json:",omitempty"`
	MemcacheServers      []string `json:",omitempty"`
	MemcachePath         string   `json:",omitempty"`
	EnableProfiling      bool     `json:",omitempty"`
	Stores               StoreSettings
}

type StoreSettings struct {
	MySQL MySQLConfigSettings
}

type MySQLConfigSettings struct {
	User                 string `json:",omitempty"`
	Password             string `json:",omitempty"`
	MetricQuery          string
	CacheMillis          int `json:",omitempty"`
	ThrottleThreshold    float64
	Port                 int      `json:"Port,omitempty"`
	IgnoreDialTcpErrors  bool     `json:",omitempty"`
	IgnoreHostsCount     int      `json:",omitempty"`
	IgnoreHostsThreshold float64  `json:",omitempty"`
	HttpCheckPort        int      `json:",omitempty"`
	HttpCheckPath        string   `json:",omitempty"`
	IgnoreHosts          []string `json:",omitempty"`
	ProxySQLAddresses    []string `json:",omitempty"`
	ProxySQLUser         string   `json:",omitempty"`
	ProxySQLPassword     string   `json:",omitempty"`
	VitessCells          []string `json:",omitempty"`
	Clusters             map[string](*MySQLClusterConfigSettings)
}

type MySQLClusterConfigSettings struct {
	User                 string   `json:",omitempty"`
	Password             string   `json:",omitempty"`
	MetricQuery          string   `json:",omitempty"`
	CacheMillis          int      `json:",omitempty"`
	ThrottleThreshold    float64  `json:",omitempty"`
	Port                 int      `json:",omitempty"`
	IgnoreHostsCount     int      `json:",omitempty"`
	IgnoreHostsThreshold float64  `json:",omitempty"`
	HttpCheckPort        int      `json:",omitempty"`
	HttpCheckPath        string   `json:",omitempty"`
	IgnoreHosts          []string `json:",omitempty"`
	StaticHostsSettings  StaticHostsConfigSettings
}

type StaticHostsConfigSettings struct {
	Hosts []string
}

type DatabaseSecret struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	User     string `json:"username"`
	Slave1   string `json:"slave1_host"`
	Slave2   string `json:"slave2_host"`
	Port     int    `json:"port"`
}

func generate_mysql_store_configuration() MySQLConfigSettings {
	return MySQLConfigSettings{
		ThrottleThreshold: 1.0,
	}
}

func process_mysql_store_configuration(location string, master_shards []string) (MySQLConfigSettings, error) {
	mysqlConfigSettings := generate_mysql_store_configuration()
	mysqlConfigSettings.Clusters = make(map[string]*MySQLClusterConfigSettings)
	current_directory, _ := os.Getwd()
	os.Chdir(location)
	for _, file := range master_shards {
		if strings.Contains(file, "shards") && !strings.Contains(file, "proxy") && !strings.Contains(file, "slave") {
			continue
		}
		var mysqlClusterConfig = new(MySQLClusterConfigSettings)
		master_shard_configuration := strings.Split(file, ".")
		master_shard := master_shard_configuration[len(master_shard_configuration)-1]
		static_hosts := []string{}
		shard_details, err := ioutil.ReadFile(file)
		if err != nil {
			return mysqlConfigSettings, errors.New(err.Error())
		}
		log.Debugf("Populating config for %s cluster", master_shard)
		var db DatabaseSecret
		err = json.Unmarshal(shard_details, &db)
		if err != nil {
			return mysqlConfigSettings, errors.New(err.Error())
		}
		err = json.Unmarshal(shard_details, &mysqlClusterConfig)
		if err != nil {
			return mysqlConfigSettings, errors.New(err.Error())
		}
		mysqlClusterConfig.Port = 3306
		mysqlClusterConfig.User = db.User
		static_hosts = append(static_hosts, db.Slave1, db.Slave2)
		mysqlClusterConfig.StaticHostsSettings = StaticHostsConfigSettings{
			Hosts: static_hosts,
		}
		mysqlConfigSettings.Clusters[master_shard] = mysqlClusterConfig
	}
	os.Chdir(current_directory)
	return mysqlConfigSettings, nil
}

func generate_store_configuration(location string, files []string) (ConfigSettings, error) {
	var err error
	var config = generate_configuration_settings()
	config.Stores.MySQL, err = process_mysql_store_configuration(location, files)
	if err != nil {
		return config, errors.New(err.Error())
	}
	return config, nil
}

func generate_configuration_settings() ConfigSettings {
	return ConfigSettings{
		ListenPort:      8189,
		RaftBind:        "127.0.0.1:10008",
		RaftDataDir:     "/var/lib/freno",
		DefaultRaftPort: 0,
		RaftNodes:       []string{},
	}
}

func generate_config(location string, files []string) error {
	var err error
	freno_config, err := generate_store_configuration(location, files)
	if err != nil {
		return errors.New(err.Error())
	}
	byte_data, err := json.MarshalIndent(freno_config, "", "\t")
	if err != nil {
		return errors.New(err.Error())
	}
	ioutil.WriteFile("/data/freno/shared/freno.conf.json", byte_data, 0)
	return nil
}

func GenerateSecretsConfig(secrets_folder string) error {
	log.Infof("Started generating secrets config")
	files, err := ioutil.ReadDir(secrets_folder) // will return files of type []os.FileInfo
	if err != nil {
		return errors.New(err.Error())
	}
	secret_files := []string{}
	current_directory, _ := os.Getwd()
	os.Chdir(secrets_folder)
	for _, f := range files {
		log.Debugf("Reading file %s", f.Name())
		info, _ := os.Stat(f.Name())
		if info.IsDir() {
			continue
		}
		log.Debugf("Adding %s to secrets_file ", f.Name())
		secret_files = append(secret_files, f.Name())
	}
	err = generate_config(secrets_folder, secret_files)
	if err != nil {
		return errors.New(err.Error())
	} else {
		fmt.Println("Freno Configuration generated successfully")
	}
	os.Chdir(current_directory)
	return nil

}
