package utils

import (
	"fmt"
	"strconv"
)
/* Valeur de configuration */
var conf = map[string] string {
	"playersToStart":"4", //nombre de joueur
	"fixedPort":"false", // vrai si le port doit être fixe sinon il sera prit aléatoire
	"port":"5000", // port utilisé si "fixedPort" est à "true"
	"useLocalIP":"true", // utilise l'ip Local si vrai. A utiliser si vous jouer sur la même machine
}

/* Renvoi le valeur sous forme d'un entier*/
func GetConfigInt(key string) (int,error) {
	i, err := strconv.Atoi(conf[key])
	if err!=nil {
		fmt.Println("Error during parsing config int (utils/config.go)..")
		return 0, err
	}

	return i, err
}

/* Renvoi le valeur sous forme d'une chaîne de caractère*/
func GetConfigString(key string) string {
	return conf[key]
}

/* Renvoi le valeur sous forme d'un booléen*/
func GetConfigBool(key string) (bool, error) {
	b, err := strconv.ParseBool(conf[key])
	if err!=nil {
		fmt.Println("Error during parsing config bool (utils/config.go)..")
		return false, err
	}

	return b, err
}