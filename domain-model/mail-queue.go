package domain_model

type MailQueuePriority int16

const (
	MAIL_QUEUE_PRIO_LOW  MailQueuePriority = 10
	MAIL_QUEUE_PRIO_MID  MailQueuePriority = 100
	MAIL_QUEUE_PRIO_HIGH MailQueuePriority = 1000
)
