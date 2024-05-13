package ledger

type Ledger interface {
	NewTransaction() Transaction
	RunTxs() error
}

type ledger struct {
	transactions []Transaction
}

func New() Ledger {
	return &ledger{}
}

func (l *ledger) NewTransaction() Transaction {
	tx := &tx{}
	l.transactions = append(l.transactions, tx)

	return tx
}

func (l *ledger) RunTxs() error {
	for _, tx := range l.transactions {
		if err := tx.Run(); err != nil {
			return err
		}
	}

	return nil
}
