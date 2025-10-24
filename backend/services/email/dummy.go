package email

import "context"

type DummyEmailProvider struct {
}

func NewDummyEmailProvider() *DummyEmailProvider {
	return &DummyEmailProvider{}
}

func (cur *DummyEmailProvider) SendTo(ctx context.Context, target string, subject string, content string) error {
	return nil
}
