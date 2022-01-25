package bot

type Bot interface {
	Notify(message string) error
	List() error
	Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string) error
	Unwatch(batchId string) error
}
