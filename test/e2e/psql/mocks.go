package psql

import (
	"fmt"
	"sync"
)

const ReservedAccountID = "000000"

type MockBank struct {
	Name          string
	Accounts      map[string]int
	TransferError []func(from string, to string, amount int) error
	sync.Mutex
}

type MockPaymentDB struct {
	Banks         map[string]*MockBank
	TransferError []func(from string, to string, amount int) error
	sync.Mutex
}

func NewMockPaymentDB() *MockPaymentDB {
	return &MockPaymentDB{
		Banks:         make(map[string]*MockBank),
		TransferError: make([]func(from string, to string, amount int) error, 0),
		Mutex:         sync.Mutex{},
	}
}

func (m *MockPaymentDB) AddBank(bank *MockBank) error {
	if _, ok := m.Banks[bank.Name]; ok {
		return fmt.Errorf("bank %s already exists", bank.Name)
	}
	m.Banks[bank.Name] = bank
	return nil
}

func (m *MockPaymentDB) InjectTransferError(fn func(from string, to string, amount int) error) {
	m.TransferError = append(m.TransferError, fn)
}

func (m *MockPaymentDB) Transfer(from string, to string, amount int) error {
	m.Lock()
	defer m.Unlock()
	for _, fn := range m.TransferError {
		if err := fn(from, to, amount); err != nil {
			return err
		}
	}
	fromBank, ok := m.Banks[from]
	if !ok {
		return fmt.Errorf("bank %s not found", from)
	}
	fromBank.Lock()
	defer fromBank.Unlock()
	toBank, ok := m.Banks[to]
	if !ok {
		return fmt.Errorf("bank %s not found", to)
	}
	toBank.Lock()
	defer toBank.Unlock()
	if fromBank.Accounts[ReservedAccountID] < amount {
		return fmt.Errorf("from balance %d < %d", fromBank.Accounts[ReservedAccountID], amount)
	}
	fromBank.Accounts[ReservedAccountID] -= amount
	toBank.Accounts[ReservedAccountID] += amount
	return nil
}

func NewMockBank(name string) *MockBank {
	b := &MockBank{
		Name:          name,
		Accounts:      make(map[string]int),
		TransferError: make([]func(from string, to string, amount int) error, 0),
		Mutex:         sync.Mutex{},
	}
	b.Accounts[ReservedAccountID] = 0
	return b
}

func (m *MockBank) AddAccount(name string, amount int) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Accounts[name]; ok {
		return fmt.Errorf("account %s already exists", name)
	}
	m.Accounts[name] = amount
	return nil
}

func (m *MockBank) InjectTransferError(fn func(from string, to string, amount int) error) {
	m.TransferError = append(m.TransferError, fn)
}

func (m *MockBank) Transfer(from string, to string, amount int) error {
	m.Lock()
	defer m.Unlock()
	for _, fn := range m.TransferError {
		if err := fn(from, to, amount); err != nil {
			return err
		}
	}
	if from == to {
		return fmt.Errorf("cannot transfer from %s to %s", from, to)
	}
	fromAcc, ok := m.Accounts[from]
	if !ok {
		return fmt.Errorf("account %s does not exist", from)
	}
	_, ok = m.Accounts[to]
	if !ok {
		return fmt.Errorf("account %s does not exist", to)
	}
	if fromAcc < amount {
		return fmt.Errorf("from balance %d < %d", fromAcc, amount)
	}
	m.Accounts[from] -= amount
	m.Accounts[to] += amount
	return nil
}
