package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var (
	usPhoneRegex *regexp.Regexp = regexp.MustCompile("^\\d{3}-\\d{3}-\\d{4}$")
	idRegex      *regexp.Regexp = regexp.MustCompile("^\\d{1,8}$")
)

const (
	maxNameCharacters = 15
)

type columnValueFunc func(*PersonRecord, string) error

// maps column header names from the input file to functions to set and validate the
// corresponding field in the output data json
//
// INTERNAL_ID : 8 digit positive integer. Cannot be empty.
// FIRST_NAME : 15 character max string. Cannot be empty.
// MIDDLE_NAME : 15 character max string. Can be empty.
// LAST_NAME : 15 character max string. Cannot be empty.
// PHONE_NUM : string that matches this pattern ###-###-####. Cannot be empty.
var name2function = map[string]columnValueFunc{
	"INTERNAL_ID": SetId,
	"FIRST_NAME":  SetFirstName,
	"MIDDLE_NAME": SetMiddleName,
	"LAST_NAME":   SetLastName,
	"PHONE_NUM":   SetPhoneNumber,
}

type PersonRecord struct {
	ID   int `json:"id"`
	Name struct {
		First  string `json:"first"`
		Middle string `json:"middle,omitempty"`
		Last   string `json:"last"`
	} `json:"name"`
	Phone string `json:"phone"`
}

// common validation for required name fields
func validateRequiredName(value string, fieldName string) error {
	if len(value) == 0 || len(value) > maxNameCharacters {
		msg := fmt.Sprintf("%s cannot be empty and must have a maximum of %d characters.", fieldName, maxNameCharacters)
		return errors.New(msg)
	}
	return nil
}

func SetId(person *PersonRecord, value string) error {
	if value == "" {
		return errors.New("ID cannot be empty")
	}
	if !idRegex.MatchString(value) {
		return errors.New("ID must be 1 to 8 digits")
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("ID encountered an invalid value: %s", value)
		return errors.New(msg)
	}
	person.ID = int(i)
	return nil
}
func SetFirstName(person *PersonRecord, value string) error {
	err := validateRequiredName(value, "FIRST_NAME")
	if err != nil {
		return err
	}
	person.Name.First = value
	return nil
}
func SetMiddleName(person *PersonRecord, value string) error {
	if len(value) > maxNameCharacters {
		msg := fmt.Sprintf("MIDDLE_NAME must have a maximum of %d characters.", maxNameCharacters)
		return errors.New(msg)
	}
	person.Name.Middle = value
	return nil
}
func SetLastName(person *PersonRecord, value string) error {
	err := validateRequiredName(value, "LAST_NAME")
	if err != nil {
		return err
	}
	person.Name.Last = value
	return nil
}
func SetPhoneNumber(person *PersonRecord, value string) error {
	formatErr := errors.New("PHONE_NUM must be in the format: ###-###-####")
	if len(value) != 12 {
		return formatErr
	}
	if !usPhoneRegex.MatchString(value) {
		return formatErr
	}
	person.Phone = value
	return nil
}
