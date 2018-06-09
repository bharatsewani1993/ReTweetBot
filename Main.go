package main

import (
    "github.com/ChimeraCoder/anaconda"
    "fmt"
    "net/url"
)



func main(){
   //API Key and Access Token
  consumerkey := ""
  consumersecret := ""
  accesstoken := ""
  accesstokensecret := ""

  //Authentication With Twitter
  anaconda.SetConsumerKey(consumerkey)
  anaconda.SetConsumerSecret(consumersecret)
  api := anaconda.NewTwitterApi(accesstoken,accesstokensecret)

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
     tweetinfo, _ := api.Retweet(tweet.Id,true)
     fmt.Printf(tweetinfo.Text)
  }

}
