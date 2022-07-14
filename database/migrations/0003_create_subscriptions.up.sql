CREATE TABLE IF NOT EXISTS channel_subscription (
  channel_id BIGINT NOT NULL,
  expediente_id VARCHAR(255) NOT NULL,
  UNIQUE (channel_id, expediente_id)
);
