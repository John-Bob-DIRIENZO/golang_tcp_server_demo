# Golang TCP Server Demo

Une rapide démonstration d'un serveur TCP simple écrit en Golang.

Le but ici est de faire une simple démonstration de l'architecture de base de ce genre de service, pas de montrer un
code exempt de défauts.

Le comportement de ce serveur est d'accépter des connexions TCP et de notifier tous les autres utilisateurs de
l'arrivée, du départ ou des messages des autres utilisateurs.

## Quick start

```bash
go run main.go
```

On peut ensuite ouvrir plusieurs terminaux et sur MacOS faire

```bash
nc localhost 1287
```