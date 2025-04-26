package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateLogin(t *testing.T) {
	var tests = []struct {
		name string
		in   string
		out  bool
	}{
		{
			"valid login",
			"amogus228",
			true,
		},
		{
			"empty login",
			"",
			false,
		},
		{
			"too long login",
			"thisloginiswaytoolongtobevalid",
			false,
		},
		{
			"login with space",
			"invalid user",
			false,
		},
		{
			"login starts with dot",
			".invalidUser",
			false,
		},
		{
			"login starts with underscore",
			"_invalidUser",
			false,
		},
		{
			"login with invalid character",
			"user@name",
			false,
		},
		{
			"login with special characters",
			"!#aboba228",
			false,
		},
	}

	for _, test := range tests {
		actual := validateLogin(test.in)
		require.Equal(t, test.out, actual, test.name)
	}
}

func TestValidatePassword(t *testing.T) {
	var tests = []struct {
		name string
		in   string
		out  bool
	}{
		{
			"valid password",
			"228Matvey!",
			true,
		},
		{
			"too short password",
			"short",
			false,
		},
		{
			"too long password",
			"ThisPasswordIsWayTooLong123!adadadad",
			false,
		},
		{
			"password without uppercase",
			"lowercase1!",
			false,
		},
		{
			"password without lowercase",
			"UPPERCASE1!",
			false,
		},
		{
			"password without digit",
			"NoDigitsHere!",
			false,
		},
		{
			"password without special character",
			"NoSpecial123",
			false,
		},
		{
			"password with space",
			"Invalid Pass1!",
			false,
		},
	}

	for _, test := range tests {
		actual := validatePassword(test.in)
		require.Equal(t, test.out, actual, test.name)
	}
}

func TestValidateCreds(t *testing.T) {
	var tests = []struct {
		name string
		in   string
		out  bool
	}{
		{
			"valid name",
			"John-Doe",
			true,
		},
		{
			"empty name",
			"",
			false,
		},
		{
			"too short name",
			"J",
			false,
		},
		{
			"too long name",
			"ThisNameIsWayTooLongForValidationMayBe",
			false,
		},
		{
			"name starts with -",
			"-John",
			false,
		},
		{
			"name ends with hyphen",
			"John-",
			false,
		},
		{
			"name with special characters",
			"John#Doe!Smith",
			false,
		},
		{
			"name with more than one hyphen",
			"John-Doe-Smith",
			false,
		},
		{
			"name with numbers",
			"John123",
			false,
		},
		{
			"chinese name test",
			"霜",
			true,
		},
		{
			"arabian",
			"عباس",
			true,
		},
	}

	for _, test := range tests {
		actual := validateCreds(test.in)
		require.Equal(t, test.out, actual, test.name)
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name      string
		inputPass string
		inputHash string
		inputSalt string
		expected  bool
	}{
		{"Correct password", "aboBus228!", "4df69e78e537818105bd5ed0d04dd3666af16b6a24c2eeaecd1558fdb9c0c468", "salt", true},
		{"Wrong password", "wrongPassword", "4df69e78e537818105bd5ed0d04dd3666af16b6a24c2eeaecd1558fdb9c0c468", "salt", false},
		{"Empty password", "", "4df69e78e537818105bd5ed0d04dd3666af16b6a24c2eeaecd1558fdb9c0c468", "salt", false},
	}

	for _, test := range tests {
		actual := CheckPassword(test.inputPass, test.inputHash, test.inputSalt)
		require.Equal(t, test.expected, actual, test.name)
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantErr  error
	}{
		{
			name:     "valid user",
			login:    "validUser123",
			password: "ValidPass123!",
			wantErr:  nil,
		},
		{
			name:     "invalid login",
			login:    " invalidUser",
			password: "ValidPass123!",
			wantErr:  errors.New("invalid login"),
		},
		{
			name:     "invalid password",
			login:    "validUser123",
			password: "short",
			wantErr:  errors.New("invalid password"),
		},
		{
			name:     "both invalid",
			login:    "",
			password: "123",
			wantErr:  errors.New("invalid login"), // первая ошибка, которую вернёт switch
		},
	}

	for _, tt := range tests {
		err := ValidateUser(tt.login, tt.password)
		if tt.wantErr != nil {
			require.EqualError(t, err, tt.wantErr.Error(), tt.name)
		} else {
			require.NoError(t, err, tt.name)
		}
	}
}

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   error
	}{
		{
			name:      "valid profile",
			firstName: "John",
			lastName:  "Doe",
			wantErr:   nil,
		},
		{
			name:      "invalid first name",
			firstName: "",
			lastName:  "Doe",
			wantErr:   errors.New("invalid first name"),
		},
		{
			name:      "invalid last name",
			firstName: "John",
			lastName:  "123Doe",
			wantErr:   errors.New("invalid last name"),
		},
		{
			name:      "both invalid",
			firstName: "!",
			lastName:  "",
			wantErr:   errors.New("invalid first name"),
		},
	}

	for _, tt := range tests {
		err := ValidateProfile(tt.firstName, tt.lastName)
		if tt.wantErr != nil {
			require.EqualError(t, err, tt.wantErr.Error(), tt.name)
		} else {
			require.NoError(t, err, tt.name)
		}
	}
}
