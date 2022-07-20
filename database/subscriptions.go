package database

import "github.com/sirupsen/logrus"

type Subscription struct {
	ChannelsID   []int64
	ExpedienteID string
}

func (db *PostgresService) ListSubscriptions() ([]*Subscription, error) {
	rows, err := db.client.Query(`
	SELECT channel_id, expediente_id
	FROM channel_subscription
	ORDER BY expediente_id ASC
	`)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to fetch subscriptions")
		return nil, err
	}
	defer rows.Close()
	subs := make([]*Subscription, 0, 0)
	for rows.Next() {
		var channelID int64
		var expedienteID string
		err = rows.Scan(&channelID, &expedienteID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("failed to fetch subscription row")
			return nil, err
		}
		if len(subs) == 0 || subs[len(subs)-1].ExpedienteID != expedienteID {
			sub := Subscription{
				ChannelsID:   []int64{channelID},
				ExpedienteID: expedienteID,
			}
			subs = append(subs, &sub)
		} else {
			subs[len(subs)-1].ChannelsID = append(subs[len(subs)-1].ChannelsID, channelID)
		}

	}

	return subs, nil
}
