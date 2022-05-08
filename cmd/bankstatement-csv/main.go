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
	flagDecoder = flag.String("d", "", "Wrap the input stream with a decoder (available: multipart)")
	flagBank    = flag.String("b", bankstatement.BankBBVA, "Bank to try reading for")
	flagFormat  = flag.String("f", bankstatement.FormatPDF, "Bank statement file format")
	flagOutdir  = flag.String("o", ".", "Output directory")
	flagHeaders = flag.Bool("h", false, "Include header row")
)

func main() {
	flag.Parse()

	var r bankstatement.Reader
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf("error opening input file: %s\n", err)
		os.Exit(1)
	}

	if flagDecoder != nil {
		switch *flagDecoder {
		case "":
			break
		case "multipart":
			r, err = new(bankstatement.MIMEMultipartDecoder).Wrap(r)
		default:
			fmt.Printf("unsupported decoder: %s\n", *flagDecoder)
			flag.Usage()
			os.Exit(1)
		}
		if err != nil {
			fmt.Printf("decoder error: %s\n", err)
			os.Exit(1)
		}
	}

	txr, err := bankstatement.NewTransactionReader(
		bankstatement.Bank(*flagBank),
		bankstatement.Format(*flagFormat),
	)
	if err != nil {
		fmt.Printf("transaction reader error: %s\n", err)
		os.Exit(1)
	}

	m, txs, err := txr.Read(r)
	if err != nil {
		fmt.Printf("error reading transactions: %s\n", err)
		os.Exit(1)
	}

	replacer := strings.NewReplacer(" ", "", ".", "", "_", "-")
	filename := fmt.Sprintf("%s_%s_%s.csv",
		replacer.Replace(m.HolderName),
		replacer.Replace(m.IBAN),
		m.ReportDate.Format("2006-01-02"))

	w, err := os.OpenFile(filepath.Join(*flagOutdir, filename), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening output file: %s\n", err)
		os.Exit(1)
	}

	cw := csv.NewWriter(w)
	if *flagHeaders {
		if err := cw.Write([]string{
			"transaction_date",
			"date",
			"subject",
			"details",
			"amount",
		}); err != nil {
			fmt.Printf("error writing headers: %s\n", err)
			os.Exit(1)
		}
	}
	for _, tx := range txs {
		if err := cw.Write([]string{
			tx.TransactionDate.Format("2006/01/02"),
			tx.Date.Format("2006/01/02"),
			tx.Subject,
			tx.Details,
			fmt.Sprintf("%.2f", tx.Amount),
		}); err != nil {
			fmt.Printf("error writing transaction: %s\n", err)
			os.Exit(1)
		}
	}
	cw.Flush()
}
