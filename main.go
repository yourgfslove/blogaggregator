package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yourgfslove/BLOGagregator/internal/config"
	"github.com/yourgfslove/BLOGagregator/internal/database"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

const TimeLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
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

var s state
var cmds commands

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
	s.cfg = &cfg
	s.db = dbquery
	cmds = commands{make(map[string]func(*state, command) error)}
	cmds.register("login", loginHandler)
	cmds.register("register", registerHandler)
	cmds.register("reset", middlewareLoggedIn(resetHandler))
	cmds.register("getusers", getUsersHandler)
	cmds.register("agg", aggHandler)
	cmds.register("addfeed", middlewareLoggedIn(addFeedHandler))
	cmds.register("feeds", feedsHandler)
	cmds.register("follow", middlewareLoggedIn(followHandler))
	cmds.register("following", middlewareLoggedIn(followingHandler))
	cmds.register("unfollow", middlewareLoggedIn(unfollowHandler))
	cmds.register("getposts", middlewareLoggedIn(GetPostshandler))
	if len(os.Args) < 2 {
		fmt.Println("No commands found")
		os.Exit(1)
	}
	cmd := command{strings.ToLower(os.Args[1]), os.Args[2:]}
	if err = cmds.run(&s, cmd); err != nil {
		log.Fatal(err)
	}
}

func loginHandler(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("usage: <Username>")
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
	if len(cmd.args) != 1 {
		return errors.New("uasge: register <Username>")
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
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

func resetHandler(s *state, cmd command, user database.User) error {
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rss RSSFeed
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		return nil, err
	}
	return &rss, nil
}

func aggHandler(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("usage: agg <Time(2m or 20s or smth)>")
	}
	timebetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("collectin feed every %v\n", timebetweenRequests)
	ticker := time.NewTicker(timebetweenRequests)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		err = scrapeFeed()
		if err != nil {
			return err
		}
	}
}

func addFeedHandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("usage <FeedName> <FeedURL>")
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UserID:    user.ID,
	})

	if err != nil {
		return errors.New("сant create feed")
	}
	fmt.Printf("New feed %s created with URL %s\n", feed.Name, feed.Url)
	url := cmd.args[1]
	followcmd := command{name: "follow", args: []string{url}}
	if err = followHandler(s, followcmd, user); err != nil {
		return err
	}
	return nil
}

func feedsHandler(s *state, cmd command) error {
	feeds, err := s.db.Feeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println(feed.Name, feed.Url, feed.Name_2)
	}
	return nil

}

func followHandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("usage: follow <FeedURl>")
	}
	feed, err := s.db.GetFeedbyurl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s Followed on %s\n", user.Name, feed.Name)
	return nil
}

func followingHandler(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetUsersFollowList(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for _, follow := range follows {
		fmt.Println(follow.Name)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return errors.New("no user logged in")
		}
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func unfollowHandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("usage: Unfollow <FeedURL>")
	}
	feed, err := s.db.GetFeedbyurl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	err = s.db.DeleteFollow(context.Background(), database.DeleteFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	fmt.Printf("%s Unfollowed on %s\n", user.Name, feed.Name)
	return nil
}

func scrapeFeed() error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	rss, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		UpdatedAt:     sql.NullTime{Time: time.Now(), Valid: true},
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:            feed.ID,
	})
	if err != nil {
		return err
	}
	for i := range rss.Channel.Item {
		pubTime, err := time.Parse(TimeLayout, rss.Channel.Item[i].PubDate)
		if err != nil {
			return err
		}
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			PublishedAt: sql.NullTime{Time: pubTime, Valid: true},
			Title:       html.UnescapeString(rss.Channel.Item[i].Title),
			Url:         rss.Channel.Item[i].Link,
			Description: sql.NullString{String: html.UnescapeString(rss.Channel.Item[i].Description), Valid: true},
			FeedID:      feed.ID,
		})

	}
	return nil
}

func GetPostshandler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("usage: GetPosts <Limit>")
	}
	Limit, err := strconv.Atoi(cmd.args[0])
	if err != nil {
		return err
	}
	posts, err := s.db.GetPosts(context.Background(), database.GetPostsParams{
		UserID: user.ID,
		Limit:  int32(Limit),
	})
	for _, post := range posts {
		fmt.Println("========================================")
		fmt.Println("Title:", post.Title, "\nPublished:", post.PublishedAt.Time.UTC().Format(time.DateTime))
		fmt.Println("Description:", post.Description.String)
		fmt.Println("URL:", post.Url)
	}
	return nil
}
