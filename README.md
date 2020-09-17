# Loup Garou GO

Jeu de loups-garous en pair-à-pair développé en GO par Baptiste BATARD, Josik SALLAUD et Maxence LEGUEDE.

## Prérequis

Go installé sur votre machine. (Voir : [Liens de téléchargement](https://golang.org/dl/))
(Testé avec la version 1.10.4)

## Installation et exécution

Depuis un terminal :
```bash
sh launch.sh
```
ou
```bash
go run Main.go
```

## Configuration

Le fichier de configuration se trouve dans utils/config.go


Il faut que les joueurs soient en accord avec le nombre de joueurs qu'il y aura. Ces derniers devront indiquer le nombre de joueurs dans le fichier de configuration.
```go
"playersToStart":"nbJoueurs"
```

Pour utiliser un port fixe ou en générer un au hasard parmis ceux disponibles, il suffit de modifier 
```go
"fixedPort":"false"
```
False pour générer un port, true pour utilisé celui indiqué dans
```go
"port":"5000"
```

Pour jouer en local modifier la configuration comme ci-dessous
```go
"useLocalIP":"true"
```
Pour jouer en distanciel (Version non stable, votre anti-virus risque de vouloir mettre le programme en quarantaine au lancement)
```go
"useLocalIP":"false"
```

## Jouer

Le loup-garou se joue à deux joueurs minimum, il n'y a pas de nombre de joueurs maximum.

Dans cette version, les rôles spéciaux ne sont pas présents. Il y a seulement les loups-garous et les villageois.
Ces rôles pourront être ajoutés en tant qu'extension.

Le premier à vouloir jouer doit lancer une partie.
Ensuite tout le monde peut se connecter à lui ou à un autre qui est déjà connecté à la partie (principe du pair-à-pair). Il faut bien sûr connaître l'IP et le port d'un joueur connecté pour rejoindre la partie (ceux-ci sont affichés une fois la partie rejointe).

Le jour, les joueurs ont 60 secondes pour voter (même s'ils sont tous d'accord pour qui voter)