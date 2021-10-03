package delimobil

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
)

var apihost = "https://b2b-api.delitime.ru"
var b2bhandler = "/b2b/company/"

type ctxKey int

var companyKey ctxKey

func NewContext(ctx context.Context, company *Company) context.Context {
	return context.WithValue(ctx, companyKey, company)
}

func FromContext(ctx context.Context) (*Company, bool) {
	company, ok := ctx.Value(companyKey).(*Company)
	return company, ok
}

func MakeAPIRequest(method string, endpoint string, reqbody io.Reader, auth *Auth) (body []byte, err error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, endpoint, reqbody)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+auth.Token())
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if resp.StatusCode > 299 {
		err := errors.New("bad request, status code: " + strconv.Itoa(resp.StatusCode))
		log.Print(err)
		return nil, err
	}

	return body, nil
}
