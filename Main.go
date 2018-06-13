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
	fmt.Printf("Followers from Twitter: %v\n",followerlist)

	//check for new followers by comparing with existing in database
	followersindb := getfollowersfromdb(db)
	fmt.Printf("Followers in DB: %v\n", followersindb)

	//compare existing followers with new followers
	/*  for i, followerid := range followerlist.Ids {
			 for j, followeriddb := range followersindb {
				  if followerid == followeriddb {
						followersindb = append(followersindb[:j], followersindb[j+1:]...)
						followerlist.Ids = append(followerlist.Ids[:i], followerlist.Ids[i+1:]...)
						 break;
					}
			 }
		}
		*/

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
				updatequery += s + " then 1 when"
			}
			updatequery = strings.TrimRight(updatequery," when")
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

func main() {
	api := twitterlogin()
	db := dbconnection()
	searchandretweet(api)
	followbackusers(api, db)
  greetusers(api,db)
	fmt.Println(reflect.TypeOf(api))
}
