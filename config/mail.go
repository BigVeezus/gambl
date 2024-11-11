package config

// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go

import (
	"fmt"
	"gambl/models"
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

func SendNewUserMail(email models.NewUserAlert) {
	from := mail.NewEmail("LearnuimAI", "info@learniumai.com")
	subject := "New User Registration"
	to := mail.NewEmail("LearnuimAI", "info@learniumai.com")

	m := mail.NewV3MailInit(from, subject, to)
	personalization := mail.NewPersonalization()
	personalization.AddTos(to)

	// Create Content with plain text and HTML
	content := []mail.Content{
		*mail.NewContent("text/plain", "A new user has registered.\n Details:\nName: "+email.First_name+" "+email.Last_name+"\nEmail: "+email.Email+"\nRoleType: "+email.User_type),
	}

	m.AddPersonalizations(personalization)
	for _, c := range content {
		m.AddContent(&c)
	}

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

func SendUserDetails(user models.User) {
	from := mail.NewEmail("LearnuimAI", "info@learniumai.com")
	subject := "New User Registration"
	to := mail.NewEmail("Learnuim-User-Alert", "info@learniumai.com")

	m := mail.NewV3MailInit(from, subject, to)
	personalization := mail.NewPersonalization()
	personalization.AddTos(to)

	// Build the content safely
	var name string
	if user.First_name != nil && user.Last_name != nil {
		name = *user.First_name + " " + *user.Last_name
	} else {
		name = "Unknown"
	}

	var email string
	if user.Email != nil {
		email = *user.Email
	} else {
		email = "No email provided"
	}

	var roleType string
	if user.User_type != nil {
		roleType = *user.User_type
	} else {
		roleType = "No role type specified"
	}

	var phone string
	if user.Phone == "" {
		phone = user.Phone
	} else {
		phone = "No phone number provided"
	}

	content := []mail.Content{
		*mail.NewContent("text/plain", fmt.Sprintf(
			"A new user has registered\nDetails:\nName: %s\nEmail: %s\nRoleType: %s\nPhoneNumber: %s",
			name, email, roleType, phone)),
	}

	// Create Content with plain text and HTML
	// content := []mail.Content{
	// 	*mail.NewContent("text/plain", "A new user has registered\nDetails:\nName: "+*user.First_name+" "+*user.Last_name+"\nEmail: "+*user.Email+"\nRoleType: "+*user.User_type+"\nPhoneNumber: "+user.Phone),
	// }

	m.AddPersonalizations(personalization)
	for _, c := range content {
		m.AddContent(&c)
	}

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
