package delimobil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sort"
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

type Files map[int]map[int][]File

type FilesDates struct {
	Invoices     []time.Time
	UPDs         []time.Time
	RentsDetails []time.Time
}

var fileTypes = [...]string{
	"invoice",
	"upd",
	"rentsDetail",
}

func (company *Company) SetFilesDates() error {
	endpoint := apihost + b2bhandler + company.ID() + "/docs"
	body, err := MakeAPIRequest("GET", endpoint, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return err
	}

	var temp struct {
		Message map[string][]struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
		} `json:"message"`
		Success bool `json:"success"`
	}
	json.Unmarshal(body, &temp)

	var invoices, upds, rentsDetails []time.Time

	for datestr, filesCounts := range temp.Message {
		for _, fileCount := range filesCounts {
			date, err := time.Parse("2006-01-02", datestr)
			if err != nil {
				return err
			}
			switch fileCount.Type {
			case "invoice":
				invoices = append(invoices, date)
			case "upd":
				upds = append(upds, date)
			case "rentsDetail":
				rentsDetails = append(rentsDetails, date)
			default:
				log.Println("New type of file")
			}
		}
	}

	if temp.Success {
		timeSort := func(slice []time.Time) func(i int, j int) bool {
			return func(i, j int) bool {
				return slice[i].After(slice[j])
			}
		}
		sort.Slice(invoices, timeSort(invoices))
		sort.Slice(upds, timeSort(upds))
		sort.Slice(rentsDetails, timeSort(rentsDetails))
		company.FilesDates.Invoices = invoices
		company.FilesDates.UPDs = upds
		company.FilesDates.RentsDetails = rentsDetails
		return nil
	} else {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return err
	}
}

func (company *Company) SetFiles(year, month int) error {
	endpoint := apihost + b2bhandler + company.ID() + "/docs/" + strconv.Itoa(year) + "/" + strconv.Itoa((int)(month))
	body, err := MakeAPIRequest("GET", endpoint, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return err
	}

	var temp struct {
		Files   []File `json:"message"`
		Success bool   `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if temp.Success {
		if company.Files[year] == nil {
			company.Files[year] = make(map[int][]File)
		}
		company.Files[year][month] = temp.Files
		return nil
	} else {
		log.Print(err)
		return err
	}
}

func (file *File) SetData(auth *Auth) error {
	body, err := MakeAPIRequest("GET", apihost+file.URL, nil, auth)
	if err != nil {
		log.Print(err)
		return err
	}
	file.Data = bytes.NewReader(body)
	file.FileName = file.Title + ".pdf"
	file.MIME = "application/pdf"
	return nil
}

func (company *Company) CreateInvoice(amount float64) (invoice *File, err error) {
	if !company.CanCreateInvoices {
		err := errors.New("it is not allowed to create invoices for " + company.Name)
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

	if !temp.Success {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return nil, err
	}

	return company.LastFileByType("invoice")
}

func (company *Company) LastFileByType(fileType string) (file *File, err error) {
	if err = company.SetFilesDates(); err != nil {
		log.Print(err)
		return nil, err
	}

	var date time.Time
	switch fileType {
	case "invoice":
		date = company.FilesDates.Invoices[0]
	case "upd":
		date = company.FilesDates.UPDs[0]
	case "rentsDetail":
		date = company.FilesDates.RentsDetails[0]
	default:
		err = errors.New("file type " + fileType + " doesn't exist")
		log.Print(err)
		return nil, err
	}

	year, month := date.Year(), int(date.Month())
	company.SetFiles(year, month)
	file = &company.Files[year][month][0]
	for _, f := range company.Files[year][month] {
		if f.Type == fileType && f.URL > file.URL {
			file = &f
		}
	}

	err = file.SetData(&company.Auth)
	return file, err
}
