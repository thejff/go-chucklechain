package ledger

type Transaction interface {
	Run() error
}

type tx struct{}

func (t *tx) Run() error {
	return nil
}

func (t *tx) verify() error {
	// Broadcast the transaction to the network and wait for results
	// Once consensus amount is reached, commit
	return nil
}

func (t *tx) commit() error {
	return nil
}
