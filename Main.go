package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/go-sql-driver/mysql"
)

//login to twitter
func twitterlogin() *anaconda.TwitterApi {
	//API Key and Access Token
	consumerkey := ""
	consumersecret := ""
	accesstoken := ""
	accesstokensecret := ""

	//Authentication With Twitter
	anaconda.SetConsumerKey(consumerkey)
	anaconda.SetConsumerSecret(consumersecret)
	return anaconda.NewTwitterApi(accesstoken, accesstokensecret)
}

//search and retweet
func searchandretweet(api *anaconda.TwitterApi) {
	//define search term
	options := url.Values{}
	options.Set("count", "150")
	options.Set("lang", "en")
	options.Set("result_type", "recent")
	options.Set("include_entities", "false")

	//Search Query
	SearchResult, err := api.GetSearch("#golangjobs", options)
	if err != nil {
     fmt.Printf("Error: %v\n",err)
	}
	//loop through search result
	for _, tweet := range SearchResult.Statuses {
		fmt.Printf("Tweet %v\n", tweet.Text)
		//ReTweet
		// tweetinfo, _ := api.Retweet(tweet.Id,true)
		//fmt.Printf(tweetinfo.Text)
	}
}

func followbackusers(api *anaconda.TwitterApi, db *sql.DB) {
	//get list of followers
	options := url.Values{}
	options.Set("skip_status", "true")
	options.Set("include_user_entities", "false")
	options.Set("count", "200")

	//get all followers from Twitter
	followerlist := <-api.GetFollowersIdsAll(options)

	//check for new followers by comparing with existing in database
	followersindb := getfollowersfromdb(db)
	fmt.Printf("%v", followersindb)

	//loop through follower list
	options.Del("skip_status")
	options.Del("include_user_entities")
	options.Del("count")
	options.Set("follow", "false")
	for _, list := range followerlist.Ids {
		//fmt.Printf("User Id %v\n",list)
		//follow back users
		_, err := api.FollowUserId(list, options)
    if err != nil {
      fmt.Printf("Error: %v\n", err)
    }
  }
	//add users to database
	storefollowersid(followerlist.Ids, db)
}

//initialize database connection
func dbconnection() *sql.DB {
	db, err := sql.Open("mysql", "root:P3NT3ST3R@/retweetbot")
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	return db
}

func storefollowersid(followerids []int64, db *sql.DB) {
	//genrate query string
	querystring := "("
	for _, followerid := range followerids {
		s := strconv.FormatInt(followerid, 10)
		querystring += s + "),("
	}
	querystring = strings.TrimRight(querystring, ",(")
	fmt.Printf("Query String %v\n", querystring)

	//prepare batch insert statement
	insertqry, err := db.Prepare("INSERT into followers values ?")
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
	fmt.Printf("Query to execute %v\n", insertqry)
}

func getfollowersfromdb(db *sql.DB) []int64 {
	var followersindb []int64
	rows, err := db.Query("SELECT * from followers where active = '1'")
	if err != nil {
     fmt.Printf("Error: %v\n",err)
	}
	for rows.Next() {
		var followerid int64
		err = rows.Scan(&followerid)
    if err != nil {
      fmt.Printf("Error: %v\n",err)
    }
		followersindb = append(followersindb, followerid)
	}
	return followersindb
}

func main() {
	api := twitterlogin()
	db := dbconnection()
	searchandretweet(api)
	followbackusers(api, db)
	fmt.Println(reflect.TypeOf(api))
}
