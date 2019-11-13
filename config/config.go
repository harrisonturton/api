package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Server   `yaml:"server"`
	Auth     `yaml:"authentication"`
	Database `yaml:"database"`
	APIs     `yaml:"api_keys"`
}

type APIs struct {
	SendGrid string `yaml:"sendgrid"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Auth struct {
	PublicKeyPath        string `yaml:"public-key-path"`
	PrivateKeyPath       string `yaml:"private-key-path"`
	RefreshTokenLifespan int    `yaml:"refresh-token-lifespan"`
	AccessTokenLifespan  int    `yaml:"access-token-lifespan"`
}

type Database struct {
	Name     string `yaml:"dbname"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var DefaultConfig = Config{
	Server{
		Host: "localhost",
		Port: 8080,
	},
	Auth{
		PublicKeyPath:  "rsa.app.pub",
		PrivateKeyPath: "rsa.app",
		//RefreshTokenLifespan: 60 * 60 * 24 * 7, // 7 days
		//AccessTokenLifespan:  60 * 15,          // 15 minutes
		RefreshTokenLifespan: 60 * 3, // 3 minutes
		AccessTokenLifespan:  30,     // 30 seconds
	},
	Database{
		Name:     "--my-placeholder-database--",
		Username: "--my-placeholder-username--",
		Password: "--my-placeholder-password--",
	},
	APIs{
		SendGrid: "--sendgrid--",
	},
}

func (config *Config) ReadFromFile(filepath string) error {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, &config)
	if config.Host == "" {
		return errors.New("server:host not specified in config")
	}
	if config.Port == 0 {
		return errors.New("server:port not specified in config")
	}
	if config.PublicKeyPath == "" {
		return errors.New("authentication:publicKeyPath not specified in config")
	}
	if config.PrivateKeyPath == "" {
		return errors.New("authentication:privateKeyPath not specified in config")
	}
	if config.Name == "" {
		return errors.New("database:name not specified in config")
	}
	if config.Username == "" {
		return errors.New("database:username not specified in config")
	}
	if config.SendGrid == "" {
		return errors.New("api_keys:sendgrid not specified in config")
	}
	return err
}

func (config *Config) Serialize() ([]byte, error) {
	return yaml.Marshal(config)
}
