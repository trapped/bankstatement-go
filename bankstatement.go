// package bankstatement provides an interface for parsing monthly bank statements.
package bankstatement

import (
	"fmt"
	"io"
	"time"
)

// Bank is a bank identifier.
type Bank string

// Format is either a file format or a combination of file format and reader type.
type Format string

const (
	// BankBBVA represents Banco Bilbao Vizcaya Argentaria.
	BankBBVA = "bbva"

	// PDF represents an original, generated PDF (not scanned).
	FormatPDF = "pdf"
)

// Metadata is optional metadata that might or might not be included in a monthly
// bank account report/extract.
type Metadata struct {
	// Date when the report was generated.
	ReportDate time.Time
	// IBAN of the account the report was generated for.
	IBAN string
	// Name of the account's holder.
	HolderName string
}

// Transaction is a single transaction executed on a bank account.
type Transaction struct {
	// Date when the transaction was requested.
	Date time.Time
	// Date when the transaction was executed and funds transferred.
	TransactionDate time.Time
	// Subject of the transaction.
	Subject string
	// Details of the transaction.
	Details string
	// Amount transferred.
	Amount float64
}

// Reader is a merged interface of io.Reader, io.ReaderAt and io.Seeker.
type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// ReaderSize is like Reader but also provides a Len method.
type ReaderSize interface {
	Reader
	Len() int
}

// TransactionReader is the interface for bank-specific file format readers.
type TransactionReader interface {
	// Read extracts optional Metadata and a list of Transactions from a ReaderAt.
	Read(Reader) (Metadata, []Transaction, error)
}

// NewTransactionReader returns an appropriate TransactionReader for the specified
// bank statement Bank and Format.
func NewTransactionReader(bank Bank, format Format) (TransactionReader, error) {
	switch {
	case bank == BankBBVA && format == FormatPDF:
		return new(BBVAPDFTransactionReader), nil
	default:
		return nil, fmt.Errorf("unsupported bank/format combination: %s/%s", bank, format)
	}
}
