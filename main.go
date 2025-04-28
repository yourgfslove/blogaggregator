package main

import (
	"errors"
	"fmt"
	"github.com/yourgfslove/BLOGagregator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

func main() {

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.CurrentUserName)
}

func loginHandler(s *state, cmd command) error {
	if len(cmd.args) < 1 || len(cmd.args) > 2 {
		return errors.New("wrong number of arguments")
	}
	s.cfg.CurrentUserName = cmd.args[0]
	fmt.Println(s.cfg.CurrentUserName + " logged in")
	return nil
}
