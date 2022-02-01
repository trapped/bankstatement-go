Bank statement parser for Go
----------------------------

This Go module provides an interface for parsing bank statements into pairs of `Metadata` and `[]Transaction`.

I built this to convert PDF bank statements to CSV for very old (3+ years) transactions, which were not available to download as XLS.

So far only supports non-scanned (i.e. generated and downloaded) PDFs from BBVA Spain, but should be easily extensible.
