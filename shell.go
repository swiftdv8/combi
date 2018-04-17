package combi

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/abiosoft/ishell.v2"
)

var (
	minErrWidth        = 37
	errTitle           = "error"
	errorInvalidOption = errors.New("Invalid option, try again")
)

// present optional query params to the user
func presentOptions(c *ishell.Context, fis []*FieldInfo) int {
	// fmt.Printf("%+q", fis)
present:
	if len(fis) > 0 {
		c.Println()
		c.Println("Optional query paramaters:")
		c.Println()
		c.Println("[0] - I'm done")
		// index := 1
		for i, fi := range fis {
			c.Printf("[%d] - %s%s\n", i+1, fi.Namespace, fi.Name)
		}

		c.Println()
		c.Println("Select an option: ")
		selected := c.ReadLine()
		intVal, err := strconv.Atoi(selected)
		if err != nil {
			shellPrintError(c, errorInvalidOption)
			goto present
		}
		intVal--

		if intVal == -1 {
			return -1
		}

		if intVal < -1 || intVal > (len(fis)-1) {
			shellPrintError(c, errorInvalidOption)
			goto present
		}

		return intVal
	}
	return -1
}

func presentOptionalFieldsMultiSelect(c *ishell.Context, fis []*FieldInfo) int {
	if len(fis) > 0 {

		opts := []string{"I'm done"}
		for _, fi := range fis {
			opts = append(opts, fi.Namespace+fi.Name)
		}
		intVal := c.MultiChoice(opts, "Optional query paramaters:")
		// c.Println()
		// c.Println("Optional query paramaters:")
		// c.Println()
		// c.Println("[0] - I'm done")
		// // index := 1
		// for i, fi := range fis {
		// 	c.Printf("[%d] - %s%s\n", i+1, fi.Namespace, fi.Name)
		// }

		// c.Println()
		// c.Println("Select an option: ")
		// selected := c.ReadLine()
		// intVal, err := strconv.Atoi(selected)
		// if err != nil {
		// 	shellPrintError(c, errorInvalidOption)
		// 	goto present
		// }
		intVal--

		if intVal == -1 {
			return -1
		}

		if intVal < -1 || intVal > (len(fis)-1) {
			return -1
			// shellPrintError(c, errorInvalidOption)
			// goto present
		}

		return intVal
	}
	return -1
}

func shellPrintError(c *ishell.Context, err error) {
	errStringWidth := len(err.Error())
	if errStringWidth < minErrWidth {
		errStringWidth = minErrWidth
	}

	padderWidth := (errStringWidth - 7) / 2

	errorBorderTop := fmt.Sprintf("%s %s %s", strings.Repeat("*", padderWidth), errTitle, strings.Repeat("*", padderWidth))
	errorBorderBot := fmt.Sprintf("%s", strings.Repeat("*", errStringWidth))

	c.Println(errorBorderTop)
	c.Printf("%s\n", err)
	c.Println(errorBorderBot)
}

// collectShellValue collects a value from the user within their shell
// @TODO extend type support
func collectShellValue(c *ishell.Context, fi *FieldInfo) error {
	if !fi.Field.CanSet() {
		log.Println(fi.Field.Kind())
		return fmt.Errorf("unable to set value (CanSet=false)")
	}

	c.Print(fi.Name + ":")
	strVal := c.ReadLine()

	switch fi.Field.Kind() {
	case reflect.String:
		fi.Field.SetString(strVal)
		return nil
	case reflect.Int:
		intVal, err := strconv.Atoi(strVal)
		if err != nil || fi.Field.OverflowInt(int64(intVal)) {
			return fmt.Errorf("failed to convert user input to integer")
		}

		fi.Field.SetInt(int64(intVal))
		return nil
	default:
		return fmt.Errorf("unsupported value type: %s", fi.Field.Kind())
	}
}
