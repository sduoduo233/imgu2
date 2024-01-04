package i18n

import (
	_ "embed"
	"encoding/json"
)

//go:embed en_us.json
var fileEnglish string

var langEnglish = make(map[string]string)

func init() {
	err := json.Unmarshal([]byte(fileEnglish), &langEnglish)
	if err != nil {
		panic(err)
	}
}

func T(key string) string {
	value, ok := langEnglish[key]
	if !ok {
		panic("translation not found: " + key)
	}
	return value
}
