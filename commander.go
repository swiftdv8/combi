package combi

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/abiosoft/ishell.v2"
)

// Commander wraps up commands and some configuration on how to run them
type Commander struct {
	commands map[string]*Command
	sync.RWMutex
	rootCmd             *cobra.Command
	preRequest          []CommandHook
	postRequest         []CommandHook
	requestHandler      RequestHandler
	responseHandler     ResponseHandler
	registrationHandler RegisterFunc
	errorHandler        ErrorHandler
	staticExec          StaticExec
	shellExec           ShellExec
}

// NewCommander returns an instantiated Commander object to contain managed commands
func NewCommander(rootCommand *cobra.Command) *Commander {
	return &Commander{
		commands:            map[string]*Command{},
		rootCmd:             rootCommand,
		errorHandler:        DefaultErrorHandler,
		responseHandler:     XMLPrettyPrintResponseHandler,
		staticExec:          GenericStaticHandler,
		shellExec:           GenericShellHandler,
		registrationHandler: DefaultRegistrationHandler,
	}
}

// HandleError will call the errorHandler for any commander errors
func (c *Commander) HandleError(err error) {
	c.errorHandler(err)
}

// SetErrorHandler sets the error handler for all commander errors
func (c *Commander) SetErrorHandler(eh ErrorHandler) {
	c.Lock()
	defer c.Unlock()
	c.errorHandler = eh
}

// RegistrationHandler will return the registered global register function
func (c *Commander) RegistrationHandler() RegisterFunc {
	c.RLock()
	defer c.RUnlock()
	return c.registrationHandler
}

// SetDefaultRegistrationHandler will set the global register function
func (c *Commander) SetDefaultRegistrationHandler(rh RegisterFunc) {
	c.Lock()
	defer c.Unlock()
	c.registrationHandler = rh
}

// ShellExec will return the registered global shell exec function
func (c *Commander) ShellExec() ShellExec {
	c.RLock()
	defer c.RUnlock()
	return c.shellExec
}

// SetDefaultShellExec will set the global shell exec function
func (c *Commander) SetDefaultShellExec(se ShellExec) {
	c.Lock()
	defer c.Unlock()
	c.shellExec = se
}

// StaticExec will return the registered global static exec function
func (c *Commander) StaticExec() StaticExec {
	c.RLock()
	defer c.RUnlock()
	return c.staticExec
}

// SetDefaultStaticExec will set the global static exec function
func (c *Commander) SetDefaultStaticExec(se StaticExec) {
	c.Lock()
	defer c.Unlock()
	c.staticExec = se
}

// DefaultRequestHandler returns the DefaultRequestHandler
func (c *Commander) DefaultRequestHandler() RequestHandler {
	c.RLock()
	defer c.RUnlock()
	return c.requestHandler
}

/*
SetDefaultRequestHandler will set a default request handler to be used on all
commands - request handlers specified on individual commands will be preferred
*/
func (c *Commander) SetDefaultRequestHandler(rh RequestHandler) {
	c.Lock()
	defer c.Unlock()
	c.requestHandler = rh
}

// DefaultResponseHandler returns the DefaultResponseHandler
func (c *Commander) DefaultResponseHandler() ResponseHandler {
	c.RLock()
	defer c.RUnlock()
	return c.responseHandler
}

/*
SetDefaultResponseHandler will set a default response handler to be used on all
commands - response handlers specified on individual commands will be preferred
*/
func (c *Commander) SetDefaultResponseHandler(rh ResponseHandler) {
	c.Lock()
	defer c.Unlock()
	c.responseHandler = rh
}

// AddPreRequestHooks provides for command decorators to be run before all requests are handled
func (c *Commander) AddPreRequestHooks(hooks ...CommandHook) {
	c.Lock()
	defer c.Unlock()
	for _, hook := range hooks {
		c.preRequest = append(c.preRequest, hook)
	}
}

// PreRequestHooks returns the refistered PreRequestHooks
func (c *Commander) PreRequestHooks() []CommandHook {
	c.RLock()
	defer c.RUnlock()
	return c.preRequest
}

// AddPostRequestHooks provides for command decorators to be run after all requests are handled
func (c *Commander) AddPostRequestHooks(hooks ...CommandHook) {
	c.Lock()
	defer c.Unlock()
	for _, hook := range hooks {
		c.postRequest = append(c.postRequest, hook)
	}
}

// PostRequestHooks returns the refistered PostRequestHooks
func (c *Commander) PostRequestHooks() []CommandHook {
	c.RLock()
	defer c.RUnlock()
	return c.postRequest
}

// FormatPrinter provides a simple interface for anything that implements Printf
type FormatPrinter interface {
	Printf(format string, a ...interface{})
}

// PrintCommandList will print a list of all registered commands with their short description
func (c *Commander) PrintCommandList(f FormatPrinter) {
	c.RLock()
	defer c.RUnlock()

	// first collect the keys for sorting and find the longest command name - so we can pad to line up descriptions
	longestCommandName := 0
	keys := []string{}
	for name := range c.commands {
		keys = append(keys, name)
		if len(name) > longestCommandName {
			longestCommandName = len(name)
		}
	}

	// alpha sort
	sort.Strings(keys)

	// print all commands, padding to line up descritions
	for _, k := range keys {
		command := c.commands[k]
		padLen := longestCommandName - len(k)

		f.Printf("%s - %s\n", (k + strings.Repeat(" ", padLen)), command.ShortDesc)
	}
}

// Cmd retrieves a command by name from the register
func (c *Commander) Cmd(name string) (*Command, error) {
	if cmd, ok := c.commands[name]; ok {
		return cmd, nil
	}

	return nil, errors.New("failed to resolve command in register")
}

// Add will add one or more commands to the command register
func (c *Commander) Add(cmds ...*Command) error {

	for _, cmd := range cmds {
		cmd.Commander = c
		c.Lock()
		c.commands[cmd.Name] = cmd
		c.Unlock()
		err := cmd.Register(c.rootCmd)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
RegisterShell will register commands with the provided shell, this should be
called after all commands have been added to the commander
*/
func (c *Commander) RegisterShell(shell *ishell.Shell) error {
	c.Lock()
	defer c.Unlock()
	for _, command := range c.commands {
		command.RegisterToShell(shell)
	}
	return nil
}

// All will return a map of all registered commands
func (c *Commander) All() map[string]*Command {
	c.Lock()
	defer c.Unlock()

	return c.commands
}
