package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

func main() {
	slackToken := os.Getenv("ARCHIVEBOT_SLACK_TOKEN")
	api := slack.New(slackToken)
	//api.SetDebug(true)

	channels, err := emptyChannels(api)

	if err != nil {
		fmt.Printf("Error when processing empty channels: %s\n", err)
		return
	}

	for _, channel := range channels {
		fmt.Printf("Archiving empty channel #%s (%s)\n", channel.Name, channel.Id)
		err = api.ArchiveChannel(channel.Id)

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
	}

	channels, err = inactiveChannels(api)

	if err != nil {
		log.Panicf("Error when processing inactive channels: %s\n", err)
	}

	for _, channel := range channels {
		fmt.Printf("Archiving #%s (%s) due to inactivity\n", channel.Name, channel.Id)
		err = api.ArchiveChannel(channel.Id)

		if err != nil {
			fmt.Printf("Couldn't archive #%s (%s): %s\n", channel.Name, channel.Id, err)
		}
	}
}

func emptyChannels(api *slack.Slack) ([]slack.Channel, error) {
	channels := []slack.Channel{}

	allChannels, err := api.GetChannels(true)
	if err != nil {
		return nil, err
	}

	for _, channel := range allChannels {
		if channel.NumMembers == 0 {
			channels = append(channels, channel)
		}
	}

	return channels, nil
}

type LastChannelMessage struct {
	Channel   slack.Channel
	Timestamp int64
}

func inactiveChannels(api *slack.Slack) ([]slack.Channel, error) {
	inactiveDays, _ := strconv.ParseInt(os.Getenv("ARCHIVEBOT_INACTIVE_DAYS"), 10, 32)
	if inactiveDays == 0 {
		inactiveDays = 30
	}

	timeout := int64(time.Now().Unix()) - (86400 * inactiveDays)
	channels := []slack.Channel{}

	allChannels, err := api.GetChannels(true)
	if err != nil {
		return nil, err
	}

	res := make(chan LastChannelMessage)
	for _, channel := range allChannels {
		go func(channel slack.Channel) {
			timestamp, _ := lastMessageTimestamp(api, channel)
			res <- LastChannelMessage{Channel: channel, Timestamp: timestamp}
		}(channel)
	}

	for i := 0; i < len(allChannels); i++ {
		lcm := <-res
		if lcm.Timestamp > 0 && lcm.Timestamp < timeout {
			channels = append(channels, lcm.Channel)
		}
	}

	close(res)
	return channels, nil
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
