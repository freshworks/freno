package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
)

//type slave struct {
//	Hosts   []string `yaml:"host"`
//}

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

type DbConfig struct {
	Staging struct {
		Database       string `yaml:"database"`
		Username       string `yaml:"username"`
		Password       string `yaml:"password"`
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		Shards mapShards `yaml:"shards"`
	} `yaml:"staging"`
}


func (settings *ConfigurationSettings) ParseDatabaseYaml(filename string) {
	var config DbConfig
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
	mySQLConfigSettings := &settings.Stores.MySQL
	if mySQLConfigSettings.Clusters == nil {
		mySQLConfigSettings.Clusters = map[string](*MySQLClusterConfigurationSettings){}
	}
	clusters := &mySQLConfigSettings.Clusters

	mySQLConfigSettings.User = config.Staging.Username
	mySQLConfigSettings.Password = config.Staging.Password
	mySQLConfigSettings.Port = config.Staging.Port

	for name, shard := range config.Staging.Shards {
		var shardSettings = new(MySQLClusterConfigurationSettings);
		shardSettings.User = shard.Username;
		shardSettings.Password = shard.Password
		shardSettings.Port = shard.Port
		if (shard.Slave.Hosts != nil) {
			shardSettings.StaticHostsSettings = StaticHostsConfigurationSettings{
				Hosts: shard.Slave.Hosts,
			}
		}
		//shardSettings.
		(*clusters)[name] = shardSettings
		fmt.Println(shardSettings)
	}
	fmt.Println(clusters)
}
