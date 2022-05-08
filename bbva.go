package bankstatement

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/pkg/errors"
)

// BBVAPDFTransactionReader extracts metadata and transactions from a BBVA monthly extract PDF.
type BBVAPDFTransactionReader struct {
}

// parseReportDate parses a full DD/MM/YYYY date.
func (tr *BBVAPDFTransactionReader) parseReportDate(v string) (time.Time, error) {
	return time.Parse("02/01/2006", v)
}

// parseTransactionDate parses a partial DD/MM date, and requires passing in a report date.
// BBVA reports are generated on the first day of the month and contain transactions from the
// previous month: a 2018-02 report would include transactions from 2018-01 to 2018-31.
func (tr *BBVAPDFTransactionReader) parseTransactionDate(reportDate time.Time, v string) (time.Time, error) {
	t, err := time.Parse("02/01", v)
	if err != nil {
		return t, err
	}
	reportingDate := reportDate.AddDate(0, -1, 0)
	return t.AddDate(reportingDate.Year(), 0, 0), nil
}

// parseAmount parses an EU-bank-format float64, which uses thousands-dot notation
// and commas for decimals.
func (tr *BBVAPDFTransactionReader) parseAmount(v string) (float64, error) {
	return strconv.ParseFloat(strings.NewReplacer(".", "", ",", ".").Replace(v), 64)
}

func (tr *BBVAPDFTransactionReader) parseLine(words []pdf.Text) string {
	ws := make([]string, 0)
	for _, word := range words {
		ws = append(ws, word.S)
	}
	return strings.Join(ws, " ")
}

// Read tries reading a BBVA monthly PDF bank statement from a Reader. If the reader is
// not buffered or doesn't support seeking/ReadAt, it will try fully buffering it in memory.
func (tr *BBVAPDFTransactionReader) Read(src Reader) (m Metadata, txs []Transaction, err error) {
	var r ReaderSize
	// if the passed in reader doesn't support the Len() method, which we need to
	// read the PDF data, try fully buffering it in memory
	if rs, ok := src.(ReaderSize); !ok {
		data, err := ioutil.ReadAll(src)
		if err != nil {
			return m, nil, err
		}
		r = bytes.NewReader(data)
	} else {
		r = rs
	}
	p, err := pdf.NewReader(r, int64(r.Len()))
	if err != nil {
		return m, nil, err
	}
	for i := 0; i < p.NumPage(); i++ {
		page := p.Page(i + 1)
		if page.V.IsNull() {
			continue
		}
		rows, err := page.GetTextByRow()
		if err != nil {
			return m, nil, err
		}
		if i == 0 {
			reportDateRow := rows[2].Content
			ibanRow := rows[3].Content
			holderRow := rows[4].Content
			reportDate, err := tr.parseReportDate(reportDateRow[len(reportDateRow)-1].S)
			if err != nil {
				return m, nil, errors.Wrap(err, "error parsing report date")
			}
			m.ReportDate = reportDate
			m.IBAN = tr.parseLine(ibanRow[1:7])
			m.HolderName = tr.parseLine(holderRow[1:])
		}
		for j := 7; j < len(rows)-5; j += 2 {
			transaction := rows[j]
			details := rows[j+1]
			txDate, err := tr.parseTransactionDate(m.ReportDate, transaction.Content[0].S)
			if err != nil {
				return m, nil, errors.Wrapf(err, "error parsing transaction date at page %d row %d", i, j-7)
			}
			date, err := tr.parseTransactionDate(m.ReportDate, transaction.Content[1].S)
			if err != nil {
				return m, nil, errors.Wrapf(err, "error parsing date at page %d row %d", i, j-7)
			}
			amount, err := tr.parseAmount(transaction.Content[len(transaction.Content)-2].S)
			if err != nil {
				return m, nil, errors.Wrapf(err, "error parsing transaction amount at page %d row %d", i, j-7)
			}
			txs = append(txs, Transaction{
				TransactionDate: txDate,
				Date:            date,
				Amount:          amount,
				Subject:         tr.parseLine(transaction.Content[2 : len(transaction.Content)-2]),
				Details:         tr.parseLine(details.Content),
			})
		}
	}
	return m, txs, nil
}
