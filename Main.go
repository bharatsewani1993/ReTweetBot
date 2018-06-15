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
	consumerkey := "mkepp9H65kniGdBbYoaH3qmrM"
  consumersecret := "5T9WfXw4wqlPIDy8YbDVttS2jIiWGvJ3934iPc15jr1N6JL7NN"
  accesstoken := "1005393626967363585-3DhoYPWQMblzFgwTcYSqJ33Vm1JJBi"
  accesstokensecret := "cio0aTlMObZlwrNhqaZS8EFyuZ25jdLnVPhWPJ1iX7yaj"

	//Authentication With Twitter
	anaconda.SetConsumerKey(consumerkey)
	anaconda.SetConsumerSecret(consumersecret)
	return anaconda.NewTwitterApi(accesstoken, accesstokensecret)
}

//search and retweet
func searchandretweet(api *anaconda.TwitterApi, db *sql.DB) {
	//define search term
	options := url.Values{}
	options.Set("count", "150")
	options.Set("lang", "en")
	options.Set("result_type", "recent")
	options.Set("include_entities", "false")

	//Search Query
	SearchResult, err := api.GetSearch("#golang #job OR #golangjob", options)
	if err != nil {
     fmt.Printf("Error: %v\n",err)
	}

	//tweeter tweets id
	  var newtweetid []int64
		for _, tweetid := range SearchResult.Statuses{
			newtweetid = append(newtweetid,tweetid.Id)
		}

	//Already done tweets
	 pasttweets := gettweetsfromdb(db)

	 //find new tweets to do.
	  tweetstodo := difference(newtweetid,pasttweets)


	//loop through search result
	for _, tweet := range SearchResult.Statuses {
		 for _, tweets := range tweetstodo {
			  if tweets == tweet.Id {
					//fmt.Printf("Tweet %v\n", tweet.Text)
					//ReTweet
					 tweetinfo, _ := api.Retweet(tweet.Id,true)
					 fmt.Printf(tweetinfo.Text)
				}
		 }
	}

	if tweetstodo != nil {
		updatetweetlist(tweetstodo,db)
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
	fmt.Printf("Followers from Twitter: %v\n",followerlist)

	//check for new followers by comparing with existing in database
	followersindb := getfollowersfromdb(db)
	fmt.Printf("Followers in DB: %v\n", followersindb)

	 //find new followers by comparint with existing.
	 userlist := difference(followerlist.Ids,followersindb)
	 fmt.Printf("New Followers: %v\n",userlist)

  fmt.Printf("Followers from TWitter: %v\n",followerlist.Ids)

	//loop through follower list
	options.Del("skip_status")
	options.Del("include_user_entities")
	options.Del("count")
	options.Set("follow", "false")

		for _, list := range userlist {
		//fmt.Printf("User Id %v\n",list)
		//follow back users
		_, err := api.FollowUserId(list, options)
    if err != nil {
      fmt.Printf("Error: %v\n", err)
    }

  }
	//add users to database
	if userlist != nil{
		storefollowersid(userlist, db)
	}

	//unfollow the unfollowers
	 followersindb = getfollowersfromdb(db)
	 userlist = difference(followersindb,followerlist.Ids)
	 for _, list := range userlist {
 	 fmt.Printf("Unfollow user %v\n",list)
 	 //Unfollow users
 	 _, err := api.UnfollowUserId(list)
 	 if err != nil {
 		 fmt.Printf("Error: %v\n", err)
 	 }
  }
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
	querystring := "INSERT into followers (followerid) values ("
	for _, followerid := range followerids {
		s := strconv.FormatInt(followerid, 10)
		querystring += s + "),("
	}
	querystring = strings.TrimRight(querystring, ",(")
	fmt.Printf("Query String %v\n", querystring)

	//prepare batch insert statement
	insertqry, err := db.Prepare(querystring)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
  fmt.Printf("Query to execute %v\n", insertqry)

	_, err = db.Exec(querystring)
	if err != nil {
		fmt.Printf("Error %v\n",err)
	}

}

func getfollowersfromdb(db *sql.DB) []int64 {
	var followersindb []int64
	rows, err := db.Query("SELECT followerid from followers where active = 1")
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

func greetusers(api *anaconda.TwitterApi, db *sql.DB){
	   followerlist := getungreetedusersfromdb(db)
		 var dm string = "Hi There,\n Thank you for following me. I'll keep you posted with the latest #GolangJobs"
		 //send dm to users
		 for _, userid := range followerlist {
			  _,err := api.PostDMToUserId(dm,userid)
				if err != nil {
					fmt.Printf("Error %v\n",err)
				}
		 }
		 if followerlist != nil {
			 updategreetedusers(followerlist,db)
		 }

}

func getungreetedusersfromdb(db *sql.DB) ([]int64){
	 var users []int64
	 rows, err := db.Query("select followerid from followers where greet = 0")
	 if err != nil {
		 fmt.Printf("Error: %v\n",err)
	 }
	 for rows.Next(){
		 var followerid int64
		 err = rows.Scan(&followerid)
		 if err != nil {
			 fmt.Printf("Error: %v\n",err)
		 }
		 users = append(users,followerid)
	 }
	 return users
}

func difference(twitteruserlist []int64, dbuserlist []int64) ([]int64) {
	  var diff []int64
		for _, twitterusers := range twitteruserlist {
			  found := false
				for _, dbusers := range dbuserlist {
					 if twitterusers == dbusers {
						 	found = true
							break
					 }
				}
				if !found {
					diff = append(diff,twitterusers)
				}
		}
		return diff;
}

func updategreetedusers(followerlist []int64, db *sql.DB) (bool){
	//creating bulk update query
	    updatequery := "UPDATE followers SET greet = (CASE followerid when "
		  for _, followerid := range followerlist {
				s := strconv.FormatInt(followerid, 10)
				updatequery += s + " then 1 when "
			}
			updatequery = strings.TrimRight(updatequery," when ")
			updatequery += " END ) where followerid in ("
			for _, followerid := range followerlist {
				s := strconv.FormatInt(followerid, 10)
				updatequery += s + ","
			}
			updatequery = strings.TrimRight(updatequery, ",")
			updatequery = updatequery + ")"
			//fmt.Println("Update Query: %v\n",updatequery)

			stmt, err := db.Prepare(updatequery)
			if err != nil {
				fmt.Printf("Error %v\n", err)
			}

			res, err := stmt.Exec()
			if err != nil {
				fmt.Println("Error: %v\n", err)
			}

			affect, err := res.RowsAffected()
			if err != nil {
				fmt.Println("Error: %v\n", err)
			}

			fmt.Println("Effected Rows: ",affect)
			return true
}

//get past tweets from db
func gettweetsfromdb(db *sql.DB) ([]int64){
	var tweetlist []int64
	rows, err := db.Query("select tweetid from tweets where active = 1")
	if err != nil{
		fmt.Printf("Error: %v\n",err)
	}
	for rows.Next(){
		var tweetid int64
		err := rows.Scan(&tweetid)
		if err != nil {
			fmt.Printf("Error %v\n",err)
		}
		tweetlist = append(tweetlist,tweetid)
	}
	return tweetlist
}

//update db with tweet list
func updatetweetlist(tweetlist []int64, db *sql.DB){
	query := "Insert into tweets (tweetid) values ("
	for _, tweetid := range tweetlist {
		 s := strconv.FormatInt(tweetid,10)
		 query += s + "),("
	}
	query = strings.TrimRight(query,",(")

	statment, err := db.Prepare(query);
	if err != nil{
		fmt.Printf("Error %v\n",err)
	}
	_, err = statment.Exec()
	if err != nil {
		fmt.Printf("Error %v\n",err)
	}
}

func main() {
	api := twitterlogin()
	db := dbconnection()
	searchandretweet(api,db)
	followbackusers(api, db)
  greetusers(api,db)
	fmt.Println(reflect.TypeOf(api))
}
