package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"
	"io"
)

var lastPrintedLine string =""
var stop = false

/*
*	Récupère une chaîne de caractères à partir de l'invite de commande du joueur
*/
func GetStringFromConsole(PrintText string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(PrintText + "\n")
	text, _ := reader.ReadString('\n')
	return strings.TrimSuffix(text, "\n")
}

/*
*	Récupère un booléen mis à vrai si le joueur rentre dans son invite de commande "y" ou faux sinon
*/
func GetBoolFromConsole(PrintText string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(PrintText + "(y/n)\n")
	text, _ := reader.ReadString('\n')
	return (text == "y\n")
}

/*
*	Ouvre un scanner qui permet de récupérer ce que l'utilisateur tape dans son invite de commande
*/
func LaunchScan() string {
	stop = false
	snr := bufio.NewScanner(os.Stdin)
	var output string
    for snr.Scan() && (!stop) {
        line := snr.Text()
        if len(line) == 0 {
            break
        }
        output = line
        break
    }
    if err := snr.Err(); err != nil {
        if err != io.EOF {
            fmt.Fprintln(os.Stderr, err)
        }
    } 
	return output
}

/*
*	Ferme le scanner
*/
func StopScan() {
	stop = true
}

/*
*	Enlève tout ce qu'il y a d'écrit dans la console du joueur
*/
func ClearConsole() {
	fmt.Println("\033[2J")
}

/*
*	Alternative à "fmt.Println()" permettant l'immersion car affiche
*	progressivement un texte dans la console comme si quelqu'un le tapait
*/
func PrintProgressively(s string) {
	timeToWait := 25
	b := []byte(s)
	for len(b)>0 {
		r, size := utf8.DecodeRune(b)
		fmt.Print(string(r) + "")
		b = b[size:]
		time.Sleep(time.Duration(timeToWait) * time.Millisecond)
	}
	fmt.Print("\n")
}

/**
* Affiche une chaine de caractères pour pouvoir la supprimer juste après sans faire un clear.
*/
func PrintWithRollBack(s string) {
	fmt.Print(s,"\r")
	lastPrintedLine = s
}

/**
*	Supprime la dernière chaîne de caractères affichée avec PrintWithRollBack
*/
func RollBackPrint() {
	fmt.Print("\r\033[K")
}

/*
*	Paramètre la couleur du texte affiché dans l'invite de commande en ROUGE
*/
func SetRed() {
	fmt.Print("\033[31m")
}

/*
*	Paramètre la couleur du texte affiché dans l'invite de commande en VERT
*/
func SetGreen() {
	fmt.Print("\033[32m")
}

/*
*	Paramètre la couleur du texte affiché dans l'invite de commande en GRAS
*/
func SetBold() {
	fmt.Print("\033[1m")
}

/*
*	Paramètre la couleur du texte affiché dans l'invite de commande en BLEU
*/
func SetCyan() {
	fmt.Print("\033[36m")
}

/*
*	Paramètre la couleur du texte affiché dans l'invite de commande à celle de base
*/
func ResetColor() {
	fmt.Print("\033[0m")
}
