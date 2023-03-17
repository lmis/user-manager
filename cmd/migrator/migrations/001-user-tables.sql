-- +migrate Up
CREATE TYPE USER_ROLE AS ENUM ('USER', 'ADMIN', 'SUPER_ADMIN');
CREATE TYPE USER_LANGUAGE AS ENUM ('EN', 'DE');
CREATE TABLE app_user (
	app_user_id                       BIGSERIAL                   PRIMARY KEY,
	language                          USER_LANGUAGE               NOT NULL,
	user_name                         TEXT                        NOT NULL,
	password_hash                     TEXT                        NOT NULL,
	email                             TEXT                        UNIQUE NOT NULL,
	email_verified                    BOOL                        NOT NULL,
	email_verification_token          TEXT                        NOT NULL DEFAULT '',
	next_email                        TEXT                        NOT NULL DEFAULT '',
	password_reset_token              TEXT                        NOT NULL DEFAULT '',
	password_reset_token_valid_until  TIMESTAMP WITH TIME ZONE,
	second_factor_token               TEXT                        NOT NULL DEFAULT '',
	temporary_second_factor_token     TEXT                        NOT NULL DEFAULT '',
	created_at                        TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT NOW(),
	updated_at                        TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT NOW()
);

CREATE TABLE app_user_role (
	app_user_role_id                  BIGSERIAL                                  PRIMARY KEY,
	app_user_id                       BIGINT REFERENCES app_user(app_user_id)    NOT NULL,
	role                              USER_ROLE                                  NOT NULL,
	created_at                        TIMESTAMP WITH TIME ZONE                   NOT NULL DEFAULT NOW(),
	deleted_at                        TIMESTAMP WITH TIME ZONE
);

CREATE TYPE USER_SESSION_TYPE AS ENUM ('LOGIN', 'SUDO', 'REMEMBER_DEVICE');
CREATE TABLE user_session (
	user_session_id                   TEXT                                       PRIMARY KEY,
	app_user_id                       BIGINT REFERENCES app_user(app_user_id)    NOT NULL,
	user_session_type                 USER_SESSION_TYPE                          NOT NULL,
	timeout_at                        TIMESTAMP WITH TIME ZONE                   NOT NULL,
	created_at                        TIMESTAMP WITH TIME ZONE                   NOT NULL DEFAULT NOW(),
	updated_at                        TIMESTAMP WITH TIME ZONE                   NOT NULL DEFAULT NOW()
);

CREATE TABLE second_factor_throttling (
	second_factor_throttling_id              BIGSERIAL                                  PRIMARY KEY,
	app_user_id                           BIGINT REFERENCES app_user(app_user_id)    UNIQUE NOT NULL,
	failed_attempts_since_last_success    INTEGER                                    NOT NULL,
	timeout_until                         TIMESTAMP WITH TIME ZONE,
	updated_at                            TIMESTAMP WITH TIME ZONE                   NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE user_session;
DROP TYPE USER_SESSION_TYPE;
DROP TABLE app_user_role;
DROP TABLE app_user;
DROP TYPE USER_ROLE;
DROP TABLE second_factor_throttling;