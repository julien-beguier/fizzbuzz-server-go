package main

import (
	"fmt"
	"testing"
)

type TestCase struct {
	qp            QueryParam
	expextedError string
}

func TestErrorCasesCheckParams(t *testing.T) {
	errorCases := []TestCase{
		// Empty set -> equivalent of not giving any parameter
		{QueryParam{"", "", "", "", ""}, "parameter limit is required\nparameter int1 is required\nparameter int2 is required\nparameter str1 is required\nparameter str2 is required"},

		// Numbers: not a numeric value, negative value & out of range integer value
		{QueryParam{Limit: "azerty", Int1: "3", Int2: "5", Str1: "abc", Str2: "def"}, "parameter limit is not a numeric value (received:azerty)"},
		{QueryParam{Limit: "100", Int1: "-456", Int2: "5", Str1: "abc", Str2: "def"}, "int type parameter cannot be less than 1 (received:-456)"},
		{QueryParam{Limit: "100", Int1: "3", Int2: "77777777777777777777777777777777777777777777777777777", Str1: "abc", Str2: "def"}, "int type parameter is out of range (received:77777777777777777777777777777777777777777777777777777)"},

		// String: not alphanumeric characters, more than 64 characters
		{QueryParam{Limit: "100", Int1: "3", Int2: "5", Str1: "!!!!", Str2: "str2"}, "parameter str1 is not an alphanumeric value (received:!!!!)"},
		{QueryParam{Limit: "100", Int1: "3", Int2: "5", Str1: "str1", Str2: "65CHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARACTERS"}, "parameter str2 cannot be over 64 characters (received:65CHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARACTERS)"},
	}

	for _, tc := range errorCases {
		_, err := checkParams(tc.qp)
		if err != nil && err.Error() != tc.expextedError {
			t.Log(fmt.Sprintf("error should be '%s' but got '%s'", tc.expextedError, err.Error()))
			t.Fail()
		}
	}
}

func TestValidCaseCheckParams(t *testing.T) {
	limitExpected := uint(50)
	int1Expected := uint(3)
	int2Expected := uint(5)
	str1Expected := "toto"
	str2Expected := "titi"
	validCase := QueryParam{
		fmt.Sprint(limitExpected), // string
		fmt.Sprint(int1Expected),  // string
		fmt.Sprint(int2Expected),  // string
		str1Expected,
		str2Expected,
	}

	statistic, err := checkParams(validCase)
	if err != nil {
		t.Log(fmt.Sprintf("error should be <nil> but got '%s'", err.Error()))
		t.Fail()
	} else {
		if statistic.Limit != limitExpected {
			t.Log(fmt.Sprintf("value should be '%d' but got '%d'", limitExpected, statistic.Limit))
			t.Fail()
		}
		if statistic.Int1 != int1Expected {
			t.Log(fmt.Sprintf("value should be '%d' but got '%d'", int1Expected, statistic.Int1))
			t.Fail()
		}
		if statistic.Int2 != int2Expected {
			t.Log(fmt.Sprintf("value should be '%d' but got '%d'", int2Expected, statistic.Int2))
			t.Fail()
		}
		if statistic.Str1 != str1Expected {
			t.Log(fmt.Sprintf("error should be '%s' but got '%s'", str1Expected, statistic.Str1))
			t.Fail()
		}
		if statistic.Str2 != str2Expected {
			t.Log(fmt.Sprintf("error should be '%s' but got '%s'", str2Expected, statistic.Str2))
			t.Fail()
		}
	}
}
