package utils

import (
	"fmt"
	"encoding/json"
)

/*
*	Converti une map associant des chaînes de caractères en JSON
* 	et retourne ce JSON sous forme de chaîne de caractère
*	= Sérialisation
*/
func MapToJson(m map[string]string) (string, error) {
	jsonString, err := json.Marshal(m)

	if err != nil {
		fmt.Println(err)
	}

	return string(jsonString), err
}

/*
*	Converti un JSON passé en paramètre sous forme de string
*	en map associant des chaînes de caractères
*	= Désérialisation
*/
func JsonToMap(j string) (map[string]string, error) {
	var m map[string]string
	err := json.Unmarshal([]byte(j), &m)

	if err != nil {
		fmt.Println(err)
	}

	return m, err
}