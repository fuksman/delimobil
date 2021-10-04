package delimobil

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
)

type Employee struct {
	ID         int    `json:"id"`
	ClientID   int    `json:"client_id"`
	Phone      string `json:"phone"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	FatherName string `json:"father_name"`
	Email      string `json:"email"`
	Limit      int    `json:"limit"`
	Balance    int    `json:"balance"`
	Depart     string `json:"depart"`
	Status     int    `json:"status"`
	Active     int    `json:"active"`
}

type Employees []Employee

func (company *Company) SetEmployees() error {
	endpoint := apihost + b2bhandler + company.ID() + "/clients"
	body, err := MakeAPIRequest("GET", endpoint, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return err
	}

	var temp struct {
		Employees Employees `json:"message"`
		Success   bool      `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if temp.Success {
		company.Employees = temp.Employees
		return nil
	} else {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return err
	}
}

func (company *Company) HasEmployee(phone string) (bool, error) {
	employee, err := company.FindEmployeeByPhone(phone)
	if employee == nil {
		return false, err
	}
	return true, err
}

func (company *Company) FindEmployeeByPhone(phone string) (*Employee, error) {
	if err := company.SetEmployees(); err != nil {
		log.Print(err)
		return nil, err
	}
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	phone = reg.ReplaceAllString(phone, "")
	if len(phone) > 10 {
		phone = phone[len(phone)-10:]
	}
	for _, employee := range company.Employees {
		empPhone := employee.Phone
		if len(empPhone) > 10 {
			empPhone = empPhone[len(empPhone)-10:]
		}
		if empPhone == phone {
			return &employee, nil
		}
	}
	return nil, nil
}
