package analyzer

import (
	"testing"
)

func TestGetTweets(t *testing.T) {
	c := apiClient{TwitterAPIUrl, "", mockClient{"examples/twitter.json"}}
	tweets, err := c.getTweets("trump", "us", "any")
	if err != nil {
		t.Fatal(err)
	}
	expected := text{id: 1039261512488112128,
		text:         "Apple Supplier Shares Slide After \"Backwater\" Trump Tells Tech Giant to Make Products in US ...\nhttps://t.co/dMo44zu4g3",
		textProvider: "twitter"}
	if expected.id != tweets[0].id || expected.text != tweets[0].text || expected.textProvider != tweets[0].textProvider {
		t.Fatalf("%v is not equal to %v", tweets[0], expected)
	}
}
