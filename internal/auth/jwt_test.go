package auth

import (
	"github.com/google/uuid"
	"log"
	"testing"
	"time"
)

func TestVaidJWT(t *testing.T) {
	type TestCase struct {
		testName string
		inputID  uuid.UUID
		secret   string
	}

	cases := []TestCase{
		{
			testName: "Base Case",
			inputID:  uuid.New(),
			secret:   "chirpy",
		},
		{
			testName: "Empty Secret",
			inputID:  uuid.New(),
			secret:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			tokenString, err := MakeJWT(c.inputID, c.secret, time.Second*10)
			if err != nil {
				log.Fatalf("Unable to create a JWT: %s", err)
			}

			outputUserID, err := ValidateJWT(tokenString, c.secret)
			if err != nil {
				log.Fatalf("Unable to validate the token string: %s", err)
			}

			if outputUserID != c.inputID {
				log.Fatal("The input user ID and output user ID don't match")
			}
		})
	}
}

func TestInvalidJWT(t *testing.T) {
	trueID := uuid.New()
	trueSecret := "chirpy"

	type TestCase struct {
		testName string
		actualID uuid.UUID
		falseID  uuid.UUID
		actualSecret   string
		falseSecret   string
	}

	cases := []TestCase{
		{
			testName: "Mismatching userID",
			actualID: trueID,
			falseID:  uuid.New(),
			actualSecret:  trueSecret,
			falseSecret: trueSecret,
		},
		{
			testName: "Empty Secret",
			actualID: trueID,
			falseID:  uuid.New(),
			actualSecret:   "",
			falseSecret:   "",
		},
		{
			testName: "Mismatching Secret",
			actualID: trueID,
			falseID:  trueID,
			actualSecret:  trueSecret, 
			falseSecret:   "yprihc",
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			tokenString, err := MakeJWT(c.actualID, c.actualSecret, time.Second*10)
			if err != nil {
				log.Fatalf("Unable to create a JWT: %s", err)
			}

			outputUserID, err := ValidateJWT(tokenString, c.falseSecret)
			if err != nil && c.falseSecret == trueSecret {
				log.Fatalf("Unable to validate the token string: %s", err)
			}

			if outputUserID == c.falseID {
				log.Fatal("False user ID is matching the actual user ID")
			}
		})
	}
}

// Will be implemented in the future
// The ValidateJWT function need to be split into helpers
// To get just the toaken, err
//func TestExpiredJWT (t *testing.T) {
//	type TestCase struct {
//		testName string
//		inputID    uuid.UUID
//		secret   string
//		waitDuration time.Duration
//	}
//
//	cases := []TestCase{
//		TestCase{
//			testName: "Base Case",
//			inputID:    uuid.New(),
//			secret:   "chirpy",
//			waitDuration: time.Second * 1,
//		},
//		TestCase{
//			testName: "Empty Secret Case",
//			inputID:    uuid.New(),
//			secret:   "",
//			waitDuration: time.Second * 1,
//		},
//	}
//
//	for _, c := range cases {
//		t.Run(c.testName, func(t *testing.T) {
//			tokenString, err := MakeJWT(c.inputID, c.secret, time.Millisecond * 500)
//			if err != nil {
//				log.Fatalf("Unable to create a JWT: %s", err)
//			}
//
//			time.Sleep(c.waitDuration)
//			outputUserID, err := ValidateJWT(tokenString, c.secret)
//			if err != nil {
//				log.Fatalf("Unable to validate the token string: %s", err)
//			}
//
//			expTime := outputuser
//			if outputUserID == c.inputID {
//				log.Fatal("Token should be expired")
//			}
//		})
//	}
//}
