package db

import (
	"database/sql"
	"user-manager/util"

	migrate "github.com/rubenv/sql-migrate"
)

func MigrateUp(db *sql.DB) (int, error) {
	migrate.SetTable("migrations")
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "001-user-tables",
				Up: []string{
					`
          CREATE TYPE USER_ROLE AS ENUM ('USER', 'ADMIN');
          CREATE TABLE app_user (
            app_user_id SERIAL PRIMARY KEY,
            role USER_ROLE NOT NULL,
            email TEXT NOT NULL,
            email_verified BOOL NOT NULL,
            email_verification_token TEXT,
            password_reset_token TEXT,
            password_reset_token_valid_until TIMESTAMP WITH TIME ZONE,
            two_factor_token TEXT,
            temprary_two_factor_token TEXT,
            password_hash TEXT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL
          );

          CREATE TABLE user_session (
            user_session_id TEXT PRIMARY KEY,
            app_user_id INTEGER REFERENCES app_user(app_user_id) NOT NULL,
            timeout_at TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL
          );
          `,
				},
				Down: []string{
					`
          DROP TABLE user_session;
          DROP TABLE app_user;
          DROP TYPE USER_ROLE;
          `,
				},
			},
			{
				Id: "002-mail-queue-tables",
				Up: []string{
					`
          CREATE TYPE EMAIL_STATUS AS ENUM ('PENDING', 'SENT', 'ERROR');
          CREATE TABLE mail_queue (
            mail_queue_id SERIAL PRIMARY KEY,
            email TEXT NOT NULL,
            content TEXT NOT NULL,
            status EMAIL_STATUS NOT NULL,
            number_of_failed_attempts SMALLINT NOT NULL,
            priority SMALLINT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL
          );
          `,
				},
				Down: []string{
					`
          DROP TABLE mail_queue;
          DROP TYPE EMAIL_STATUS;
          `,
				},
			},
			{
				Id: "003-two-factor-throttling",
				Up: []string{
					`
          CREATE TABLE two_factor_throttling (
            two_factor_throttling_id SERIAL PRIMARY KEY,
				    app_user_id INTEGER UNIQUE REFERENCES app_user(app_user_id) NOT NULL,
				    failed_attempts_since_last_success INTEGER NOT NULL,
				    timeout_until TIMESTAMP WITH TIME ZONE,
				    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
          );
          `,
				},
				Down: []string{
					`
			    DROP TABLE two_facor_throwttling;
			    `,
				},
			},
		},
	}

	numApplied, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return 0, util.Wrap("MigrateUp", "issue executing migration", err)
	}
	return numApplied, nil
}
