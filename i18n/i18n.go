package i18n

import (
	_ "embed"
	"encoding/json"
	"imgu2/services"
)

//go:embed en_us.json
var fileEnglish string

//go:embed zh_cn.json
var fileChineseSimplified string

var langEnglish = make(map[string]string)
var langChineseSimplified = make(map[string]string)

func init() {
	err := json.Unmarshal([]byte(fileEnglish), &langEnglish)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(fileChineseSimplified), &langChineseSimplified)
	if err != nil {
		panic(err)
	}
}

func T(key string) string {
	lang, err := services.Setting.GetLanguage()
	if err != nil {
		panic(err)
	}

	var value string
	var ok bool
	switch lang {
	case "en_us":
		value, ok = langEnglish[key]
	case "zh_cn":
		value, ok = langChineseSimplified[key]
	default:
		panic("invalid language setting: " + lang)
	}

	if !ok {
		panic("translation not found: " + key)
	}
	return value
}
