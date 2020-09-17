package types

/* Structure d'un joeur */
type Player struct {
	Id      int //son id
	Pseudo  string
	Role    string 
	Votes   int //le nombre de votes qu'il y a sur lui
	IsAlive bool // vrai si il est en vie
}
