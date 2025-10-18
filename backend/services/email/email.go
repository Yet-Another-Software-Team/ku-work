package email

type EmailProvider interface {
	SendTo(target string, subject string, content string) error
}
