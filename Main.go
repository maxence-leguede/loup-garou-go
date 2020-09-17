/*
* @author Maxence Leguede, Josik Sallaud, Baptiste Batard
* @version 1.0
* @date 25/03/2020
*/package main

import (
	"fmt"
	party "loup-garou-go/partie"
	peer "loup-garou-go/peertopeer"
	utils "loup-garou-go/utils"
	"strings"
)
/* Fonction principal du jeu*/
func main() {
	pseudo := utils.GetStringFromConsole("Entrez-votre pseudonyme :")
	fmt.Println(pseudo)
	createParty := utils.GetBoolFromConsole("Voulez-vous créer la partie ?")

	go peer.PeerToPeer(pseudo) //Lance le système de connexion
	peer.AddOnMessageReceiveListener(party.HandleIncomingPackets)

	if !createParty {
		serverToConnect := utils.GetStringFromConsole("Quelle est l'ip à rejoindre ?")
		if !strings.Contains(serverToConnect, ":") {
			fmt.Println("Erreur, l'ip indiquée n'est pas correcte, le format doit être ip:port")
			return
		}

		peer.UpdatePlayerListener(party.UpdateMyPlayer)
		peer.TryConnect(serverToConnect)

	} else {
		party.UpdateMyPlayer(1, pseudo)
		peer.ShowMyIP()
	}
	for { // boucle infinie pour éviter que le programme s'arrête

	}
}
