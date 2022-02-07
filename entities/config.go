package entities

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// An MID MUST be 18 digits
const midLength = 18

type Config struct {
	Token string `toml:"token"`
	MID   string `toml:"mid"`
}

func (c *Config) Load(filename string) error {
	tomlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	_, err = toml.Decode(string(tomlData), &c)
	if err != nil {
		return err
	}
	if len(c.MID) != midLength {
		return fmt.Errorf("MID is not %d digits", midLength)
	}
	if c.Token == "" {
		return errors.New("empty token")
	}
	return nil
}
