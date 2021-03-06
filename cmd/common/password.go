package common

import (
	"flag"
	"fmt"
	"os"

	"github.com/howeyc/gopass"
)

// GetPassword gets password from user input
func GetPassword() ([]byte, error) {
	fmt.Printf("Password:")
	passwd, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	return passwd, nil
}

// GetConfirmedPassword gets double confirmed password from user input
func GetConfirmedPassword() ([]byte, error) {
	fmt.Printf("Password:")
	first, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Re-enter Password:")
	second, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	if len(first) != len(second) {
		fmt.Println("Unmatched Password")
		os.Exit(1)
	}
	for i, v := range first {
		if v != second[i] {
			fmt.Println("Unmatched Password")
			os.Exit(1)
		}
	}
	return first, nil
}

// GetFlagPassword gets node's wallet password from command line or user input
func GetFlagPassword() ([]byte, error) {
	var password []byte
	var err error
	if len(os.Args) == 1 {
		password, err = GetPassword()
		if err != nil {
			return nil, err
		}
	} else {
		var p string
		flag.StringVar(&p, "p", "", "wallet password")
		flag.Parse()
		if p == "" {
			fmt.Println("Invalid parameter, use '-p <password>' to specify a not nil wallet password.")
			os.Exit(1)
		}
		password = []byte(p)
	}

	return password, nil
}
