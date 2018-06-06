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

func extractDuration(hashtag string) (int, int, int) {
	r, _ := regexp.Compile(`(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m?)?`)

	found := r.FindStringSubmatch(hashtag)

	d := 0
	h := 0
	m := 0

	if s, err := strconv.Atoi(found[1]); err == nil {
		d = s
	}

	if s, err := strconv.Atoi(found[2]); err == nil {
		h = s
	}

	if s, err := strconv.Atoi(found[3]); err == nil {
		m = s
	}

	return d, h, m
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

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
	fmt.Printf("User's ACCOUNT: %v\n", user.ScreenName)

	var tweets []twitter.Tweet

	userTimelineParams := &twitter.UserTimelineParams{
		IncludeRetweets: twitter.Bool(true),
	}
	tweets, _, _ = client.Timelines.UserTimeline(userTimelineParams)
	fmt.Println("User's TIMELINE:")
	for _, tweet := range tweets {
		fmt.Println(tweet.Text)
		created, _ := tweet.CreatedAtTime()
		created = created.UTC()
		fmt.Println(tweet.Retweeted)
		day, hour, minute := 0, 0, 0
		for _, hashtag := range tweet.Entities.Hashtags {
			d, h, m := extractDuration(hashtag.Text)
			day += d
			hour += h
			minute += m
		}
		if day == 0 && hour == 0 && minute == 0 {
			fmt.Println("Skipping ...")
			continue
		}
		fmt.Printf("%d %d %d\n", day, hour, minute)
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
		fmt.Println(created)
		fmt.Println(then)
		fmt.Println(now)
		if then.Before(now) {

			statusDestroyParams := &twitter.StatusDestroyParams{}
			client.Statuses.Destroy(tweet.ID, statusDestroyParams)
			fmt.Println("Delete this now")
		} else {
			fmt.Println("Delete in the future")
		}
	}

	// Retweets of Me Timeline
	retweetTimelineParams := &twitter.RetweetsOfMeTimelineParams{
		Count:     2,
		TweetMode: "extended",
	}

	tweets, _, _ = client.Timelines.RetweetsOfMeTimeline(retweetTimelineParams)
	fmt.Println("User's 'RETWEETS OF ME' TIMELINE:")
	for _, tweet := range tweets {
		fmt.Println(tweet.FullText)
	}
}
