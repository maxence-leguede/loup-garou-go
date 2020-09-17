package utils

import (
	"math/rand"
	"time"
)

/*
*   Map associant le rôle à sa description
*/
var roles = map[string]string{
	"Voleur":     "Son objectif n'est pas fixe : il peut choisir son rôle parmi les deux cartes qui n'ont pas encore été distribuées.",
	"Villageois": "Son objectif est d'éliminer tous les Loups-Garous.",
	"Cupidon":    "Son objectif est d'éliminer tous les Loups-Garous. Dès le début de la partie, il doit former un couple de deux joueurs. Leur objectif sera de survivre ensemble, car si l'un d'eux meurt, l'autre se suicidera.",
	"Voyante":    "Son objectif est d'éliminer tous les Loups-Garous. Chaque nuit, elle peut espionner un joueur et découvrir sa véritable identité...",
	"Loup-garou": "Son objectif est d'éliminer tous les innocents (ceux qui ne sont pas Loups-Garous). Chaque nuit, il se réunit avec ses compères Loups pour décider d'une victime à éliminer...",
}

/*
*   Retourne la description associée au rôle passé en paramètre
*/
func GetRoleDescription(roleName string) string {
	return roles[roleName]
}

/*
*   Distribue n rôles aléatoirements ou n est le nombre de joueurs passé en paramètre
*/
func GenerateRoles(playerCount int) map[int]string {
	takenRoles := map[int]string{}
    for i := 0; i < playerCount; i++ {
        if i%4 == 0 { // Pour être sûr d'avoir au moins un loup-garou, puis d'en avoir un tous les 3 joueurs
            takenRoles[i] = "Loup-garou"
        } else {
            takenRoles[i] = "Villageois"
        }
    }
    rand.Seed(time.Now().UnixNano()) // On initialise le random avec le temps actuel

    var a [] int
    for i,_ := range takenRoles{
        a = append(a, i)    // on créé un tableau de clés des rôles qu'on a tirés
    }
    rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] }) // on mélange le tableau pour l'aléatoire
    returnedRoles := map[int]string{}
    for i,v := range a{
        returnedRoles[i] = takenRoles[v]
    }
    return returnedRoles
}
