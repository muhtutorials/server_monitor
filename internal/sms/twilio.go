package sms

import (
	"encoding/json"
	"fmt"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"server_monitor/internal/config"
)

func SendSMSViaTwilio(app *config.AppConfig, to, msg string) error {
	accountSid := app.Preferences["twilio_sid"]
	authToken := app.Preferences["twilio_auth_token"]

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &twilioApi.CreateMessageParams{}
	//params.SetTo(to)
	//params.SetFrom(app.Preferences["twilio_phone_number"])
	params.SetFrom("+15005550006")
	params.SetTo("+15005550004")
	params.SetBody(msg)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println("Error sending SMS message: " + err.Error())
	} else {
		response, _ := json.Marshal(*resp)
		fmt.Println("Response: " + string(response))
	}

	return nil
}
