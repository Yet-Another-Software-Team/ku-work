package email

import "context"

type EmailProvider interface {
	SendTo(ctx context.Context, target string, subject string, content string) error
}
