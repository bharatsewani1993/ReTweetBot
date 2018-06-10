package main

import (
    "github.com/ChimeraCoder/anaconda"
    "fmt"
    "net/url"
    "reflect"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "strconv"
    "strings"
)


//login to twitter
func twitterlogin() (*anaconda.TwitterApi) {
  //API Key and Access Token
  consumerkey := ""
  consumersecret := ""
  accesstoken := ""
  accesstokensecret := ""

 //Authentication With Twitter
 anaconda.SetConsumerKey(consumerkey)
 anaconda.SetConsumerSecret(consumersecret)
 return anaconda.NewTwitterApi(accesstoken,accesstokensecret)
}


//search and retweet
func searchandretweet(api *anaconda.TwitterApi){
  //define search term
  options := url.Values{}
  options.Set("count","150")
  options.Set("lang","en")
  options.Set("result_type","recent")
  options.Set("include_entities","false")

  //Search Query
  SearchResult, _ := api.GetSearch("#golangjobs",options)

  //loop through search result
  for _, tweet := range SearchResult.Statuses{
     fmt.Printf("Tweet %v\n",tweet.Text)
     //ReTweet
    // tweetinfo, _ := api.Retweet(tweet.Id,true)
     //fmt.Printf(tweetinfo.Text)
  }
}

func followbackusers(api *anaconda.TwitterApi, db *sql.DB){
    //get list of followers
    options := url.Values{}
    options.Set("skip_status","true")
    options.Set("include_user_entities","false")
    options.Set("count","200")

    //get all followers from Twitter
    followerlist := <-api.GetFollowersIdsAll(options)

    //check for new followers by comparing with existing in database
      followersindb := getfollowersfromdb(db)
      fmt.Printf("%v",followersindb)

    //loop through follower list
    options.Del("skip_status")
    options.Del("include_user_entities")
    options.Del("count")
    options.Set("follow","false")
    for _, list := range followerlist.Ids{
      //fmt.Printf("User Id %v\n",list)
      //follow back users
      api.FollowUserId(list,options)
    }
    //add users to database
    storefollowersid(followerlist.Ids,db)
}

//initialize database connection
func dbconnection() (*sql.DB) {
  db, _ := sql.Open("mysql","root:P3NT3ST3R@/retweetbot")
  return db
}

func storefollowersid(followerids []int64, db *sql.DB){
   //genrate query string
     querystring := "("
     for _, followerid := range followerids {
       s := strconv.FormatInt(followerid,10)
       querystring += s + "),("
     }
     querystring = strings.TrimRight(querystring, ",(")
     fmt.Println("Query String %v\n",querystring)

    //prepare batch insert statement
      insertqry, _ := db.Prepare("INSERT into followers values " + querystring)
      fmt.Println("Query to execute %v\n",insertqry)
}

func getfollowersfromdb(db *sql.DB) ([]int64){
   var followersindb []int64
   rows, _ := db.Query("SELECT * from followers where active = '1'")
   for rows.Next(){
     var followerid int64
       _ = rows.Scan(&followerid)
      followersindb = append(followersindb,followerid)
   }
   return followersindb
}

func main(){
  api := twitterlogin()
  db := dbconnection()
  searchandretweet(api)
  followbackusers(api,db)
  fmt.Println(reflect.TypeOf(api))
}
