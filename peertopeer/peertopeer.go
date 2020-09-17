package peertopeer

import (
	"bufio"
	"fmt"
	utils "loup-garou-go/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"io/ioutil"
	"net/http"
	"bytes"
)
/*
	Déclaration des variables globales
*/
var myPseudo string
var myID int
var currentHost int


var myIp string
var suc string

var currentPeersLength int

var events []func(map[string]string)
var updateMyPlayer func(int, string)


/*
*	pseudo : Chaîne de caractères indiquant le pseudo actuel du joueur
* 	Créer un serveur sur la machine aux quels les peers pourront se connecter.
*/
func PeerToPeer(pseudo string) {
	// Initialisation des informations (ID du client, pseudonyme et la longueur de la chaîne des peers)
	myPseudo = pseudo
	myID = 1
	currentPeersLength = 1

	// Définition du port du serveur, ":0" pour utiliser un port random
	currentPort := ":0"
	// Récupération dans la configuration de la variable "fixedPort" qui indique si nous devons utiliser un port défini ou non
	useFixedPort, boolError := utils.GetConfigBool("fixedPort")

	// Si une erreur survient lors du parsing en booléen de la configuration alors on l'affiche et quitte l'application.
	if boolError != nil {
		fmt.Println(boolError)
		os.Exit(1)
	}

	// Si le port doit être fixe, on modifie le port du serveur par celui demandé dans la configuration
	if useFixedPort {
		currentPort = ":"+utils.GetConfigString("port")
	}

	// Lancement du serveur en TCP
	listener, err := net.Listen("tcp", currentPort)
	if err != nil {
		fmt.Println(err)
	}

	// Récupération de l'adresse du serveur sous forme ip:port
	myIp = getIp() + ":" + strconv.FormatInt(int64(listener.Addr().(*net.TCPAddr).Port), 10)

	for {
		//mise en place de la connexion
		conn, err := listener.Accept()
		// Si une erreur est survenue, on affiche l'erreur et quitte l'application
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Lancement du gestionnaire de connexion qui va gérer tous les paquets entrants
		go handleConn(conn)
	}
}

/*
*	conn : Connexion à gérer
*	Gère une connexion et les paquets qu'elle envoit
*/
func handleConn(conn net.Conn) {
	// Création d'une lecture et d'une écriture sur la connexion
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Reception des paquets entrant
	message := ReceiveMessage(reader)

	// Récupération de l'id du client à l'origine du paquet
	ID, _ := strconv.Atoi(message["id"])

	// Si mon id est différent de celui qui est à l'origine du paquet
	if ID != myID {

		// Traitement du paquet en fonction de l'action qui est décrite à l'intérieur
		switch message["action"] {
		case "askSucc": // Le client demande son successeur pour rejoindre la partie
			// Récupération de l'ip du client
			newConnectionIp := message["ip"]
			// Incrémentation de la taille des peers connectés
			currentPeersLength++
			// Préparation du paquet qui va être envoyé au client lui indiquant son successeur
			var toSend = map[string]string{
				"action": "setSuc",
				"ip":     myIp, // Par défaut, on envoie notre adresse IP si nous n'avons pas de successeur, on devient le successeur du client et lui le notre
				"setId":  strconv.Itoa(currentPeersLength), // On envoie au client le nombre de peers connectés actuellement
			}
			if len(suc) > 0 { // Si nous avons déjà un successeur
				toSend["ip"] = suc // Le successeur du client sera notre successeur actuel
			}

			// Election leader : la dernière personne qui a fait entrer un client dans la chaîne devient l'host de la partie à son lancement
			currentHost = myID
			// On définit le client comme notre successeur
			suc = newConnectionIp

			// On envoie le paquet
			sendMessage(toSend, writer)
			time.Sleep(250 * time.Millisecond)
			// Préparation du paquet qui va servir pour mettre à jour le nombre de peers connectés actuellement
			toSend = map[string]string{
				"action": "updatePeerLength",
				"length": strconv.Itoa(currentPeersLength), // On indique le nombre de peers connectés
			}

			// Envoi du paquet à son successeur, qui le transmettera à son successeur etc
			sendMessageTo(suc, toSend)
			break
		case "setSuc": // On reçoit notre successeur
			// Notre successeur devient celui indiqué dans le paquet
			suc = message["ip"]
			// On récupère notre ID dans la chaîne des peers
			myID, _ = strconv.Atoi(message["setId"])

			// On prépare le paquet indiquant que nous avons rejoint la chaîne des peers
			var toSend = map[string]string{
				"action":     "playerJoined",
				"playerName": myPseudo, // On envoie notre pseudo
			}

			// Envoi du paquet à notre successeur
			SendMessageToSuc(toSend)
			// Mise à jour de notre joueur dans la partie, permet de récupérer le pseudonyme entré au lancement du programme.
			go updateMyPlayer(myID, myPseudo)
			break
		case "updatePeerLength": // Le nombre des peers connectés a été modifié
			// L'host devient celui qui a fait rejoindre le nouveau client
			currentHost = ID
			// Mise à jour du nombre de personnes connectées
			currentPeersLength, _ = strconv.Atoi(message["length"])
			break
		}

		// On transfère le message à notre successeur
		tryTransfert(message)
	}

	// On lance tous les listeners enregistrés à la réception d'un message
	onNewMessageReceive(message)
}

/*
*	m : paquet à transféré à son successeur
*	Transfère un paquet à son successeur
*/
func tryTransfert(m map[string]string) {
	// Récupération de l'action demandée dans le paquet
	action := m["action"]

	// Si l'action n'est pas une demande de successeur ou une reception de son successeur
	if action != "askSucc" && action != "setSuc" {
		// On transfère le message à son successeur
		SendMessageToSuc(m)
	}
}

/*
*	Retourne le nombre de peers connectés
*/
func GetPeersLength() int {
	return currentPeersLength
}

/*
*	Retourne true si nous sommes l'host de la partie, false sinon
*/
func AmIHost() bool {
	return (myID == currentHost)
}

/*
*	m : paquet à envoyer
*	Envoi un message à son successeur
*/
func SendMessageToSuc(m map[string]string) {
	sendMessageTo(suc, m)
}

/*
*	ip : Adresse ip du destinataire
*	m : paquet à envoyer
*	Envoi un paquet à l'adresse ip indiquée
*	Renvoi la connexion effectuée à la suite de l'envoi du paquet
*/
func sendMessageTo(ip string, m map[string]string) net.Conn {
	// Connexion au destinataire en TCP
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println(err)
	}

	// Création d'une écriture pour la connexion
	writer := bufio.NewWriter(conn)
	// Envoi du message
	sendMessage(m, writer)

	// Retourne la connexion effectuée
	return conn
}

/*
*	message : Paquet à envoyer
*	writer  : Ecriture de la connexion
*	Envoi un message sur un writer d'une connexion
*/
func sendMessage(message map[string]string, writer *bufio.Writer) {
	go func() {
		// Si l'id d'origine n'est pas indiquée dans le paquet
		if len(message["id"]) == 0 {
			// On associe notre id au paquet
			message["id"] = strconv.Itoa(myID)
		}

		// Sérialisation de la map en json
		json, err := utils.MapToJson(message)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Ecriture du json
		_, err = writer.WriteString(json + "\n")
		if err != nil {
			fmt.Println(err)
		}

		// Envoi du message
		writer.Flush()
	}()
}

/*
*	reader : Lecteur de la connexion
*	Reçoit les messages sur une connexion
*	Retourne le message reçu sous forme d'une map
*/
func ReceiveMessage(reader *bufio.Reader) map[string]string {
	// Lecture du message reçu
	message, err := reader.ReadString('\n')
	// Si la longueur du message est égale à 0 (message null)
	if len(message) == 0 {
		// On retourne une map indiquant la réception d'un paquet null
		var voidMessage = map[string]string{
			"id":     "-1",
			"action": "voidMessage",
		}
		return voidMessage
	}

	message = strings.TrimSuffix(message, "\n") //enlève le \n de la chaine de caratère
	if err != nil {
		fmt.Println("Error in trimSuffix (receive Message)", err)
		return nil
	}

	// Déserialisation du json en map
	m, err := utils.JsonToMap(message)
	if err != nil {
		os.Exit(1)
	}

	// Retourne la map
	return m
}

/*
*	ip : Adresse ip du client pour rejoindre la partie
*	Se connecte à un des peers d'une partie et la rejoint
*/
func TryConnect(ip string) {
	// Initialisation de notre ID à -1 le temps d'en recevoir un de la partie
	myID = -1

	// Préparation du paquet demandant un successeur
	var toSend = map[string]string{
		"action": "askSucc",
		"ip":     myIp, // Correspond à notre adresse  IP, pour pouvoir recevoir une réponse
	}

	// Envoi du message à l'adresse ip et récupération de la connexion associée
	conn := sendMessageTo(ip, toSend)
	// Gestionnaire de la connexion
	handleConn(conn)
	// Affiche notre adresse IP à l'écran pour inviter d'autres personnes à rejoindre la partie
	ShowMyIP()
}

/*
*	Récupère notre adresse ip
*	Retourne une chaîne de caractère correspondant à notre adresse ip locale ou externe
*/
func getIp() string {
	// Récupération du booléen useLocalIP dans la configuration
	useLocalIP, boolErr := utils.GetConfigBool("useLocalIP")

	// S'il faut utiliser l'ip locale ou qu'il y a une erreur dans la récupération de la configuration
	if useLocalIP || boolErr != nil {
		// Renvoit l'adresse ip locale
		return "127.0.0.1"
	} else { // Sinon
		// Connexion au site http://checkip.amazonaws.com qui nous renvoit notre adresse ip externe
		rsp, err := http.Get("http://checkip.amazonaws.com")
		if err != nil {
			fmt.Println("Une erreur est survenue, utilisation de l'ip locale.")
			return "127.0.0.1"
		}
		// Fermeture de la connexion une fois les instructions terminées
		defer rsp.Body.Close()

		// Lecture de la réponse donnée par le site
		buf, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			fmt.Println("Une erreur est survenue, utilisation de l'ip locale.")
			return "127.0.0.1"
		}

		// Renvoit l'adresse ip externe
		return string(bytes.TrimSpace(buf))
	}
	

}

/*
*	Affiche notre adresse IP
*/
func ShowMyIP() {
	for len(myIp) == 0 {
		time.Sleep(200*time.Millisecond)
	}
	fmt.Println("Voici votre adresse IP :", myIp, "\nVos amis peuvent vous rejoindre en indiquant cette adresse.")
}

/*
*	f : Fonction qui recevra les paquets pour pouvoir les traiter
*	Ajoute une fonction en listener qui s'éxecutera à chaque réception de paquet
*/
func AddOnMessageReceiveListener(f func(map[string]string)) {
	events = append(events, f)
}

/*
*	f : Fonction qui met à jour le joueur
*	Modifie la fonction qui met à jour le joueur une fois qu'il a rejoint la partie
*/
func UpdatePlayerListener(f func(int, string)) {
	updateMyPlayer = f
}

/*
*	message : Paquet reçu du prédécesseur
*	Lance les listeners écoutant les paquets reçus
*/
func onNewMessageReceive(message map[string]string) {
	// Récupère tous les listeners
	for _, f := range events {
		// Lance le listener avec le nouveau paquet comme argument
		go f(message)
	}
}
