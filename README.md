# slack-archivebot

A simple Go app that will automatically archive public channels on the
following criteria:

* if the channel is empty; or
* if the channel has had no activity for the last X days.

## Use

The poject uses [`godep`](https://github.com/tools/godep) to manage
dependencies, so you'll need that.  Once you've cloned this repo into your
`$GOPATH`:

```sh
cd path/to/slack-archivebot
godep go build
./slack-archivebot
```

## Deployment

Heroku is the simplest option.  The script can run quite happily on a free dyno
using the Heroku Scheduler add-on.

[![Deploy to Heroku](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/timblair/slack-archivebot)

Note: the above will deploy the app to your Heroku account, and add the
Scheduler add-on, but _won't_ configure it to run.  To do this, go to your
[dashboard](https://dashboard.heroku.com/apps), find the appropriate app, open
the Scheduler add-on, and add a new job that runs `slack-archivebot` every 10
minutes.

## Configuration

The following environment variables are used to configure the script:

* `ARCHIVEBOT_SLACK_TOKEN`: the Slack [Web API key](https://api.slack.com/web)
  for your team.
* `ARCHIVEBOT_INACTIVITY_DAYS`: the number of days' inactivity after which to
  archive a channel (default: `30`).
* `ARCHIVEBOT_NOTIFY`: a Slack user or channel (e.g. `#general` or `@tblair`) to notify when something goes wrong.

Note: you must use an API key for a regular Slack user account.  You _cannot_
use a bot user account, because bot users don't have permission to archive
channels.

## Licensing and Attribution

slack-archivebot is released under the MIT license as detailed in the LICENSE
file that should be distributed with this library; the source code is [freely
available](http://github.com/timblair/slack-archivebot).

slack-archivebot was developed by [Tim Blair](http://tim.bla.ir/) during a
[Venntro](http://venntro.com/) hack day.
