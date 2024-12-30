package config

// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go

import (
	"fmt"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendOTPMail(email string, otp string) {
	from := mail.NewEmail("LearnuimAI", "info@learniumai.com")
	subject := "OTP"
	to := mail.NewEmail("Hello", email)

	m := mail.NewV3MailInit(from, subject, to) //  content,

	m.Personalizations[0].SetDynamicTemplateData("otp", otp)
	// m.Personalizations[0].SetSubstitution("-otp-", "10028")
	m.SetTemplateID("d-178e0acf6fa74b8a8216cc96d6874b8c")

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_KEY"))
	response, err := client.Send(m)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

}

func SendPrecisionVerifyMail(email string, id string) {
	from := mail.NewEmail("LearnuimAI", "info@learniumai.com")
	subject := "Verify account"
	to := mail.NewEmail("Hello", email)

	m := mail.NewV3MailInit(from, subject, to) //  content,

	completeLink := "http://frontend.com/" + id
	m.Personalizations[0].SetDynamicTemplateData("link", completeLink)
	m.SetTemplateID("d-258ea1b6842540bda156c6242db69d2c")

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_KEY"))
	response, err := client.Send(m)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
