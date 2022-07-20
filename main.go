package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/seppo0010/juscaba-to-telegram/database"
	"github.com/seppo0010/libjuscaba"
	"github.com/sirupsen/logrus"
)

const MAX_NOTIFICATIONS = 4

func sendActuacionDocumento(bot *tgbotapi.BotAPI, exp *libjuscaba.Ficha, act *libjuscaba.Actuacion, doc *libjuscaba.Documento, channelID int64) error {
	res, err := http.Get(doc.URL)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"url":   doc.URL,
		}).Warn("Failed to get url")
		return err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"url":   doc.URL,
		}).Error("Failed to read url data")
		return err
	}

	msg := tgbotapi.NewDocument(channelID, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("%v: %v - %v.pdf", act.Firmantes, act.Titulo, doc.Nombre),
		Bytes: content,
	})
	_, err = bot.Send(msg)
	return err
}

func sendActuacion(bot *tgbotapi.BotAPI, exp *libjuscaba.Ficha, act *libjuscaba.Actuacion, channelID int64) error {
	msg := tgbotapi.NewMessage(channelID,
		fmt.Sprintf("%v: %v", act.Firmantes, act.Titulo),
	)
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}
	documentos, err := libjuscaba.FetchDocumentos(exp, act)
	if err != nil {
		return err
	}
	for _, doc := range documentos {
		_ = sendActuacionDocumento(bot, exp, act, doc, channelID)
	}
	return nil
}

func createBot() *tgbotapi.BotAPI {
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
	return bot
}

func createDatabase() *database.PostgresService {
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
	return db
}

func notifyExpedienteUpdates(bot *tgbotapi.BotAPI, db *database.PostgresService, sub *database.Subscription) error {
	exp, err := libjuscaba.GetExpediente(sub.ExpedienteID)
	if err != nil {
		return err
	}
	err = db.AddExpediente(exp)
	if err != nil {
		return err
	}
	actuaciones, err := exp.GetActuaciones()
	if err != nil {
		return err
	}
	notified := 0
	for i := range actuaciones {
		act := actuaciones[len(actuaciones)-i-1]
		exists, err := db.HasActuacion(exp, act)
		if err != nil {
			return err
		}
		if !exists {
			notified++
			for _, channelID := range sub.ChannelsID {
				if notified < MAX_NOTIFICATIONS {
					err = sendActuacion(bot, exp, act, channelID)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"error": err.Error(),
						}).Error("failed to post to telegram")
						return err
					}
				} else if notified == MAX_NOTIFICATIONS {
					msg := tgbotapi.NewMessage(channelID, "(y más...)")
					_, err = bot.Send(msg)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"error": err.Error(),
						}).Error("failed to post to telegram")
						return err
					}
				}
			}

			logrus.WithFields(logrus.Fields{
				"exp": exp,
				"act": act,
			}).Info("new actuacion")
			err = db.AddActuacion(exp, act)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	bot := createBot()
	db := createDatabase()
	subs, err := db.ListSubscriptions()
	if err != nil {
		os.Exit(1)
	}
	for {
		for _, sub := range subs {
			_ = notifyExpedienteUpdates(bot, db, sub)
		}
		time.Sleep(5 * time.Minute)
	}
}
