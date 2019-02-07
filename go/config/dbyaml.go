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

type StagingConfig struct {
	Staging struct {
		Database       string `yaml:"database"`
		Username       string `yaml:"username"`
		Password       string `yaml:"password"`
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		Shards mapShards `yaml:"shards"`
	} `yaml:"staging"`
}

type DevelopmentConfig struct {
	Development struct {
		Database string    `yaml:"database"`
		Username string    `yaml:"username"`
		Password string    `yaml:"password"`
		Host     string    `yaml:"host"`
		Port     int       `yaml:"port"`
		Shards   mapShards `yaml:"shards"`
	} `yaml:"development"`
}

type ProductionConfig struct {
	Production struct {
		Database string    `yaml:"database"`
		Username string    `yaml:"username"`
		Password string    `yaml:"password"`
		Host     string    `yaml:"host"`
		Port     int       `yaml:"port"`
		Shards   mapShards `yaml:"shards"`
	} `yaml:"production"`
}


func (settings *ConfigurationSettings) ParseDatabaseYaml(filename string) {
	var stagingConfig StagingConfig
	var productionConfig ProductionConfig
	var developmentConfig DevelopmentConfig
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &productionConfig)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &stagingConfig)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &developmentConfig)
	if err != nil {
		panic(err)
	}
	mySQLConfigSettings := &settings.Stores.MySQL
	if mySQLConfigSettings.Clusters == nil {
		mySQLConfigSettings.Clusters = map[string](*MySQLClusterConfigurationSettings){}
	}
	clusters := &mySQLConfigSettings.Clusters

	if (developmentConfig.Development.Database != "") {
		if developmentConfig.Development.Username != "" {
			mySQLConfigSettings.User = developmentConfig.Development.Username
		}

		if developmentConfig.Development.Password != "" {
			mySQLConfigSettings.Password = developmentConfig.Development.Password
		}

		if developmentConfig.Development.Port != 0 {
			mySQLConfigSettings.Port = developmentConfig.Development.Port
		} else {
			mySQLConfigSettings.Port = 3306 //default MySQL port is 3306
		}

		for name, shard := range developmentConfig.Development.Shards {
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
	} else if (stagingConfig.Staging.Database != "") {
		mySQLConfigSettings.User = stagingConfig.Staging.Username
		mySQLConfigSettings.Password = stagingConfig.Staging.Password
		mySQLConfigSettings.Port = stagingConfig.Staging.Port

		for name, shard := range stagingConfig.Staging.Shards {
			var shardSettings= new(MySQLClusterConfigurationSettings);
			shardSettings.User = shard.Username;
			shardSettings.Password = shard.Password
			shardSettings.Port = shard.Port
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
	} else if (productionConfig.Production.Database != "") {
		mySQLConfigSettings.User = productionConfig.Production.Username
		mySQLConfigSettings.Password = productionConfig.Production.Password
		mySQLConfigSettings.Port = productionConfig.Production.Port

		for name, shard := range productionConfig.Production.Shards {
			var shardSettings= new(MySQLClusterConfigurationSettings);
			shardSettings.User = shard.Username;
			shardSettings.Password = shard.Password
			shardSettings.Port = shard.Port
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
	} else {
		panic("Invalid Database yml file.")
	}
	fmt.Println(clusters)
}
