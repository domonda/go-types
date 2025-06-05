package bank

import (
	"time"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
)

type CAMT53 struct {
	MessageID            string          `xml:"BkToCstmrStmt>GrpHdr>MsgId"`
	Created              time.Time       `xml:"BkToCstmrStmt>GrpHdr>CreDtTm"`
	StatementID          string          `xml:"BkToCstmrStmt>Stmt>Id"`
	ElectronicSequenceNr string          `xml:"BkToCstmrStmt>Stmt>ElctrncSeqNb,omitempty"`
	LegalSequenceNr      string          `xml:"BkToCstmrStmt>Stmt>LglSeqNb,omitempty"`
	FromDate             time.Time       `xml:"BkToCstmrStmt>Stmt>FrToDt>FrDtTm"`
	ToDate               time.Time       `xml:"BkToCstmrStmt>Stmt>FrToDt>ToDtTm"`
	IBAN                 IBAN            `xml:"BkToCstmrStmt>Stmt>Acct>Id>IBAN"`
	Currency             money.Currency  `xml:"BkToCstmrStmt>Stmt>Acct>Ccy"`
	BankName             string          `xml:"BkToCstmrStmt>Stmt>Acct>Svcr>FinInstnId>Nm,omitempty"`
	BIC                  BIC             `xml:"BkToCstmrStmt>Stmt>Acct>Svcr>FinInstnId>BIC"`
	Balance              []CAMT53Balance `xml:"BkToCstmrStmt>Stmt>Bal"`
	Entries              []CAMT53Entry   `xml:"BkToCstmrStmt>Stmt>Ntry"`
}

type CAMT53Amount struct {
	Amount   money.Amount   `xml:",chardata"`
	Currency money.Currency `xml:"Ccy,attr"`
}

type CAMT53Balance struct {
	Type          string       `xml:"Tp>CdOrPrtry>Cd"` // PRCD: Endsaldo gebucht vorheriger Auszug   "MSIN" "CNFA" "DNFA" "CINV" "CREN" "DEBN" "HIRI" "SBIN" "CMCN" "SOAC" "DISP" "BOLD" "VCHR" "AROI" "TSUT"
	Proprietary   string       `xml:"Tp>CdOrPrtry>Prtry"`
	Amount        CAMT53Amount `xml:"Amt"`
	CreditOrDebit string       `xml:"CdtDbtInd"` // Soll (DBIT) oder Haben (CRDT)
	Date          date.Date    `xml:"Dt>Dt"`
}

type CAMT53Entry struct {
	Amount        CAMT53Amount `xml:"Amt"`
	CreditOrDebit string       `xml:"CdtDbtInd"` // Soll (DBIT) oder Haben (CRDT)
	Status        string       `xml:"Sts"`       // BOOK, PDNG, INFO
	BookingDate   date.Date    `xml:"BookgDt>Dt"`
	ValueDate     date.Date    `xml:"ValDt>Dt"`
	ReferenceCode string       `xml:"AcctSvcrRef"`
	// BkTxCd
	DebitorName  string   `xml:"NtryDtls>TxDtls>RltdPties>Dbtr>Nm"`
	DebitorAddr  []string `xml:"NtryDtls>TxDtls>RltdPties>Dbtr>PstlAdr>AdrLine,omitempty"`
	DebitorIBAN  IBAN     `xml:"NtryDtls>TxDtls>RltdPties>DbtrAcct>Id>IBAN"`
	CreditorName string   `xml:"NtryDtls>TxDtls>RltdPties>Cdtr>Nm"`
	CreditorAddr []string `xml:"NtryDtls>TxDtls>RltdPties>Cdtr>PstlAdr>AdrLine,omitempty"`
	CreditorIBAN IBAN     `xml:"NtryDtls>TxDtls>RltdPties>CdtrAcct>Id>IBAN"`
	DebitorBIC   BIC      `xml:"NtryDtls>TxDtls>RltdAgts>DbtrAgt>FinInstnId>BIC"`
	CreditorBIC  BIC      `xml:"NtryDtls>TxDtls>RltdAgts>CdtrAgt>FinInstnId>BIC"`
	Reference    string   `xml:"NtryDtls>TxDtls>RmtInf>Strd>CdtrRefInf>Ref"`
}
