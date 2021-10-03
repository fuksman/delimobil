package delimobil

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
)

type Company struct {
	CompanyInfo
	Auth
}

type CompanyInfo struct {
	Id               int
	Name             string
	Balance          float64
	CanCreateInvoice bool
	MinInvoiceAmount float64
	Rides            Rides
}

func (company *Company) Authenticate(login, password string) error {
	if company.IsValid() {
		return nil
	}

	company.Login = login
	company.Password = password

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
		CompanyInfo struct {
			Id               int     `json:"id"`
			Name             string  `json:"company_name"`
			Balance          float64 `json:"total_sum"`
			CanCreateInvoice bool    `json:"isCreatingInvoicesAllowed"`
			MinInvoiceAmount float64 `json:"minInvoiceAmount"`
			Rides            Rides
		} `json:"message"`
		Success bool `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if temp.Success {
		company.CompanyInfo = CompanyInfo(temp.CompanyInfo)
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
