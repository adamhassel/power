package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// An MID MUST be 18 digits
const midLength = 18

type confdata struct {
	Token string `toml:"token"`
	MID  string `toml:"mid"`
}

type Config struct {
	token string `toml:"token"`
	mid  string `toml:"mid"`
}

var conf Config

func init() {

}

func LoadConfig(filename string) (Config, error) {
	err := conf.Load(filename)
	return conf, err
}

func GetConf() Config {
	return conf
}

func (c Config) Token() string {
	return c.token
}

func (c Config) MID() string {
	return c.mid
}

func (c *Config) Load(filename string) error {
	tomlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	var d confdata
	_, err = toml.Decode(string(tomlData), &d)
	if err != nil {
		return err
	}
	c.mid = d.MID
	c.token = d.Token
	if len(c.mid) != midLength {
		return fmt.Errorf("MID is not %d digits", midLength)
	}
	if c.token == "" {
		return errors.New("empty token")
	}
	return nil
}
