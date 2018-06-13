package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
)

func extractInt(value string) int {
	if s, err := strconv.Atoi(value); err == nil {
		return s
	}
	return 0
}

func extractDuration(hashtag string) (int, int, int) {
	d, h, m := 0, 0, 0

	r, _ := regexp.Compile(`(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m?)?`)

	found := r.FindStringSubmatch(hashtag)

	d = extractInt(found[1])
	h = extractInt(found[2])
	m = extractInt(found[3])

	return d, h, m
}

func main() {
	godotenv.Load()

	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")

	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	flags.Parse(os.Args[1:])

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)

	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	user, _, _ := client.Accounts.VerifyCredentials(verifyParams)
	fmt.Printf("Username: %v\n", user.ScreenName)

	var tweets []twitter.Tweet

	userTimelineParams := &twitter.UserTimelineParams{
		IncludeRetweets: twitter.Bool(true),
		ExcludeReplies:  twitter.Bool(false),
		TweetMode:       "extended",
	}

	tweets, _, _ = client.Timelines.UserTimeline(userTimelineParams)

	fmt.Println("User's timeline:")
	for _, tweet := range tweets {
		created, _ := tweet.CreatedAtTime()
		created = created.UTC()

		day, hour, minute := 0, 0, 0

		if tweet.Retweeted {
			day = 1
		} else {
			for _, hashtag := range tweet.Entities.Hashtags {
				d, h, m := extractDuration(hashtag.Text)
				day += d
				hour += h
				minute += m
			}
		}

		if day == 0 && hour == 0 && minute == 0 {
			fmt.Println("Skipping ...")
			continue
		}

		now := time.Now().UTC()
		then := time.Date(
			created.Year(),
			created.Month(),
			created.Day()+day,
			created.Hour()+hour,
			created.Minute()+minute,
			created.Second(),
			created.Nanosecond(),
			time.UTC)

		if then.Before(now) {
			if tweet.Retweeted {
				statusUnretweetParams := &twitter.StatusUnretweetParams{}
				client.Statuses.Unretweet(tweet.ID, statusUnretweetParams)
			} else {
				statusDestroyParams := &twitter.StatusDestroyParams{}
				client.Statuses.Destroy(tweet.ID, statusDestroyParams)
			}
			fmt.Println("Delete this now")
		}
	}
}
