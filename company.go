package delimobil

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
)

type Company struct {
	Auth
	Info
	Employees
	Rides
	Files
	FilesDates
}

type Info struct {
	Id                int
	Name              string
	Balance           float64
	CanCreateInvoices bool
	MinInvoiceAmount  float64
}

func NewCompany(login, password string) *Company {
	var company Company
	company.login = login
	company.password = password
	company.Files = make(Files)
	return &company
}

func (company *Company) Authenticate() error {
	if company.IsValid() {
		return nil
	}

	err := company.RetrieveToken()
	if err == nil {
		company.Id = int(company.Auth.Company[0].ID)
	}
	return err
}

func (company *Company) ID() string {
	return strconv.Itoa(company.Id)
}

func (company *Company) SetInfo() error {
	endpoint := apihost + b2bhandler + company.ID() + "/info"
	body, err := MakeAPIRequest("GET", endpoint, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return err
	}

	var temp struct {
		Info struct {
			Id                int     `json:"id"`
			Name              string  `json:"company_name"`
			Balance           float64 `json:"total_sum"`
			CanCreateInvoices bool    `json:"isCreatingInvoicesAllowed"`
			MinInvoiceAmount  float64 `json:"minInvoiceAmount"`
		} `json:"message"`
		Success bool `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if temp.Success {
		company.Info = Info(temp.Info)
		return nil
	} else {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return err
	}
}

func (company *Company) IsBalanceOK(limit float64) bool {
	return company.Balance > limit
}
