package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// OracleConfig implements Oracle Database configuration
type OracleConfig struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	SID      string `json:"sid" yaml:"sid"`
}

// ConnectionString returns connection string in Oracle format
func (c *OracleConfig) ConnectionString() string {
	return fmt.Sprintf("%s/%s@%s:%s/%s", c.Username, c.Password, c.Host, c.Port, c.SID)
}

// Validate validates Oracle Database configuration
func (c *OracleConfig) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Username, validation.Required),
		validation.Field(&c.Password, validation.Required),
		validation.Field(&c.Host, validation.Required, is.Host),
		validation.Field(&c.Port, validation.Required, is.Port),
		validation.Field(&c.SID, validation.Required),
	)
}

const (
	minFetchPeriod   = 100
	minReviewerCount = 1
)

// CheckingConfig implements configuration for solution checking process
type CheckingConfig struct {
	FetchPeriod         int `json:"fetch_period" yaml:"fetch_period"`
	RestrictionCheckers int `json:"restriction_checkers" yaml:"restriction_checkers"`
	SelectionCheckers   int `json:"selection_checkers" yaml:"selection_checkers"`
}

// Validate validates solution checking configuration
func (c *CheckingConfig) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.FetchPeriod, validation.Required, validation.Min(minFetchPeriod)),
		validation.Field(&c.RestrictionCheckers, validation.Required, validation.Min(minReviewerCount)),
		validation.Field(&c.SelectionCheckers, validation.Required, validation.Min(minReviewerCount)),
	)
}

// Config implements common configuration for the service
type Config struct {
	LoggerConfig        zap.Config              `json:"logger" yaml:"logger"`
	MainDBConfig        OracleConfig            `json:"main_db" yaml:"main_db"`
	SelectionDBsConfigs map[string]OracleConfig `json:"selection_dbs" yaml:"selection_dbs"`
	CheckingConfig      CheckingConfig          `json:"checking" yaml:"checking"`
}

// Validate validates worker configuration
func (c *Config) Validate() error {
	if _, err := c.LoggerConfig.Build(); err != nil {
		return err
	}
	if err := c.MainDBConfig.Validate(); err != nil {
		return err
	}
	for _, cc := range c.SelectionDBsConfigs {
		if err := cc.Validate(); err != nil {
			return err
		}
	}
	if err := c.CheckingConfig.Validate(); err != nil {
		return err
	}
	return nil
}

// LoadFromFile loads worker configuration from a file located in the path
func (c *Config) LoadFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	switch ext := filepath.Ext(path); ext {
	case ".json":
		if err := json.Unmarshal(data, c); err != nil {
			return err
		}
	case ".yaml":
	case ".yml":
		if err := yaml.Unmarshal(data, c); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported configuration file format: %s", ext)
	}

	return c.Validate()
}
