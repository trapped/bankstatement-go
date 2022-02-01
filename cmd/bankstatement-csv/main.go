package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/trapped/bankstatement"
)

var (
	flagDecoder = flag.String("d", "", "Wrap the input stream with a decoder")
	flagBank    = flag.String("b", bankstatement.BankBBVA, "Bank to try reading for")
	flagFormat  = flag.String("f", bankstatement.FormatPDF, "Bank statement file format")
	flagOutdir  = flag.String("o", ".", "Output directory")
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

	replacer := strings.NewReplacer(" ", "", ".", "", "_", "-")
	filename := fmt.Sprintf("%s_%s_%s.csv",
		replacer.Replace(m.HolderName),
		replacer.Replace(m.IBAN),
		m.ReportDate.Format("2006-01-02"))

	w, err := os.OpenFile(filepath.Join(*flagOutdir, filename), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening output file: %s", err)
		os.Exit(1)
	}

	cw := csv.NewWriter(w)
	for _, tx := range txs {
		cw.Write([]string{
			tx.TransactionDate.Format("2006/01/02"),
			tx.Date.Format("2006/01/02"),
			tx.Subject,
			tx.Details,
			fmt.Sprintf("%.2f", tx.Amount),
		})
	}
	cw.Flush()
}
