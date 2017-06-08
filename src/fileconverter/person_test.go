package main

import (
	"testing"
)

func TestIdValidationNonNumber(t *testing.T) {
	p := PersonRecord{}
	if err := SetId(&p, "Must be a number"); err == nil {
		t.Errorf("%s: expected an error condition, but did not receive one.", t.Name())
		t.Fail()
	}
}

func TestIdValidationTooLong(t *testing.T) {
	p := PersonRecord{}
	if err := SetId(&p, "1234567890"); err == nil {
		t.Errorf("%s: expected an error condition, but did not receive one.", t.Name())
		t.Fail()
	}
}

func TestIdValidationOK(t *testing.T) {
	p := PersonRecord{}
	if err := SetId(&p, "12345678"); err != nil {
		t.Errorf("%s: 12345678 is a valid id, but received an error.", t.Name())
		t.Fail()
	}
}

func TestPhoneNumberValidationFail(t *testing.T) {
	p := PersonRecord{}
	if err := SetPhoneNumber(&p, "+1-304-555-5555"); err == nil {
		t.Errorf("%s: expected an error condition, but did not receive one.", t.Name())
		t.Fail()
	}
}

/// etc.
