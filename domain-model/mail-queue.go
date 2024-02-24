package domain_model

import "go.mongodb.org/mongo-driver/bson/primitive"

type MailQueueID primitive.ObjectID
type MailQueuePriority int16
type MailStatus string

const (
	MailQueuePrioLow  MailQueuePriority = 10
	MailQueuePrioMid  MailQueuePriority = 100
	MailQueuePrioHigh MailQueuePriority = 1000
	MailStatusPending MailStatus        = "pending"
	MailStatusSent    MailStatus        = "sent"
	MailStatusFailed  MailStatus        = "failed"

	MailQueueCollectionName = "MailQueue"
)

type Mail struct {
	ID                     MailQueueID `bson:"_id,omitempty"`
	FromAddress            string
	ToAddress              string
	Content                string
	Subject                string
	Status                 MailStatus
	Priority               MailQueuePriority
	NumberOfFailedAttempts int8
}

type MailInsert struct {
	FromAddress string
	ToAddress   string
	Content     string
	Subject     string
	Priority    MailQueuePriority
}
