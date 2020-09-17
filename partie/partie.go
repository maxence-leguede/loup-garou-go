
package partie

import (
	"fmt"
	peer "loup-garou-go/peertopeer"
	types "loup-garou-go/types"
	utils "loup-garou-go/utils"
	"os"
	"strconv"
	"time"
)
/*
	Déclaration des variables globales
*/
var myPlayer types.Player
var currentPlayers []types.Player
var tour int = 0
var isNight bool
var lastVote int = -1
var victim int = -1
var isInVote bool = false

/* Listener qui gère la réception des messages concernant la partie*/
func HandleIncomingPackets(message map[string]string) {
	// Récupération de l'action indiquée dans le paquet
	action := message["action"]

	// Récupération de l'id du peer à l'origine du message
	id, _ := strconv.Atoi(message["id"])
	switch action {
	case "updatePeerLength": // S'il faut mettre à jour le nombre de peers connectés
		// Récupération du nombre de joueurs pour lancer la partie
		minP, _ := utils.GetConfigInt("playersToStart")

		// Si le nombre de peers connectés correspond au nombre de joueurs pour lancer la partie
		if peer.GetPeersLength() == minP {
			// Clear de la console
			utils.ClearConsole()
			fmt.Println("Lancement de la partie en cours...")
			// Si je suis le host
			if peer.AmIHost() {
				// Génération des rôles et distribution de ceux-ci aux joueurs
				distributeRoles(utils.GenerateRoles(peer.GetPeersLength()))
			}
		}

	case "setSuc": // Si nous venons de recevoir notre successeur
		// Préparation du paquet lui demandant les joueurs actuels
		var toSend = map[string]string{
			"action": "requestPlayers",
		}
		// Envoi du paquet au successeur
		peer.SendMessageToSuc(toSend)
		break
	case "playerJoined": // Si un joueur vient de rejoindre
		// Récupération de son pseudonyme
		pseudo := message["playerName"]

		// Si notre liste des joueurs ne contient pas déjà celui-ci
		if contains(currentPlayers, id) == false {
			// Ajout du joueur dans la liste
			currentPlayers = append(currentPlayers, types.Player{Id: id, Pseudo: pseudo, Role: "null", Votes: 0, IsAlive: true})
			fmt.Println(pseudo, "a rejoint la partie !")
		}

		break
	case "requestPlayers": // S'il a été demandé les joueurs
		// Préparation du paquet indiquant que nous sommes dans la partie
		var toSend = map[string]string{
			"action":     "playerJoined",
			"playerName": myPlayer.Pseudo,
		}
		// Envoie du paquet au successeur
		peer.SendMessageToSuc(toSend)

	case "setRole": // Si nous devons mettre à jour un rôle
		// Nous récupérons l'id du joueur destinataire
		targetId, _ := strconv.Atoi(message["targetId"])
		// Récupération de son rôle
		role := message["role"]
		
		// Si l'id du destinataire est le notre
		if targetId == myPlayer.Id {
			// Mise à jour de notre rôle
			myPlayer.Role = role
			utils.ClearConsole()

			// Affichage de notre rôle et de sa description
			utils.SetBold()
			if isLoupGarou() {
				utils.SetRed()
			} else {
				utils.SetGreen()
			}
			fmt.Println("Vous êtes " + role)
			utils.ResetColor()
			fmt.Println(utils.GetRoleDescription(role))
			fmt.Println("")
			utils.PrintProgressively("La partie est sur le point de se lancer...")
		}

		// On retrouve le joueur en question dans la liste des joueurs
		player, index := findPlayerById(targetId)
		// Si le joueur en question existe bien
		if index >= 0 {
			// Mise à jour du joueur dans la liste des joueurs avec son rôle
			newPlayer := types.Player{Id: player.Id, Role: role, Pseudo: player.Pseudo, Votes: 0, IsAlive: true}
			currentPlayers[index] = newPlayer
		}
	case "launchParty": // Si on doit lancer la partie
		// Lancement de la partie
		Partie()

	case "chat": // Si on reçoit un nouveau message dans le chat
		if id != myPlayer.Id { // Si celui qui envoit le message n'est pas le client actuel
			// On récupère celui qui a envoyé le message
			sender, _ := findPlayerById(id)
			// On récupère son message
			msg := message["message"]
			// On récupère le channel du message
			channel := message["channel"]

			// Si le channel est celui des loups-garou
			if channel == "Loup-garou" {
				// Si je suis loup-garou
				if myPlayer.Role == channel {
					// Affichage du message
					utils.SetBold()
					utils.SetRed()
					fmt.Print(sender.Pseudo, ": ")
					utils.ResetColor()
					fmt.Print(msg, "\n")
				}
			} else {
				// Affichage du message
				fmt.Print(sender.Pseudo, ": ")
				utils.ResetColor()
				fmt.Print(msg, "\n")
			}
		}
	case "vote": // Si un vote a été effectué
		// Récupération de l'id du joueur à l'origine du vote
		sender, _ := findPlayerById(id)
		// Récupération de l'id du joueur voté
		targetId, _ := strconv.Atoi(message["targetId"])
		// Récupération du joueur voté
		target, index := findPlayerById(targetId)
		// Incrémentation de son nombre de votes
		target.Votes = target.Votes + 1
		currentPlayers[index] = target

		// Récupération du vote à retirer (si le joueur avait voté pour quelqu'un d'autre précédemment)
		removeId, _ := strconv.Atoi(message["removeVoteFor"])

		// S'il avait bien voté pour quelqu'un d'autre précédemment
		if removeId > -1 {
			// On retire un vote au joueur
			rem, ind := findPlayerById(removeId)
			rem.Votes = rem.Votes - 1
			currentPlayers[ind] = rem
		}

		// Si ce n'est pas la nuit ou que je suis loup-garou
		if !isNight || myPlayer.Role == "Loup-garou" {
			// Affichage du vote
			utils.SetCyan()
			fmt.Println(sender.Pseudo, "a voté contre", target.Pseudo)
			utils.ResetColor()
		}

		// Si je suis le host
		if peer.AmIHost() {
			// Si c'est la nuit
			if isNight {
				// Si les loups-garous se sont mis d'accord sur la victime
				if target.Votes == GetLoupsCount() {

					// Préparation du paquet indiquant la fin de la nuit
					var toSend = map[string]string{
						"action": "endNight",
						"target": strconv.Itoa(targetId), // On définit la victime des loups-garous
					}
					// Envoi du paquet à son successeur
					peer.SendMessageToSuc(toSend)
				}
			}

		}

	case "endNight": // S'il s'agit de la fin de la nuit
		// Récupération de l'id de la victime des loups
		victim, _ = strconv.Atoi(message["target"])
		// Définition du jour
		isNight = false

		// Récupération du joueur qui est la victime et on le décrit comme étant mort
		player, ind := findPlayerById(victim)
		player.IsAlive = false
		currentPlayers[ind] = player

		// Si le joueur tué est égal au client
		if player.Id == myPlayer.Id {
			// On se décrit comme mort
			p := myPlayer
			p.IsAlive = false
			myPlayer = p
		}

	case "endDay": // S'il s'agit de la fin du jour
		// Récupération de l'id de la victime des votes
		victim, _ = strconv.Atoi(message["target"])

		// Si aucun joueur n'a été sélectionné par le village
		if victim == -1 {
			// On affiche l'égalité
			utils.PrintProgressively("Le village a été indécis...")
		} else { // Sinon
			// On récupère le joueur étant la victime du vote et le décrit comme mort
			player, ind := findPlayerById(victim)
			player.IsAlive = false
			currentPlayers[ind] = player

			// Si le joueur tué est égal au client
			if player.Id == myPlayer.Id {
				// On se décrit comme mort
				p := myPlayer
				p.IsAlive = false
				myPlayer = p
			}

			// On affiche le mort et son rôle
			utils.PrintProgressively("Le village a décidé d'éliminer " + player.Pseudo + ", il était...")
			time.Sleep(500 * time.Millisecond)
			utils.SetBold()
			if player.Role == "Loup-garou" {
				utils.SetRed()
			} else {
				utils.SetGreen()
			}

			fmt.Println(player.Role, "!")
			time.Sleep(time.Second)
			// On définit qu'il s'agit de la nuit
			isNight = true

			// On vérifie si c'est la fin de la partie (plus de loups-garous ou villageois)
			if verifVictoire() {
				partyEnd()
			}
		}
	}

}

/* Fonction qui permet de mettre à jour son joueur */
func UpdateMyPlayer(myID int, myPseudo string) {
	myPlayer = types.Player{Id: myID, Pseudo: myPseudo, Role: "null", Votes: 0, IsAlive: true}
	currentPlayers = append(currentPlayers, types.Player{Id: myID, Pseudo: myPseudo, Role: "null", Votes: 0, IsAlive: true})
}
/* Fonction destiner a distribué les roles */
func distributeRoles(roles map[int]string) {
	// Pour chaque roles tirés
	for index, r := range roles {
		// On l'applique au joueur désigné
		currentPlayers[index].Role = r

		// Préparation du paquet indiquant le role du joueur
		var toSend = map[string]string{
			"action":   "setRole",
			"targetId": strconv.Itoa(currentPlayers[index].Id), // Décrit l'id du joueur
			"role":     r, // Décrit son rôle tiré
		}

		// Envoi du message à son successeur
		peer.SendMessageToSuc(toSend)
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(2 * time.Second)
	// Préparation du paquet indiquant le début de la partie
	var toSend = map[string]string{
		"action": "launchParty",
	}
	// Envoi du paquet à son successeur
	peer.SendMessageToSuc(toSend)
}
/* Vérifie si le player est un Loup-garou */
func isLoupGarou() bool {
	return myPlayer.Role == "Loup-garou"
}

/* fonction qui décrit le déroulement d'une nuit */
func nuit() {
	// On remet tous les votes à 0
	resetVotes()
	// On arrête tous les scans de la console
	utils.StopScan()
	isNight = true
	// On remet la console à 0
	utils.ClearConsole()
	utils.PrintProgressively("La nuit tombe sur le village...")
	time.Sleep(time.Second)
	utils.SetBold()
	utils.SetRed()
	utils.PrintProgressively("Les loup-garou se réveillent...")
	utils.ResetColor()

	// Si le joueur est encore en vie
	if myPlayer.IsAlive {
		// S'il est loup-garou
		if isLoupGarou() {
			// On affiche son chat lui permettant de communiquer avec les autres loups-garous et de voter
			go launchLoupGarouChat()
		}
	}
	// Attente de la fin de la nuit
	for isNight {
		time.Sleep(500*time.Millisecond)
	}
	
}

/* Démare le chat pour les loups-garou */
func launchLoupGarouChat() {
	fmt.Println("Ecrivez pour parler dans le chat. Ecrivez /vote pour voter.")
	isInVote = false
	lastVote = -1
	for isNight {
		generateVotesInstance() // Génération de l'instance de votes et de chat
	}
}

/* Fonction qui détermine le temps avant que le jour ne se termine*/
func dayTimer() {
	// On laisse 60 secondes pour voter
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
	}
	// Préparation du paquet indiquant la fin du jour
	var toSend = map[string]string{
		"action": "endDay",
		"target": strconv.Itoa(GetVotesVictim()), // On envoie la victime des votes
	}
	// Envoi du message à son successeur
	peer.SendMessageToSuc(toSend)
}

/* Fonction qui décrit le déroulement d'une journée */
func jour() {
	// Remise à 0 des votes
	resetVotes()
	// On arrête tous les scans en cours
	utils.StopScan()
	lastVote = -1
	utils.ClearConsole()
	utils.PrintProgressively("Le jour se lève...")
	// Si un joueur est bien mort durant la nuit
	if victim != -1 {
		// Récupération du joueur tué et on l'indique comme étant mort
		player, ind := findPlayerById(victim)
		player.IsAlive = false

		// Si le joueur est le client actuel
		if player.Id == myPlayer.Id {
			// On s'indique comme mort
			p := myPlayer
			p.IsAlive = false
			myPlayer = p
		}

		// On affiche qui est mort durant la nuit
		currentPlayers[ind] = player
		utils.SetBold()
		utils.SetRed()
		utils.PrintProgressively(player.Pseudo + " est mort durant la nuit !")
		utils.ResetColor()
	}

	// Si les loups-garous n'ont pas gagné après la nuit
	if verifVictoire() == false {
		// Si je suis le host
		if peer.AmIHost() {
			// On lance le timer du jour
			go dayTimer()
		}

		// Si le client est encore en vie
		if myPlayer.IsAlive {
			fmt.Println("Ecrivez pour parler dans le chat. Ecrivez /vote pour voter.")
			isInVote = false
			// Tant que c'est encore le jour
			go func() {
				for isNight == false {
					generateVotesInstance() // Génération de l'instance de vote et chat
				}
			}()
			
		} else { // Sinon
			// Tant que c'est encore le jour
			go func() {
				for isNight == false {
					utils.LaunchScan() // Lancement d'un scan pour éviter que ce qu'écrit le joueur ne soit exécuté à la fin du programme
				}
			}()
		}

		for isNight == false {
			time.Sleep(500*time.Millisecond)
		}

	} else {
		partyEnd()
	}
}

/* Vérify si il y a une victoire soit de la part des loup-garou soit des villageois */
func verifVictoire() bool {
	return (GetLoupsCount() == 0 || GetVillagersCount() == 0)
}

/* Fonction permettant la succesion de la nuit et du jour tant que la partie n'est pas fini */
func Partie() {
	for verifVictoire() == false {
		nuit()
		jour()
	}
}

/* Affiche qui a gangné une fois que la partie est fini ainsi que la liste des loup-garou */
func partyEnd() {
	if GetLoupsCount() == 0 {
		utils.SetBold()
		utils.SetGreen()
		fmt.Println("Les villageois ont gagné !")
		utils.ResetColor()
	} else if GetVillagersCount() == 0 {
		utils.SetBold()
		utils.SetRed()
		fmt.Println("Les loups-garou ont gagné !")
		utils.ResetColor()
	}
	utils.SetBold()
	utils.SetRed()
	fmt.Println("\n")
	fmt.Println("Les loups-garou étaient :")
	for _, p := range currentPlayers {
		if p.Role == "Loup-garou" {
			fmt.Println("- ", p.Pseudo)
		}
	}
	utils.ResetColor()
	time.Sleep(3 * time.Second)
	os.Exit(1)
}

/* Fonction qui gère le système de vote */
func generateVotesInstance() {
	// Récupération du message écrit par le client dans la console
	m := utils.LaunchScan()
	// Si le client n'avait pas fait /vote dans la console précédemment
	if isInVote == false {
		// S'il vient d'effectuer la commande /vote
		if m == "/vote" {
			// On décrit qu'il est entrain de voter
			isInVote = true
			// On remet la console à 0
			utils.ClearConsole()

			// Pour chaque joueur connecté
			for i := 0; i < len(currentPlayers); i++ {
				// Récupération du joueur
				player := currentPlayers[i]
				// Si le joueur est en vie
				if player.IsAlive {

					if i == 0 {
						fmt.Println("+-+-+-+-+-+-+-+-+-+-+-+-+")
					}

					// Si c'est la nuit et que le joueur est aussi loup-garou
					if player.Role == "Loup-garou" && isNight {
						// Affichage de son pseudo en rouge
						utils.SetBold()
						utils.SetRed()
					}

					// Si le joueur correspond au client
					if player.Id == myPlayer.Id {
						// Affichage du pseudo en vert
						utils.ResetColor()
						utils.SetBold()
						utils.SetGreen()
					}

					// Affichage du joueur et ses votes actuels
					fmt.Printf("%d\t%s (%d votes)\n", player.Id, player.Pseudo, player.Votes)
					utils.ResetColor()
					fmt.Println("+-+-+-+-+-+-+-+-+-+-+-+-+")
				}

			}
			utils.SetBold()
			utils.SetRed()
			fmt.Println("\n\nEntrez l'id du myPlayer que vous voulez éliminer.")
			utils.ResetColor()

		} else { // Si le joueur a écrit dans le chat
			channel := "all"

			// S'il s'agit de la nuit
			if isNight { 
				// On définit le channel comme étant celui des loups-garous
				channel = "Loup-garou"
			}

			// Préparation du paquet indiquant un nouveau message dans le chat
			var toSend = map[string]string{
				"action":  "chat",
				"channel": channel, // Définition du channel
				"message": m, // Définition du message
			}

			// Envoi du paquet à son successeur
			peer.SendMessageToSuc(toSend)
		}
	} else { // Si on avait effectué la commande /vote avant
		// Récupération de l'id du joueur voté
		targetId, err := strconv.Atoi(m)

		// Si l'id indiqué n'est pas un nombre
		if err != nil {
			utils.SetRed()
			fmt.Println("L'id indiqué n'est pas valable.")
			utils.ResetColor()
		} else { // Sinon
			utils.SetBold()
			utils.SetGreen()
			// On regarde s'il existe bien un joueur pour cet id
			player, _ := findPlayerById(targetId)
			// Si le joueur voté est en vie
			if player.IsAlive {
				// On décrit que nous ne sommes plus entrain de voter
				isInVote = false  
				utils.ClearConsole()
				fmt.Println("Vous avez voté pour eliminer :", player.Pseudo)
				utils.ResetColor()
				fmt.Println("Ecrivez pour parler dans le chat. Ecrivez /vote pour voter.")

				// Préparation du paquet annonçant le vote
				var toSend = map[string]string{
					"action":        "vote",
					"targetId":      strconv.Itoa(targetId), // On décrit le joueur voté
					"removeVoteFor": strconv.Itoa(lastVote), // On indique pour quel joueur nous avions voté auparavant pour retirer notre vote
				}
				// L'ancien joueur voté devient celui pour qui nous venons de voter
				lastVote = targetId
				// Envoi du message au successeur
				peer.SendMessageToSuc(toSend)
			} else { // Sinon (joueur mort ou inexistant)
				// On indique que l'id n'est pas valide
				utils.SetRed()
				fmt.Println("L'id indiqué n'est pas valable.")
				utils.ResetColor()
			}

		}
	}
}

/* Retourn un jour grace a son ID */
func findPlayerById(id int) (types.Player, int) {

	for ind, player := range currentPlayers {
		if player.Id == id {
			return player, ind
		}
	}
	return types.Player{Pseudo: "", Role: "", Id: -1}, -1
}

/* Renvoie le nombre retant de loup-garou */
func GetLoupsCount() int {
	var count int = 0
	for _, player := range currentPlayers {
		if player.Role == "Loup-garou" && player.IsAlive {
			count++
		}
	}

	return count
}

/* Renvoie le nombre retant de villageois */
func GetVillagersCount() int {
	var count int = 0
	for _, player := range currentPlayers {
		if player.Role != "Loup-garou" && player.IsAlive {
			count++
		}
	}

	return count
}

/* Fonction qui récupère la victim du vote */
func GetVotesVictim() int {
	target := -1
	maxVotes := 0

	for _, player := range currentPlayers {
		if player.IsAlive && player.Votes > maxVotes {
			maxVotes = player.Votes
			target = player.Id
		}
	}

	return target
}

/* Fonction qui remet les votes à 0 */
func resetVotes() {
	for ind, player := range currentPlayers {
		player.Votes = 0
		currentPlayers[ind] = player
	}
}

/* Fonction qui return vrai si un client est déjà contenu dans la liste des joueurs */
func contains(s []types.Player, e int) bool {
	for _, a := range s {
		if a.Id == e {
			return true
		}
	}
	return false
}
