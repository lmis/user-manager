package domain_model

type MailQueuePriority int16

const (
	MailQueuePrioLow  MailQueuePriority = 10
	MailQueuePrioMid  MailQueuePriority = 100
	MailQueuePrioHigh MailQueuePriority = 1000
)
