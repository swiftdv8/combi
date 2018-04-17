package combi

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
	"gopkg.in/abiosoft/ishell.v2"
)

// GenericStaticHandler provides a generic handler for static cli commands
var GenericStaticHandler = func(command *Command, cmd *cobra.Command, args []string) error {

	// @TODO remove or make verbose
	// fmt.Println(cmd.Name())

	// run validation
	result, err := govalidator.ValidateStruct(command.Request)
	if err != nil {
		println(result)
		return fmt.Errorf("validation error: %s" + err.Error())
	}

	err = command.HandleRequest(command.Request, command.Response)
	if err != nil {
		return fmt.Errorf("error from request handler: %s", err)
	}

	return command.HandleResponse(command.Response)
}

var DefaultRegistrationHandler = func(parentCmd *cobra.Command, cmd *Command) error {

	// generate cobra command to handle static calls
	staticCmd := cmd.Static()
	// fmt.Println("pre-inspect", cmd.Name)
	fis, err := InspectStruct(cmd.Request)
	if err != nil {
		return err
	}

	// fmt.Printf("%+q", fis)

	// loop over fields and generate flags
	for _, fi := range fis {
		if fi.LFlag != "" && fi.Hint != "" {
			switch ptr := fi.FieldPtr.(type) {
			case *string:
				if len(fi.SFlag) == 1 {
					staticCmd.Flags().StringVarP(ptr, fi.LFlag, fi.SFlag, "", fi.Hint)
				} else {
					staticCmd.Flags().StringVar(ptr, fi.LFlag, "", fi.Hint)
				}
			case *int:
				if len(fi.SFlag) == 1 {
					staticCmd.Flags().IntVarP(ptr, fi.LFlag, fi.SFlag, 0, fi.Hint)
				} else {
					staticCmd.Flags().IntVar(ptr, fi.LFlag, 0, fi.Hint)
				}
			default:
				log.Println("returning error")
				return errors.New("unhandled type")
			}

		}
	}

	// assign command
	parentCmd.AddCommand(staticCmd)

	return nil
}

// DefaultErrorHandler provides a simple fatal error implementation
var DefaultErrorHandler = func(err error) {
	log.Fatal(err)
}

/*
GenericShellHandler provides a generic handler for shell commands, uses reflection
to identify required fields and prompt user for values via shell, may need bypassing
for more advanced structs
*/
var GenericShellHandler = func(command *Command, c *ishell.Context) error {

	fis, err := InspectStruct(command.Request)
	if err != nil {
		return err
	}

	// fmt.Printf("\n%+v\n\n", fis)
	// fmt.Printf("%q\n", fis)

	required, optional := splitRequiredFields(fis)

	// collect values for all required fields
	for _, fi := range required {
		err = collectShellValue(c, fi)
		if err != nil {
			return err
		}
	}

	// once we have collected all required vals, present optional fields
	selected := 1
	for selected >= 0 {
		selected = presentOptions(c, optional)
		if selected >= 0 {
			err = collectShellValue(c, optional[selected])
			if err != nil {
				return err
			}
		}
	}

	err = command.HandleRequest(command.Request, command.Response)
	if err != nil {
		return fmt.Errorf("error from request handler: %s", err)
	}

	return command.HandleResponse(command.Response)
}

func XMLCompactPrintResponseHandler(resp interface{}) error {

	res, err := xml.Marshal(resp)
	if err != nil {
		return fmt.Errorf("unable to marshal response: %s", err)
	}

	fmt.Println(string(res))

	return nil
}

func XMLPrettyPrintResponseHandler(resp interface{}) error {

	res, err := xml.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal response: %s", err)
	}

	fmt.Println(string(res))
	// fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	return nil
}
