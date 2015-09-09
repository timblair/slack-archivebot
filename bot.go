package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nlopes/slack"
)

func main() {
	slackToken := os.Getenv("ARCHIVEBOT_SLACK_TOKEN")
	api := slack.New(slackToken)
	//api.SetDebug(true)

	channels, err := api.GetChannels(true)
	if err != nil {
		log.Printf("Error when loading channels: %s\n", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(c []slack.Channel) {
		defer wg.Done()
		archiveEmpty(api, c)
	}(channels)

	go func(c []slack.Channel) {
		defer wg.Done()
		archiveInactive(api, c)
	}(channels)

	wg.Wait()
}

func archiveEmpty(api *slack.Slack, c []slack.Channel) {
	for _, channel := range c {
		if channel.NumMembers == 0 {
			fmt.Printf("Archiving empty channel #%s (%s)\n", channel.Name, channel.Id)
			if err := api.ArchiveChannel(channel.Id); err != nil {
				log.Printf("Error archiving #%s (%s): %s\n", channel.Name, channel.Id, err)
			}
		}
	}
}

func archiveInactive(api *slack.Slack, c []slack.Channel) {
	inactive := inactiveChannels(api, c)
	for _, channel := range inactive {
		fmt.Printf("Archiving #%s (%s) due to inactivity\n", channel.Name, channel.Id)
		if err := api.ArchiveChannel(channel.Id); err != nil {
			log.Printf("Error archiving #%s (%s): %s\n", channel.Name, channel.Id, err)
		}
	}
}

type LastChannelMessage struct {
	Channel   slack.Channel
	Timestamp int64
}

func inactiveChannels(api *slack.Slack, c []slack.Channel) []slack.Channel {
	inactiveDays, _ := strconv.ParseInt(os.Getenv("ARCHIVEBOT_INACTIVE_DAYS"), 10, 32)
	if inactiveDays == 0 {
		inactiveDays = 30
	}

	timeout := int64(time.Now().Unix()) - (86400 * inactiveDays)
	channels := []slack.Channel{}

	res := make(chan LastChannelMessage)
	for _, channel := range c {
		go func(channel slack.Channel) {
			timestamp, _ := lastMessageTimestamp(api, channel)
			res <- LastChannelMessage{Channel: channel, Timestamp: timestamp}
		}(channel)
	}

	for i := 0; i < len(c); i++ {
		lcm := <-res
		if lcm.Timestamp > 0 && lcm.Timestamp < timeout {
			channels = append(channels, lcm.Channel)
		}
	}

	close(res)
	return channels
}

func lastMessageTimestamp(api *slack.Slack, channel slack.Channel) (int64, error) {
	var latest string

	for {
		historyParams := slack.HistoryParameters{Count: 5}
		if latest != "" {
			historyParams.Latest = latest
		}

		history, err := api.GetChannelHistory(channel.Id, historyParams)

		if err != nil {
			return -1, err
		}

		if len(history.Messages) == 0 {
			return -1, nil
		}

		for _, msg := range history.Messages {
			latest = msg.Msg.Timestamp

			if msg.SubType != "channel_join" && msg.SubType != "channel_leave" {
				msgStamp := strings.Split(msg.Msg.Timestamp, ".")
				if timestamp, err := strconv.ParseInt(msgStamp[0], 10, 32); err == nil {
					return timestamp, nil
				}
			}
		}
	}
}
