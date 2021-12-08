package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/bgentry/speakeasy"
	"github.com/msteinert/pam"
)

func authService() error {
	t, err := pam.StartFunc("", "", func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return speakeasy.Ask(msg)
		case pam.PromptEchoOn:
			fmt.Print(msg + " ")
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			return input[:len(input)-1], nil
		case pam.ErrorMsg:
			fmt.Print(msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println(msg)
			return "", nil
		}
		return "", errors.New("Unrecognized message style")
	})
	if err != nil {
		return err
	}
	err = t.Authenticate(0)
	if err != nil {
		return err
	}
	//fmt.Println("Authentication succeeded!")
	return nil

}

func main() {
	if runtime.GOOS != "linux" { // also can be specified to FreeBSD
		fmt.Println("PAM(Pluggable Authentication Modules) only works with Linux!!")
		os.Exit(0)
	}

	err := authService()

	if err != nil {
		fmt.Println("Unable to authenticate your password or username! Abort!")
	} else {
		fmt.Println("Credential accepted! Proceeding to .....")
		// this is where you want to execute your main function for authenticated users.
	}
}
