package validator

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^\w+$`).MatchString
	isValidName     = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateStringLen(val string, minLen int, maxLen int) error {
	n := len(val)
	if n < minLen || n > maxLen {
		return fmt.Errorf("must contain between %d-%d characters", minLen, maxLen)
	}
	return nil
}

func ValidateUsername(uname string) error {
	if err := ValidateStringLen(uname, 3, 100); err != nil {
		return err
	}
	if !isValidUsername(uname) {
		return fmt.Errorf("must contain only letters, digits or underscore")
	}
	return nil
}

func ValidatePassword(pwd string) error {
	return ValidateStringLen(pwd, 6, 64)
}

func ValidateEmail(email string) error {
	if err := ValidateStringLen(email, 6, 64); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func ValidateName(name string) error {
	if err := ValidateStringLen(name, 3, 100); err != nil {
		return err
	}
	if !isValidName(name) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}
