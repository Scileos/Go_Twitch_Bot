//Package main conmtains all the executables for out project
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"time"

	"./ChatModeration"
	"./Config"
)

//Define data structure for bot
type Bot struct {
	server     string
	port       string
	nickname   string
	channel    string
	connection net.Conn
}

var chatLimit = (20 / 30)

// Regex for parsing PRIVMSG strings.
//
// First matched group is the user's name and the second matched group is the content of the
// user's message.
var msgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

// Regex for parsing user commands, from already parsed PRIVMSG strings.
//
// First matched group is the command name and the second matched group is the argument for the
// command.
var cmdRegex *regexp.Regexp = regexp.MustCompile(`^!(\w+)\s?(\w+)?`)

//Initialise bot with relevant channel information
func CreateBot() *Bot {
	return &Bot{
		server:     "irc.chat.twitch.tv",
		port:       "6667",
		nickname:   "", //MUST BE IN LOWERCASE
		connection: nil,
		channel:    "", //MUST BE IN LOWERCASE
	}
}

//Connects to the default twitch server.
func (twitchBot *Bot) Connect() {
	var err error
	twitchBot.connection, err = net.Dial("tcp", twitchBot.server+":"+twitchBot.port)
	if err != nil {
		fmt.Printf("There were errors conencting to the twitch server: %v, trying again", err)
		twitchBot.Connect()
		return
	}
	fmt.Println("Connected to: " + twitchBot.server + ": " + twitchBot.channel)
}

//Listens for chat events and passes the username and message string into our channel to be actioned by the writer.
func (twitchBot *Bot) chatListener(c chan []string, removedMessagesMap *ChatModeration.UserRemovedMessagesCount) {

	fmt.Printf("Listenting on %s\n", twitchBot.channel)

	tp := textproto.NewReader(bufio.NewReader(twitchBot.connection))

	for {
		line, err := tp.ReadLine()

		if err != nil {
			twitchBot.connection.Close()
			fmt.Println("Failed to read line from channel. Disconnected")
			break
		}
		if line == "PING :tmi.twitch.tv" {
			twitchBot.connection.Write([]byte("PONG :tmi.twitch.tv\r\n"))
		} else {
			matches := msgRegex.FindStringSubmatch(line)
			if matches != nil {
				userName := matches[1]
				msgType := matches[2]

				switch msgType {
				case "PRIVMSG":
					msg := matches[3]

					if ChatModeration.IsMessageOk(msg) {
						cmdMatches := cmdRegex.FindStringSubmatch(msg)
						if cmdMatches != nil {
							cmd := cmdMatches[1]
							c <- []string{userName, cmd}
						}
					} else {
						userRemovedMessageCount, ok := removedMessagesMap.GetRemovedMessagesCountForUser(userName)
						if ok && userRemovedMessageCount > 3 {
							twitchBot.Timeout(userName, 300)
							twitchBot.connection.Write([]byte("PRIVMSG " + twitchBot.channel + " :3 strikes and you're out! " + userName + " will be back in 5 \r\n"))
							removedMessagesMap.DeleteRemovedMessagesCountForUser(userName)
						} else {
							twitchBot.Timeout(userName, 1)
							twitchBot.connection.Write([]byte("PRIVMSG " + twitchBot.channel + " :Woah there " + userName + ". Let's keep it clean - Any more and you'll be timed out. \r\n"))
							removedMessagesMap.IncrementRemovedMessages(userName)
						}
					}
				default:
					//do nothing
				}
			}
		}
	}
}

//Receives data from the channel and acts upon relevant commands
func (twitchBot *Bot) chatWriter(c chan []string) {
	channelData := <-c

	//Used for broadcaster commands, could potentially move to moderator commands
	username := channelData[0]
	command := channelData[1]

	//Handle commands
	switch command {
	case "quit":
		if username == twitchBot.channel[1:] {
			fmt.Printf("Quit command received. Shutting down")
			twitchBot.connection.Close()
			os.Exit(1)
		} else {
			//Say you have invalid permissions
		}
	case "timeout":
		//call timeout function here
	default:
		//do nothing
	}

	time.Sleep(time.Duration(1 / chatLimit))
}

//Timeout a given user for a specified amount of time.
func (twitchBot *Bot) Timeout(username string, numberOfSeconds int) {
	//timeout user, if the number of seconds is low this acts as removing their message
	twitchBot.connection.Write([]byte("PRIVMSG " + twitchBot.channel + " : /timeout " + username + " " + strconv.Itoa(numberOfSeconds) + "\r\n"))

}

//Main entry point for execution
func main() {

	twitchBot := CreateBot()
	config := Config.InitConfig()
	removedMessagesMap := ChatModeration.InitChatModeration()
	pass, ok := config.GetConfigValue("twitch_bot_token")

	c := make(chan []string, 10)

	if !ok {
		fmt.Println("Error reading password")
		twitchBot.connection.Close()
		os.Exit(1)
	}

	twitchBot.Connect()

	twitchBot.connection.Write([]byte("PASS " + pass + "\r\n"))
	twitchBot.connection.Write([]byte("NICK " + twitchBot.nickname + "\r\n"))
	twitchBot.connection.Write([]byte("JOIN " + twitchBot.channel + "\r\n"))

	go twitchBot.chatListener(c, &removedMessagesMap)
	go twitchBot.chatWriter(c)

	for {
	}
}
