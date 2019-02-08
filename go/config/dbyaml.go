package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
)

type slave struct {
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
}

type mapShards map[string]ShardSettings


type ShardSettings struct {
	Database  string `yaml:"database"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Host      string `yaml:"host"`
	NotAShard bool   `yaml:"not_a_shard"`
	Port      int    `yaml:"port"`
	Encoding  string `yaml:"encoding"`
	Slave  slave `yaml:"slave"`
}

type DBConfig struct {
	Database       string `yaml:"database"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Shards mapShards `yaml:"shards"`
}

type GlobalConfig struct {
	configMap map [string] DBConfig
}

func isDBConfigValid (dbConfig DBConfig) bool {
	return (dbConfig.Database != "" && dbConfig.Username != "" && dbConfig.Password != "")
}


func setDBConfig(dbConfig DBConfig, mySQLConfigSettings *MySQLConfigurationSettings  ) {
	clusters := &mySQLConfigSettings.Clusters
	mySQLConfigSettings.User = dbConfig.Username
	mySQLConfigSettings.Password = dbConfig.Password

	if dbConfig.Port != 0 {
		mySQLConfigSettings.Port = dbConfig.Port
	} else {
		mySQLConfigSettings.Port = 3306 //default MySQL port is 3306
	}

	for name, shard := range dbConfig.Shards {
		var shardSettings= new(MySQLClusterConfigurationSettings);
		if shard.Username != "" {
			shardSettings.User = shard.Username;
		} else {
			shardSettings.User = mySQLConfigSettings.User;
		}
		if shard.Password != "" {
			shardSettings.Password = shard.Password
		} else {
			shardSettings.Password = mySQLConfigSettings.Password
		}
		if shard.Port != 0 {
			shardSettings.Port = shard.Port
		} else {
			shardSettings.Port = 3306
		}
		hosts := make([]string, 1)
		if shard.Slave.Host != "" {
			hosts[0] = shard.Slave.Host
		}
		shardSettings.StaticHostsSettings = StaticHostsConfigurationSettings{
			Hosts: hosts,
		}
		(*clusters)[name] = shardSettings
		fmt.Println(shardSettings)
	}
	fmt.Println(clusters)
}


func (settings *ConfigurationSettings) ParseDatabaseYaml(filename string) {
	var globalConfig map [string] DBConfig
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &globalConfig)
	if err != nil {
		panic(err)
	}
	mySQLConfigSettings := &settings.Stores.MySQL
	if mySQLConfigSettings.Clusters == nil {
		mySQLConfigSettings.Clusters = map[string](*MySQLClusterConfigurationSettings){}
	}

	var selectedConfig DBConfig;
	var exists bool;

	if selectedConfig, exists = globalConfig["development"]; !exists {
		if selectedConfig, exists = globalConfig["staging"]; !exists {
			if selectedConfig, exists = globalConfig["production"]; !exists {
				panic("Invalid Database yml file.")
			}
		}
	}

	if  exists && isDBConfigValid(selectedConfig) {
		setDBConfig(selectedConfig, mySQLConfigSettings)
	}
}
