package main

import (
	"encoding/csv"
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
	s3ObjectName = "nytimes-feed.csv"
)

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
	feed, err := fp.ParseURL("https://www.thehindubusinessline.com/economy/agri-business/feeder/default.rss")
	if err != nil {
		fmt.Println("Error fetching the feed:", err)
		return
	}
	if feed == nil {
		fmt.Println("Feed is nil.")
		return
	}

	csvData := convertToCSV(feed)
	err = saveToS3([]byte(csvData))
	if err != nil {
		fmt.Println("Error saving to S3:", err)
		return
	}

	fmt.Println("Feed data saved to S3 successfully.")
}

func convertToCSV(feed *gofeed.Feed) string {
	var csvRows [][]string

	header := []string{"Title", "Link", "Description", "Author"}
	csvRows = append(csvRows, header)

	// Data rows
	for _, post := range feed.Items {

		// Add data to the CSV row
		row := []string{
			post.Title,
			post.Link,
			post.Description,
			post.Published,
		}
		csvRows = append(csvRows, row)
	}

	// Convert to CSV string
	var csvData strings.Builder
	w := csv.NewWriter(&csvData)
	w.WriteAll(csvRows)
	w.Flush()

	return csvData.String()
}

func main() {

	c := cron.New()
	c.AddFunc("*/30 * * * * *", fetchAndSaveFeed)
	c.Start()
	fetchAndSaveFeed()
	select {}
}
