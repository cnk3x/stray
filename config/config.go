package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tidwall/jsonc"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

var ErrUnknownSourceType = errors.New("unknown source type")

func LoadFile[T any](cfg *T, fn string) (err error) {
	var (
		exts          = []string{".json", ".yaml", ".toml", ".jsonc", ".json5", ".ini", ".yml", ".tml"}
		dir, filename = filepath.Split(fn)
		currentExt    = filepath.Ext(filename)
		name          = strings.TrimSuffix(filename, currentExt)
	)

	if idx := slices.IndexFunc(exts, strEqualFold(currentExt)); idx > 0 {
		current := exts[idx]
		for i := 0; i < idx; i++ {
			exts[i+1] = exts[i]
		}
		exts[0] = current
	} else if idx < 0 {
		name = filename
	}

	for _, ext := range exts {
		if e := loadSource(cfg, filepath.Join(dir, name+ext), uniType(ext)); e != nil {
			if os.IsNotExist(e) {
				continue
			}
			err = e
			return
		}
	}
	return
}

func loadSource[T any](cfg *T, source any, typ string) (err error) {
	var data []byte
	switch v := source.(type) {
	case string:
		data, err = os.ReadFile(v)
	case []byte:
		data = v
	default:
		err = ErrUnknownSourceType
	}

	if err != nil {
		return
	}

	switch uniType(typ) {
	case "json":
		err = json.Unmarshal(jsonc.ToJSONInPlace(data), cfg)
	case "yaml":
		err = yaml.Unmarshal(data, cfg)
	case "toml":
		err = toml.Unmarshal(data, cfg)
	case "ini":
		f, e := ini.LoadSources(ini.LoadOptions{SkipUnrecognizableLines: true}, data)
		if e != nil {
			err = e
		} else {
			err = f.MapTo(cfg)
		}
	default:
		err = fmt.Errorf("%w: %s", ErrUnknownSourceType, typ)
	}

	return
}

func strEqualFold(src string) func(string) bool {
	return func(dst string) bool { return len(src) == len(dst) && strings.EqualFold(src, dst) }
}

func uniType(in string) (out string) {
	switch out = strings.ToLower(strings.TrimPrefix(in, ".")); out {
	case ".json", ".jsonc", ".json5":
		out = "json"
	case ".yaml", ".yml":
		out = "yaml"
	case ".toml", ".tml":
		out = "toml"
	case ".ini":
		out = "ini"
	}
	return
}
