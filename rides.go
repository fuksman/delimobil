package delimobil

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

type Rides []Ride

type Ride struct {
	RentID            int       `json:"rent_id"`
	RentStartTime     time.Time `json:"rent_start_time"`
	RentEndTime       time.Time `json:"rent_end_time"`
	Duration          int       `json:"duration"`
	Cost              float64   `json:"cost"`
	Currency          string    `json:"currency"`
	Car               string    `json:"car"`
	VehicleNumber     string    `json:"vehicle_number"`
	ClientBio         string    `json:"client_bio"`
	Distance          int       `json:"distance"`
	StartPointAddress string    `json:"start_point_address"`
	EndPointAddress   string    `json:"end_point_address"`
}

func (company *Company) SetRides(limit, page int) error {
	endpoint := apihost + b2bhandler + company.ID() + "/transfers/all?limit=" + strconv.Itoa(limit) + "&page=" + strconv.Itoa(page)
	body, err := MakeAPIRequest("GET", endpoint, nil, &company.Auth)
	if err != nil {
		log.Print(err)
		return err
	}

	var temp struct {
		Rides   Rides `json:"message"`
		Success bool  `json:"success"`
	}
	json.Unmarshal(body, &temp)

	if temp.Success {
		company.Rides = temp.Rides
		return nil
	} else {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return err
	}
}

func (ride *Ride) String() string {
	if ride.Currency == "rub" {
		ride.Currency = "₽"
	}
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("UTC", 0)
	}
	return ride.RentStartTime.In(loc).Format("02.01.06 15:04") + "—" + ride.RentEndTime.In(loc).Format("15:04") +
		" " + strconv.Itoa(ride.Duration) + " мин., " + strconv.FormatFloat(ride.Cost, 'f', 2, 64) + " " + ride.Currency + "\n" +
		ride.StartPointAddress + " 👉 " + ride.EndPointAddress + ", " + ride.ClientBio + ", " + ride.Car + " (" + strings.ToUpper(ride.VehicleNumber) + ")"
}

func (rides Rides) String() (list string) {
	list = ""
	for _, ride := range rides {
		list += ride.String() + "\n\n"
	}
	return
}
