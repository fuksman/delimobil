package delimobil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"time"
)

type File struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	Status      string `json:"status"`
	URL         string `json:"url"`
	Data        io.Reader
	FileName    string
	MIME        string
}

func (company *Company) CreateInvoice(amount float64) (invoice *File, err error) {
	if !company.CanCreateInvoice {
		err := errors.New("user is not allowed to create invoices")
		log.Print(err)
		return nil, err
	}
	if amount < company.MinInvoiceAmount {
		err := errors.New("can't crate invoice with this amount, minimim amount is " +
			strconv.FormatFloat(company.MinInvoiceAmount, 'f', 2, 64))
		log.Print(err)
		return nil, err
	}
	var invoiceData struct {
		Invoice struct {
			Amount      float64 `json:"amount"`
			Bill_number int     `json:"bill_number"`
			Description string  `json:"description"`
			Created_at  string  `json:"created_at"`
		} `json:"invoice"`
	}

	invoiceData.Invoice.Amount = amount
	invoiceData.Invoice.Bill_number = company.Id
	invoiceData.Invoice.Description = "Счет"
	invoiceData.Invoice.Created_at = time.Now().Format("2006-01-02T15:04:05.000Z")

	jsonInvoice, err := json.Marshal(invoiceData)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	endpoint := apihost + b2bhandler + company.ID() + "/invoice/new"
	body, err := MakeAPIRequest("POST", endpoint, bytes.NewBuffer(jsonInvoice), &company.Auth)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var temp struct {
		Message float64 `json:"message"`
		Success bool    `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if temp.Success {
		return company.LastInvoice()
	} else {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return nil, err
	}
}

func (company *Company) LastInvoice() (invoice *File, err error) {
	month := time.Now().Month()
	year := time.Now().Year()
	endpoint := apihost + b2bhandler + company.ID() + "/docs/" + strconv.Itoa(year) + "/" + strconv.Itoa((int)(month))
	body, err := MakeAPIRequest("GET", endpoint, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var temp struct {
		Files   []File `json:"message"`
		Success bool   `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if !temp.Success {
		log.Print(err)
		return nil, err
	}

	if len(temp.Files) == 0 {
		err := errors.New("no files in this month: " + time.Month(month).String() + " " + strconv.Itoa(year))
		log.Print(err)
		return nil, err
	}

	invoice = &temp.Files[0]
	for _, file := range temp.Files {
		if file.URL > invoice.URL {
			invoice = &file
		}
	}
	body, err = MakeAPIRequest("GET", apihost+invoice.URL, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	invoice.Data = bytes.NewReader(body)
	invoice.FileName = invoice.Title + ".pdf"
	invoice.MIME = "application/pdf"
	return invoice, nil
}
