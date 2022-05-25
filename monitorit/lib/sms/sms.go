package sms

import (
	"fmt"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// Client provides functionality for sending SMS messages using Twilio.
type Client struct {
	twilioClient      *twilio.RestClient
	twilioPhoneNumber string
}

func NewClient(twilioAccountSID, twilioAuthToken string, twilioPhoneNumber string) *Client {
	return &Client{
		twilioClient: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: twilioAccountSID,
			Password: twilioAuthToken,
		}),
		twilioPhoneNumber: twilioPhoneNumber,
	}
}

func (c *Client) SendMessage(phoneNumber, message string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(c.twilioPhoneNumber)
	params.SetBody(message)
	resp, err := c.twilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Check for an error response from twilio
	if resp.ErrorCode != nil {
		return fmt.Errorf("twilio send message unsuccessful: %d: %s", *resp.ErrorCode, *resp.ErrorMessage)
	}
	return nil
}
