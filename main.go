package main

import (
	"errors"
	"fmt"
	"github.com/yourgfslove/BLOGagregator/internal/config"
	"os"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commandMap[cmd.name]
	if !ok {
		return errors.New("command not found")
	}
	return handler(s, cmd)
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	var s state
	s.cfg = &cfg
	cmds := commands{make(map[string]func(*state, command) error)}
	cmds.register("login", loginHandler)
	if len(os.Args) < 2 {
		fmt.Println("to many commands")
		os.Exit(1)
	}
	cmd := command{os.Args[1], os.Args[2:]}
	if err = cmds.run(&s, cmd); err != nil {
		fmt.Println(err)
	}
}

func loginHandler(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("wrong number of arguments")
	}
	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Println(s.cfg.CurrentUserName + " logged in")
	return nil
}
