package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron"
)

const (
	s3BucketName = "your-s3-bucket-name"
	s3ObjectName = "nytimes-feed.xml"
)

func formatAuthor(author string) string {
	return strings.TrimSpace(strings.Trim(author, "&{}"))
}

func saveToS3(data []byte) error {
	// Create a new AWS session using your credentials
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("your-aws-region"),
		// Add your AWS credentials or use other methods for authentication
		Credentials: credentials.NewStaticCredentials("your-access-key-id", "your-secret-access-key", ""),
	})
	if err != nil {
		return err
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Upload the data to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3ObjectName),
		Body:   aws.ReadSeekCloser(strings.NewReader(string(data))),
	})
	return err
}

func fetchAndSaveFeed() {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://www.nytimes.com/svc/collections/v1/publish/http://www.nytimes.com/topic/subject/agriculture-and-farming/rss.xml")
	if err != nil {
		fmt.Println("Error fetching the feed:", err)
		return
	}
	if feed == nil {
		fmt.Println("Feed is nil.")
		return
	}

	for _, post := range feed.Items {
		author := formatAuthor(post.Author.Name)
		fmt.Println("Title:", post.Title)
		fmt.Println("Link:", post.Link)
		fmt.Println("Description:", post.Description)
		fmt.Println("Author:", author)
		fmt.Println("----")
	}

	feedContent := ""
	if feed != nil {
		feedContent = fmt.Sprintf("%#v", feed)
	}

	err = saveToS3([]byte(feedContent))
	if err != nil {
		fmt.Println("error in saving s3:", err)
		return
	}

	fmt.Println("feed savedin s3 successfully.")
}

func main() {

	c := cron.New()
	c.AddFunc("*/30 * * * * *", fetchAndSaveFeed)
	c.Start()
	// fetchAndSaveFeed()
	select {}
}
