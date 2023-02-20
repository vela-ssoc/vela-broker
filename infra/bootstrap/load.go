package bootstrap

import (
	"encoding/json"
	"os"

	"github.com/vela-ssoc/vela-broker/infra/encipher"
)

func AutoLoad(cfg, exe string, v any) error {
	if cfg != "" {
		return unmarshalJSON(cfg, v)
	}

	return readFile(exe, v)
}

func unmarshalJSON(name string, v any) error {
	open, err := os.Open(name)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer open.Close()

	return json.NewDecoder(open).Decode(v)
}

func readFile(name string, v any) error {
	return encipher.ReadFile(name, v)
}
