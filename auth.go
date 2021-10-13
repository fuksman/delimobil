package delimobil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt"
)

type Auth struct {
	Login    string
	Password string
	Token    string
	Company  []struct { // Comes from JWT after parsing
		ID        float64 `json:"company_id"`
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
	} `json:"user"`
	jwt.StandardClaims
}

func (auth *Auth) IsValid() bool {
	return (auth.Valid() == nil) && auth.ExpiresAt != 0
}

func (auth *Auth) RetrieveToken() error {
	endpoint := apihost + "/b2b/auth"
	userData := map[string]string{
		"login":    auth.Login,
		"password": auth.Password,
	}
	jsonUser, err := json.Marshal(userData)
	if err != nil {
		log.Print(err)
		return err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonUser))
	if err != nil {
		log.Print(err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return err
	}

	if resp.StatusCode > 299 {
		err := errors.New("bad request, status code: " + strconv.Itoa(resp.StatusCode))
		log.Print(err)
		return err
	}

	var temp struct {
		Token   string `json:"message"`
		Success bool   `json:"success"`
	}

	if json.Unmarshal(body, &temp); !temp.Success {
		err := errors.New("can't retrieve information via API")
		log.Print(err)
		return err
	}

	token, _, err := new(jwt.Parser).ParseUnverified(temp.Token, &Auth{})
	if err != nil {
		log.Print(err)
		return err
	}

	if claims, ok := token.Claims.(*Auth); ok {
		claims.Token = token.Raw
		*auth = *claims
		return nil
	} else {
		err := errors.New("JWT is not valid")
		log.Print(err)
		return err
	}
}
