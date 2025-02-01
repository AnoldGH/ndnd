package ndncert

import "fmt"

type ChallengeEmail struct {
	// Email address to send the challenge to.
	Email string
	// Callback to get the code from the user.
	CodeCallback func(status string) string
}

func (*ChallengeEmail) Name() string {
	return KwEmail
}

func (c *ChallengeEmail) Request(input ParamMap, status *string) (ParamMap, error) {
	// Validate challenge configuration
	if len(c.Email) == 0 || c.CodeCallback == nil {
		return nil, fmt.Errorf("email challenge not configured")
	}

	// Initial request parameters
	if input == nil {
		return ParamMap{
			KwEmail: []byte(c.Email),
		}, nil
	}

	// Challenge response code
	if status != nil && (*status == "need-code" || *status == "wrong-code") {
		code := c.CodeCallback(*status)
		if code == "" {
			return nil, fmt.Errorf("no code provided")
		}

		return ParamMap{
			KwCode: []byte(code),
		}, nil
	}

	// Unknown status
	return nil, fmt.Errorf("unknown input to email challenge")
}
