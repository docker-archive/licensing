package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var states = map[string]bool{
	"CA": true,
	"NV": true,
}

type User struct {
	FirstName string
	LastName  string
	Email     string
	Website   string
	Addresses []Address
}

type Address struct {
	AddressLine string
	City        string
	State       string
	Zip         string
	Country     string
}

func (u *User) Validate() Errors {
	var errs Errors

	if IsEmpty(u.FirstName) {
		errs = append(errs, InvalidEmpty("first_name"))
	}

	if IsEmpty(u.LastName) {
		errs = append(errs, InvalidEmpty("last_name"))
	}

	if !IsEmail(u.Email) {
		errs = append(errs, InvalidEmail("email", u.Email))
	}

	if !IsURL(u.Website) {
		errs = append(errs, InvalidURL("website", u.Website))
	}

	for i, address := range u.Addresses {
		addrErrs := validateAddress(i, address)
		errs = append(errs, addrErrs...)
	}

	return errs
}

func validateAddress(index int, a Address) Errors {
	var errs Errors

	if IsEmpty(a.AddressLine) {
		errs = append(errs,
			InvalidEmpty(fmt.Sprintf("address[%v] address_line_1", index)))
	}

	if IsEmpty(a.City) {
		errs = append(errs,
			InvalidEmpty(fmt.Sprintf("address[%v] city", index)))
	}

	if IsEmpty(a.Country) {
		errs = append(errs,
			InvalidEmpty(fmt.Sprintf("address[%v] country", index)))
	}

	if IsEmpty(a.State) {
		errs = append(errs,
			InvalidEmpty(fmt.Sprintf("address[%v] state", index)))
	}

	// custom error example
	if _, ok := states[a.State]; !ok {
		errs = append(errs, &Error{
			FieldName:  fmt.Sprintf("address[%v] state", index),
			FieldValue: a.State,
		})
	}

	if !Matches(a.Zip, "^\\d{5}$") {
		errs = append(errs, &Error{
			FieldName:  fmt.Sprintf("address[%v] zip", index),
			FieldValue: a.Zip,
		})
	}

	return errs
}

func testUser() User {
	return User{
		FirstName: "testFirstName",
		LastName:  "testLastName",
		Email:     "bar@foo.com",
		Website:   "https://youtu.be/vqCOss4hqnE",
		Addresses: []Address{
			{
				AddressLine: "1234 Foo st",
				City:        "SF",
				State:       "CA",
				Country:     "USA",
				Zip:         "94107",
			},
		},
	}
}

func TestValidateIsValid(t *testing.T) {
	t.Parallel()

	usr := testUser()

	errs := usr.Validate()
	require.Equal(t, 0, len(errs))
}

func TestValidateIsInvalid(t *testing.T) {
	t.Parallel()

	// invalid first name
	usr := testUser()
	usr.FirstName = ""

	errs := usr.Validate()
	require.Equal(t, 1, len(errs))
	require.Equal(t, "first_name invalid: not provided", errs[0].Error())

	// invalid email
	usr = testUser()
	usr.Email = "foobar"

	errs = usr.Validate()
	require.Equal(t, 1, len(errs))
	require.Equal(t, "email invalid: foobar is an invalid email address", errs[0].Error())

	// invalid website
	usr = testUser()
	usr.Website = "foobar"

	errs = usr.Validate()
	require.Equal(t, 1, len(errs))
	require.Equal(t, "website invalid: foobar is an invalid url", errs[0].Error())

	// invalid address
	usr = testUser()
	usr.Addresses[0].AddressLine = ""

	errs = usr.Validate()
	require.Equal(t, 1, len(errs))
	require.Equal(t, "address[0] address_line_1 invalid: not provided", errs[0].Error())

	// multiple invalid fields
	usr = testUser()
	usr.LastName = ""
	usr.Addresses[0].City = ""
	usr.Addresses[0].Zip = "25"

	errs = usr.Validate()
	require.Equal(t, 3, len(errs))
	require.Equal(t, "last_name invalid: not provided", errs[0].Error())
	require.Equal(t, "address[0] city invalid: not provided", errs[1].Error())
	require.Equal(t, "address[0] zip invalid: '25' is not a valid value", errs[2].Error())
}
