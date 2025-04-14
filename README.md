# API Golang avec Gin

Une API simple et efficace construite avec Go et le framework Gin.

## Prérequis

- Go 1.21 ou supérieur
- Git

## Installation

1. Cloner le repository

```bash
git clone <votre-repo>
cd pec2-backend
```

2. Installer les dépendances

```bash
go mod tidy
```

3. Lancer l'application

```bash
go run main.go
```

L'API sera disponible sur `http://localhost:8080`

## Généré la document swagger 

```bash
swag init
```

## Le Swagger se trouve à l'url suivante
```bash
http://localhost:8080/swagger/index.html
```

## Lancer les tests

```bash
go test ./handlers/auth -v
```