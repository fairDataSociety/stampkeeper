package mock

type Bot struct{}

func (b Bot) Notify(string) error {
	return nil
}

func (b Bot) List() error {
	return nil
}

func (b Bot) Watch(_, _, _, _, _, _ string) error {
	return nil
}

func (b Bot) Unwatch(string) error {
	return nil
}
