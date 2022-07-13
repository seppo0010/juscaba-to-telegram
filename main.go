package main

import (
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/seppo0010/juscaba-to-telegram/database"
	"github.com/seppo0010/libjuscaba"
	"github.com/sirupsen/logrus"
)

func main() {
	telegramToken, err := os.ReadFile("/run/secrets/telegram-token")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to read telegram token")
		os.Exit(1)
	}
	bot, err := tgbotapi.NewBotAPI(strings.TrimSpace(string(telegramToken)))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to start telegram bot")
		os.Exit(1)
	}

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
		exists, err := db.HasActuacion(exp, act)
		if err != nil {
			os.Exit(1)
		}
		if !exists {
			msg := tgbotapi.NewMessage(-524601465, act.Titulo)
			bot.Send(msg)

			logrus.WithFields(logrus.Fields{
				"exp": exp,
				"act": act,
			}).Info("new actuacion")
			err = db.AddActuacion(exp, act)
			if err != nil {
				os.Exit(1)
			}
		}
	}
}
