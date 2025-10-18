package email

type DummyEmailProvider struct {
}

func NewDummyEmailProvider() *DummyEmailProvider {
	return &DummyEmailProvider{}
}

func (cur *DummyEmailProvider) SendTo(target string, subject string, content string) error {
	return nil
}
