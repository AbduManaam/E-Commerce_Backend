package config

import (
	"errors"
	"log"

	"os"

	"github.com/go-yaml/yaml"
)

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type JWTConfig struct {
	AccessSecret string `yaml:"access_secret"`
	RefreshSecret string `yaml:"refresh_secret"`
	AccessExpiry int    `yaml:"access_expiry"`
	RefreshExpiry int    `yaml:"refresh_expiry"`
}

type SMTPConfig struct{
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type AppConfig struct {
	Environment string        `yaml:"environment"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	SMTP     SMTPConfig     `yaml:"smtp"`
}

func LoadConfig(path string) (*AppConfig, error) {
	cfg := &AppConfig{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(file, cfg); err != nil {
		return nil, err
	}

	if err:= cfg.validate();err!=nil{
		return nil,err
	}

	log.Println("✅ Config loaded")
	return cfg, nil
}

func(c *AppConfig) validate()error{

	//Server
    if c.Server.Port<=0{
		return errors.New("server port must be greater than 0")
	}
	
	//Database
	if c.Database.Host==""|| c.Database.User==""|| c.Database.DBName==""|| c.Database.Port<=0{
		return errors.New("database host, user, and dbname are required and  port must be greater than 0")
	}

	//JWT
	if c.JWT.AccessSecret==""|| c.JWT.RefreshSecret==""{
		return errors.New("jwt secret must not be empty")
	}
	if c.JWT.AccessExpiry<=0 || c.JWT.RefreshExpiry<=0{
		return errors.New("jwt expiry must be greater than 0")
	}

	//SMTP
	if c.SMTP.Host!="" && c.SMTP.Port<=0{
		return errors.New("smt.port must be greater than 0")
	}

	return nil
}

/*
Config is the part of an application that stores and loads external settings like server port, database credentials,
and JWT secrets from a file instead of hardcoding them, converts those values into structured data the app can use,
so the application can load them at startup and run correctly without hardcoding values.

and allows the same codebase to run in different environments by changing only the configuration.

“Instead of hardcoding them” means not writing fixed values directly inside your source code.
*/



