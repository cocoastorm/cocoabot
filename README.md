# cocoabot
A relatively simple music bot for Discord.

Huge shout out to [SpeakerBot](https://github.com/dustinblackman/speakerbot)
for most of the discord musical magic of which inspiration was hugely taken from.

Huge thanks to [DiscordGo](https://github.com/bwmarrin/discordgo) for the painless Discord API Go Bindings.

## Building

cocoabot relies on [FFmpeg](https://ffmpeg.org/) for audio encoding. In particular, it relies on the opus codec, please have FFmpeg compiled with `--enable-libopus`.

```
go get
go build
```

## Usage

After building:

```
BOT_TOKEN=<discord bot token> \
YOUTUBE_KEY=<youtube api key> \
./cocoabot
```

The YouTube API Key is solely used for searching and retrieving videos from a playlist.

## Contribution
If you notice any bugs or want a new feature, PRs are more than welcome. :)

## LICENSE
[AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.en.html)
