package database

import "github.com/sirupsen/logrus"

type Subscription struct {
	ChannelID    int64
	ExpedienteID string
}

func (db *PostgresService) ListSubscriptions() ([]*Subscription, error) {
	rows, err := db.client.Query(`
	SELECT channel_id, expediente_id
	FROM channel_subscription
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
		var s Subscription
		err = rows.Scan(&s.ChannelID, &s.ExpedienteID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("failed to fetch subscription row")
			return nil, err
		}
		subs = append(subs, &s)
	}

	return subs, nil
}
