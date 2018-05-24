package cli

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/RTradeLtd/Temporal/server"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

type CommandLine struct {
	Shell *ishell.Shell
}

// Initialize is used to init our command line app
func Initialize() {
	var cli CommandLine
	//  generate a new shell
	shell := ishell.New()
	cli.Shell = shell
	// setup the shell
	cli.SetupShell()
	// run the shell
	cli.Shell.Run()
}

func (cl *CommandLine) SetupShell() {
	// print out some greeting information when the shell is ran
	cl.Shell.Println("Temporal Command Line Interactive Shell")
	cl.Shell.Println("Version 0.0.5alpha")

	// Setup the commands

	cl.Shell.AddCmd(&ishell.Cmd{
		Name: "register-payment",
		Help: "registers a payment",
		Func: func(c *ishell.Context) {
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Println("enter uploader address")
			scanner.Scan()
			uploaderAddressString := scanner.Text()
			fmt.Println(uploaderAddressString)
			fmt.Println("enter the content hash the user is paying for")
			scanner.Scan()
			contentHashString := scanner.Text()
			fmt.Println("enter retention period in months")
			scanner.Scan()
			retentionPeriodInMonthsString := scanner.Text()
			fmt.Println("enter amount to be charged in units of wei")
			scanner.Scan()
			chargeAmountInWeiString := scanner.Text()
			fmt.Println("enter payment method. 0 = rtc , 1 = eth")
			scanner.Scan()
			paymentMethod := scanner.Text()
			switch paymentMethod {
			case "0":
				break
			case "1":
				break
			default:
				fmt.Println("not a valid option, exiting command process")
				return
			}
			retentionPeriodInt, err := strconv.ParseInt(retentionPeriodInMonthsString, 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			chargeAmountInt, err := strconv.ParseInt(chargeAmountInWeiString, 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			methodUint, err := strconv.ParseUint(paymentMethod, 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			retentionPeriodBig := big.NewInt(retentionPeriodInt)
			chargeAmountBig := big.NewInt(chargeAmountInt)
			manager := server.Initialize()
			tx, err := manager.RegisterPaymentForUploader(uploaderAddressString, contentHashString, retentionPeriodBig, chargeAmountBig, uint8(methodUint))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(tx)
		},
	})
}
