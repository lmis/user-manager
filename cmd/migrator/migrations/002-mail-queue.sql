-- +migrate Up
CREATE TYPE EMAIL_STATUS AS ENUM ('PENDING', 'SENT', 'ERROR');
CREATE TABLE mail_queue (
	mail_queue_id               BIGSERIAL                    PRIMARY KEY,
	from_address                TEXT                         NOT NULL,
	to_address                  TEXT                         NOT NULL,
	subject                     TEXT                         NOT NULL,
	content                     TEXT                         NOT NULL,
	status                      EMAIL_STATUS                 NOT NULL,
	number_of_failed_attempts   SMALLINT                     NOT NULL DEFAULT 0,
	priority                    SMALLINT                     NOT NULL,
	created_at                  TIMESTAMP WITH TIME ZONE     NOT NULL DEFAULT NOW(),
	updated_at                  TIMESTAMP WITH TIME ZONE     NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE mail_queue;
DROP TYPE EMAIL_STATUS;