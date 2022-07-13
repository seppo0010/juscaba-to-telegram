package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/seppo0010/juscaba-to-telegram/database"
	"github.com/seppo0010/libjuscaba"
	"github.com/sirupsen/logrus"
)

func main() {
	dbPassword, err := os.ReadFile(os.Getenv("POSTGRES_PASSWORD_FILE"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to read postgres password")
		os.Exit(1)
	}

	db, err := database.NewPostgresService(fmt.Sprintf(
		"postgres://%v:%v@%v/%v?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		strings.TrimSpace(string(dbPassword)),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"),
	))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to connect to database")
		os.Exit(1)
	}

	exp, err := libjuscaba.GetExpediente("182908/2020-0")
	if err != nil {
		os.Exit(1)
	}
	err = db.AddExpediente(exp)
	if err != nil {
		os.Exit(1)
	}
	actuaciones, err := exp.GetActuaciones()
	if err != nil {
		os.Exit(1)
	}
	for _, act := range actuaciones {
		fmt.Printf("%#v\n", act)
		err = db.AddActuacion(exp, act)
		if err != nil {
			os.Exit(1)
		}
	}
}
