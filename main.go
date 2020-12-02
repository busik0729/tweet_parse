package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"time"

	"log"
	"net/http"
	"os"
)

// point_radius:[30.303898 59.964434 2km]

const SECONDSINMOUNTH = 2592000

type _Twit struct {
	Text       string
	Created_at string
	Twit_id    int64
	Location   *twitter.Place
	User       *twitter.User
	Tags       []twitter.HashtagEntity
}

type _Twits []_Twit

func main() {
	query := flag.String("query", "", "query string for search tweets. For radius search use \"point_radius:[lon lat radius]\". For more information watch https://developer.twitter.com/en/docs/twitter-api/v1/tweets/search/guides/premium-operators")
	flag.Parse()

	if *query == "" {
		log.Println("Необходимо задать параметр -query=")
		os.Exit(1)
	}

	client := auth()
	// a1client := getAuth1Client()

	// getTwits(client)
	// getUsers(a1client)

	getPremS(client, *query)
}

func auth() *twitter.Client {
	config := &clientcredentials.Config{
		ClientID:     "VqZK2vrevJYFsv4ZgqEr8tBzr",
		ClientSecret: "1td3YWNeUvZmu8vqrcLVXJSIBV1JHPV4kzITb6eNHWKjYuFXPj",
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext)

	// Twitter client
	client := twitter.NewClient(httpClient)

	return client
}

func getAuth1Client() *twitter.Client {
	config := oauth1.NewConfig("VqZK2vrevJYFsv4ZgqEr8tBzr", "1td3YWNeUvZmu8vqrcLVXJSIBV1JHPV4kzITb6eNHWKjYuFXPj")
	token := oauth1.NewToken("551821707-1U7zwxy2vG4NeLn1fx8swwp9XwouDF3PKsCy5IPR", "4MjtQpRTWBZi7S54VsfUQ6xCWDMZLfAdqrG78EftMbuah")
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// twitter client
	client := twitter.NewClient(httpClient)

	return client
}

func getTwits(client *twitter.Client) {
	s, _, _ := client.Search.Tweets(&twitter.SearchTweetParams{Query: "", Geocode: "59.962320,,2km", Lang: "en"})

	for _, a := range s.Statuses {
		log.Println(a.Text)
		log.Println(a.Place)
		log.Println(a.Lang)
		log.Println(a.User.Name)
		log.Println(a.CreatedAt)
		log.Println(a.CreatedAtTime())
		log.Println(a.ID)
		log.Println("......................................................................")
	}
}

func getUsers(client *twitter.Client) {
	u, r, e := client.Users.Search("Hatfield", &twitter.UserSearchParams{})
	for _, a := range u {
		log.Println(a.Name)
		log.Println(a.Location)
		log.Println("......................................................................")
	}
	// log.Println(u)
	log.Println(r)
	log.Println(e)
}

func writeInFile(i []byte, file *os.File) {
	f := string(i)
	file.WriteString(f)
}

func getTw(client *twitter.Client, query string, next string) (*twitter.PremiumSearch, *http.Response, error) {
	return client.PremiumSearch.Search30Days(&twitter.PremiumSearchTweetParams{Query: query, Next: next}, "prod")
}

func getPremS(client *twitter.Client, query string) {

	file, err := os.Open("dump.json")
	log.Println(err)
	log.Println(file)
	log.Println("______________")
	if err != nil {
		file, err = os.Create("dump.json")
	}

	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}

	defer file.Close()

	s, r, e := getTw(client, query, "")
	log.Println(r)
	log.Println(s)
	if e != nil {
		log.Println(e)
	}
	writeInFile([]byte("["), file)
	for _, a := range s.Results {
		var t _Twit
		t.Text = a.Text
		t.Created_at = a.CreatedAt
		t.Twit_id = a.ID
		t.User = a.User
		t.Location = a.Place
		t.Tags = a.Entities.Hashtags

		j, _ := json.Marshal(t)
		writeInFile(j, file)
		writeInFile([]byte(","), file)
	}

	log.Println("sleep")
	time.Sleep(2 * time.Second)

	i := 0
	if len(s.Next) != 0 {
		for {
			s, _, e := getTw(client, query, s.Next)
			if e != nil {
				log.Println(e)
			}
			for _, a := range s.Results {
				var t _Twit
				t.Text = a.Text
				t.Created_at = a.CreatedAt
				t.Twit_id = a.ID
				t.User = a.User
				t.Location = a.Place
				t.Tags = a.Entities.Hashtags

				j, _ := json.Marshal(t)
				writeInFile(j, file)
				writeInFile([]byte(","), file)
			}
			log.Println(s.Next)
			if len(s.Next) == 0 {
				log.Println("not Next")
				break
			}
			i++

			log.Println("sleep")
			time.Sleep(5 * time.Second)
		}
	}
	log.Println(i)

	writeInFile([]byte("]"), file)
}
