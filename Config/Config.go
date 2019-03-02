package Config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type waitTime struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
}

type global struct {
	PositionRange  float64  `yaml:"position_range"`
	BaseLng        float64  `yaml:"base_lng"`
	BaseLat        float64  `yaml:"base_lat"`
	WaitTime       waitTime `yaml:"wait_time"`
	DefaultAddress string   `yaml:"default_address"`
}

type db struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	UserName  string `yaml:"user_name"`
	Password  string `yaml:"password"`
	TableName string `yaml:"table_name"`
	Charset   string `yaml:"charset"`
	Location  string `yaml:"location"`
}

type email struct {
	MasterEmail string `yaml:"master_email"`
	Email       string `yaml:"email"`
	Password    string `yaml:"password"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
}

type dbMap struct {
	Ak string `yaml:"ak"`
}

var Config struct {
	Global global `yaml:"global"`
	Db     db     `yaml:"db"`
	Email  email  `yaml:"email"`
	DbMap  dbMap  `yaml:"db_map"`
}

func init() {
	yamlContent, err := ioutil.ReadFile("./ASCFA_Conf.yaml")

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlContent, &Config)

	if err != nil {
		panic(err)
	}
}
