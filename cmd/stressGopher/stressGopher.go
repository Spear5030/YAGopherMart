package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/joeljunstrom/go-luhn"
	"log"
	"time"
)

func main() {
	accAdr := "http://localhost:8081"
	log.Println(accAdr)
	gmAdr := "http://localhost:8080"

	m := []byte(`
			{
				"login": "testLogin",
				"password": "testPassword"
			}
		`)
	client := resty.New()
	req := client.
		R().
		SetHeader("Content-Type", "application/json").
		SetBody(m)
	resp, err := req.Post(gmAdr + "/api/user/login")
	if err != nil {
		log.Println(err)
	}
	authHeader := resp.Header().Get("Authorization")
	if authHeader != "" {
		client.SetHeader("Authorization", authHeader)
	}
	for i := 0; i < 10000; i++ {

		go func() {
			orderNumber := luhn.Generate(8)
			fmt.Println(orderNumber)
			req = client.
				R().
				SetBody([]byte(orderNumber))
			resp, err = req.Post(gmAdr + "/api/user/orders")
			if err != nil {
				log.Println(err)
			}
			fmt.Println("Status code " + resp.Status())
		}()
	}
	time.Sleep(10 * time.Second)
}
