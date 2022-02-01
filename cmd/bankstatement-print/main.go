package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/trapped/bankstatement"
)

var (
	flagDecoder = flag.String("d", "", "Wrap the input stream with a decoder")
	flagBank    = flag.String("b", bankstatement.BankBBVA, "Bank to try reading for")
	flagFormat  = flag.String("f", bankstatement.FormatPDF, "Bank statement file format")
)

func main() {
	flag.Parse()

	var r bankstatement.Reader
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf("error opening input file: %s", err)
		os.Exit(1)
	}

	if flagDecoder != nil {
		switch *flagDecoder {
		case "":
			break
		case "multipart":
			r, err = new(bankstatement.MIMEMultipartDecoder).Wrap(r)
		default:
			fmt.Printf("unsupported decoder: %s", *flagDecoder)
			os.Exit(1)
		}
		if err != nil {
			fmt.Printf("decoder error: %s", err)
			os.Exit(1)
		}
	}

	txr, err := bankstatement.NewTransactionReader(
		bankstatement.Bank(*flagBank),
		bankstatement.Format(*flagFormat),
	)
	if err != nil {
		fmt.Printf("transaction reader error: %s", err)
		os.Exit(1)
	}

	m, txs, err := txr.Read(r)
	if err != nil {
		fmt.Printf("error reading transactions: %s", err)
		os.Exit(1)
	}
	fmt.Printf(`Metadata:
	Report date: %v
	IBAN: %v
	Holder: %v

Transactions:`+"\n\n", m.ReportDate.Format("2006/01/02"), m.IBAN, m.HolderName)
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	for _, tx := range txs {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
			tx.TransactionDate.Format("2006/01/02"),
			tx.Date.Format("2006/01/02"),
			tx.Subject,
			tx.Details,
			tx.Amount)
	}
	w.Flush()
}
