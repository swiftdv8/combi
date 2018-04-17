package combi

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"gopkg.in/abiosoft/ishell.v2"
)

/*
StaticExec provides a function signature to provide static cli execution (not within shell)
matches cobra command Run signature
*/
type StaticExec func(command *Command, cmd *cobra.Command, args []string) error

// ShellExec provides a function signature to provide shell based command execution
type ShellExec func(command *Command, c *ishell.Context) error

// RegisterFunc should register the command and bind any required variables to the request object
type RegisterFunc func(parentCmd *cobra.Command, cmd *Command) error

// RequestHandler is the function called to make the request and populate the response
type RequestHandler func(request interface{}, response interface{}) error

// ResponseHandler is the function called to handle presentation of the response
type ResponseHandler func(response interface{}) error

// CommandHook provides a hook signature to provide hooks to be run before or after all request handlers
type CommandHook func(c *Command) error

// ErrorHandler provides a handler fo all commander errors
type ErrorHandler func(err error)

// Command represents an OMP operation or other cli call
type Command struct {
	Commander       *Commander
	Name            string
	ShortDesc       string
	LongDesc        string
	StaticExec      StaticExec
	ShellExec       ShellExec
	RequestHandler  RequestHandler
	ResponseHandler ResponseHandler
	Request         interface{}
	Response        interface{}
	RegisterFunc    RegisterFunc
}

// Register is called by the Command Register to handle the specifics of command registration
func (c *Command) Register(rootCmd *cobra.Command) error {

	if c.RegisterFunc == nil {
		globalRegistrationHandler := c.Commander.RegistrationHandler()
		if globalRegistrationHandler == nil {
			return errors.New("no registration handler defined")
		}

		err := globalRegistrationHandler(rootCmd, c)
		if err != nil {
			return fmt.Errorf("error from commander registration handler: %s", err)
		}

	} else {
		err := c.RegisterFunc(rootCmd, c)
		if err != nil {
			return fmt.Errorf("error from command registration handler: %s", err)
		}
	}
	return nil
}

// Static will generate a static cobra command from the Command
func (c *Command) Static() *cobra.Command {
	return &cobra.Command{
		Use:   c.Name,
		Short: c.ShortDesc,
		Long:  c.LongDesc,
		Run:   c.handleStatic,
	}
}

// RegisterToShell will register the command to the supplied shell
func (c *Command) RegisterToShell(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: c.Name,
		Help: c.ShortDesc,
		Func: c.handleShell,
	})
}

// HandleRequest calls the appropriate request handler for the command, local preferred
func (c *Command) HandleRequest(req, resp interface{}) error {

	if c.RequestHandler == nil {
		globalRequestHandler := c.Commander.DefaultRequestHandler()
		if globalRequestHandler == nil {
			return errors.New("no request handler defined")
		}

		err := globalRequestHandler(req, resp)
		if err != nil {
			return fmt.Errorf("error from commander request handler: %s", err)
		}

	} else {
		err := c.RequestHandler(req, resp)
		if err != nil {
			return fmt.Errorf("error from command request handler: %s", err)
		}
	}

	return nil
}

// HandleResponse calls the appropriate response handler for the command, local preferred
func (c *Command) HandleResponse(resp interface{}) error {

	if c.ResponseHandler == nil {
		globalResponseHandler := c.Commander.DefaultResponseHandler()
		if globalResponseHandler == nil {
			return errors.New("no response handler defined")
		}
		err := globalResponseHandler(resp)
		if err != nil {
			return fmt.Errorf("error from commander response handler: %s", err)
		}

	} else {
		err := c.ResponseHandler(resp)
		if err != nil {
			return fmt.Errorf("error from command response handler: %s", err)
		}
	}

	return nil
}

func (c *Command) handleStatic(cmd *cobra.Command, args []string) {

	// run preHooks
	preHooks := c.Commander.PreRequestHooks()
	for _, preHook := range preHooks {
		err := preHook(c)
		if err != nil {
			c.Commander.HandleError(err)
		}
	}

	// run static exec function, fallback to global if not defined on command
	if c.StaticExec == nil {
		globalStaticExec := c.Commander.StaticExec()
		if globalStaticExec == nil {
			c.Commander.HandleError(errors.New("no static exec handler defined"))
		} else {
			err := globalStaticExec(c, cmd, args)
			if err != nil {
				c.Commander.HandleError(fmt.Errorf("error from commander static exec: %s", err))
			}

		}
	} else {
		err := c.StaticExec(c, cmd, args)
		if err != nil {
			c.Commander.HandleError(fmt.Errorf("error from command static exec: %s", err))
		}
	}

	// run postHooks
	postHooks := c.Commander.PostRequestHooks()
	for _, postHook := range postHooks {
		err := postHook(c)
		if err != nil {
			c.Commander.HandleError(err)
		}
	}
}

func (c *Command) resetStruct(structPtr interface{}) (interface{}, error) {

	// validate is pointer
	ptrVal := reflect.ValueOf(structPtr)
	if ptrVal.Kind() != reflect.Ptr {
		return nil, ErrStructPtrExpected
	}

	// reflect the underlying value and validate is struct
	val := reflect.Indirect(ptrVal)
	if val.Kind() != reflect.Struct {
		return nil, ErrStructPtrExpected
	}

	return reflect.New(val.Type()).Interface(), nil
}

func (c *Command) handleShell(sc *ishell.Context) {
	var err error
	c.Request, err = c.resetStruct(c.Request)
	if err != nil {
		c.Commander.HandleError(err)
	}
	c.Response, err = c.resetStruct(c.Response)
	if err != nil {
		c.Commander.HandleError(err)
	}

	// run preHooks
	preHooks := c.Commander.PreRequestHooks()
	for _, preHook := range preHooks {
		err := preHook(c)
		if err != nil {
			c.Commander.HandleError(err)
		}
	}

	// run static exec function, fallback to global if not defined on command
	if c.ShellExec == nil {
		globalShellExec := c.Commander.ShellExec()
		if globalShellExec == nil {
			c.Commander.HandleError(errors.New("no shell exec handler defined"))
		} else {
			err := globalShellExec(c, sc)
			if err != nil {
				c.Commander.HandleError(fmt.Errorf("error from commander shell exec: %s", err))
			}

		}
	} else {
		err := c.ShellExec(c, sc)
		if err != nil {
			c.Commander.HandleError(fmt.Errorf("error from command shell exec: %s", err))
		}
	}

	// run postHooks
	postHooks := c.Commander.PostRequestHooks()
	for _, postHook := range postHooks {
		err := postHook(c)
		if err != nil {
			c.Commander.HandleError(err)
		}
	}
}
