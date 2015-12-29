package commands

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/danjac/podbaby/api"
	"github.com/danjac/podbaby/database"
	"github.com/danjac/podbaby/feedparser"
	"github.com/danjac/podbaby/models"
	"github.com/jmoiron/sqlx"
)

// should be settings
const (
	defaultStaticURL = "/static/"
	defaultStaticDir = "./static/"
	devStaticURL     = "http://localhost:8080/static/"
)

// Serve runs the webserver
func Serve(url string, port int, secretKey, env string) {

	db := database.New(sqlx.MustConnect("postgres", url))

	log := logrus.New()

	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	}

	log.Info("Starting web service...")

	var staticURL string
	if env == "dev" {
		staticURL = devStaticURL
	} else {
		staticURL = defaultStaticURL
	}

	api := api.New(db, log, &api.Config{
		StaticURL: staticURL,
		StaticDir: defaultStaticDir,
		SecretKey: secretKey,
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), api.Handler()); err != nil {
		panic(err)
	}

}

// Fetch retrieves latest podcasts
func Fetch(url string) {

	db := database.New(sqlx.MustConnect("postgres", url))

	log := logrus.New()

	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	}

	log.Info("Starting podcast fetching...")

	channels, err := db.Channels.GetAll()

	if err != nil {
		panic(err)
	}

	for _, channel := range channels {

		result, err := feedparser.Fetch(channel.URL)

		if err != nil {
			log.Error(err)
			continue
		}

		// update channel

		log.Info("Channel:" + channel.Title)

		channel.Title = result.Channel.Title
		channel.Image = result.Channel.Image.Url
		channel.Description = result.Channel.Description

		if err := db.Channels.Create(&channel); err != nil {
			log.Error(err)
			return
		}

		for _, item := range result.Items {
			podcast := &models.Podcast{
				ChannelID:   channel.ID,
				Title:       item.Title,
				Description: item.Description,
			}
			if len(item.Enclosures) == 0 {
				log.Debug("Item has no enclosures")
				continue
			}
			podcast.EnclosureURL = item.Enclosures[0].Url
			pubDate, _ := item.ParsedPubDate()
			podcast.PubDate = pubDate

			log.Info("Podcast:" + podcast.Title)

			if err := db.Podcasts.Create(podcast); err != nil {
				log.Error(err)
				continue
			}
		}

	}

}