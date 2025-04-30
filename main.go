package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yourgfslove/BLOGagregator/internal/config"
	"github.com/yourgfslove/BLOGagregator/internal/database"
	"log"
	"os"
	"time"
)

type state struct {
	db  *database.Queries
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
		log.Fatal(err)
	}
	dbconn, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbquery := database.New(dbconn)
	var s state
	s.cfg = &cfg
	s.db = dbquery
	cmds := commands{make(map[string]func(*state, command) error)}
	cmds.register("login", loginHandler)
	cmds.register("register", registerHandler)
	cmds.register("reset", resetHandler)
	cmds.register("GetUsers", getUsersHandler)
	if len(os.Args) < 2 {
		fmt.Println("to many commands")
		os.Exit(1)
	}
	cmd := command{os.Args[1], os.Args[2:]}
	if err = cmds.run(&s, cmd); err != nil {
		log.Fatal(err)
	}
}

func loginHandler(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("wrong number of arguments")
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return errors.New("user not found")
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Println(s.cfg.CurrentUserName + " logged in")
	return nil
}

func registerHandler(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("wrong number of arguments")
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		return errors.New("user already exists")
	}
	user, err = s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Name:      cmd.args[0],
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Println(user.Name + " registered")
	return nil
}

func resetHandler(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("reset successfully")
	return nil
}

func getUsersHandler(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return errors.New("no users found")
	}
	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Println(s.cfg.CurrentUserName + " (current)")
		} else {
			fmt.Println(user.Name)
		}
	}
	return nil
}
