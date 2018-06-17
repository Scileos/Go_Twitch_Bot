//Package ChatModeration contains all logic realted to managing the chat.
//This mainly involves removing banned words/phrases and managing user timeouts
package ChatModeration

import (
	"strings"
)

type UserRemovedMessagesCount map[string]int

type ChatModeration interface {
	InitChatModeration()
	IsMessageOk(message string)
	IncrementRemovedMessages(username string)
	GetRemovedMessagesCountForUser(username string)
	DeleteRemovedMessagesCountForUser(username string)
}

//Define banned words/phrases
var BannedWords = []string{} //Banned words here

func InitChatModeration() UserRemovedMessagesCount {
	return make(map[string]int)
}

//Check if message contains any of the banned words
func IsMessageOk(message string) bool {
	for _, bannedWord := range BannedWords {
		if strings.Contains(message, bannedWord) {
			return false
		}
	}
	return true
}

//Increment the count for a users removed messages. This is used to know whether to timeout a user
func (removedMessages *UserRemovedMessagesCount) IncrementRemovedMessages(username string) {
	(*removedMessages)[username]++
}

//Get the current amount of removed messages for a user. If that user has had 0 removed messages then return default.
func (removedMessages UserRemovedMessagesCount) GetRemovedMessagesCountForUser(username string) (int, bool) {
	if val, ok := removedMessages[username]; ok {
		return val, true
	}
	return 0, false
}

//Remove all tracking for a users removed messages. Used after a user is timed out **Potentailly change this later on to an incremental timeout system**
func (removedMessages *UserRemovedMessagesCount) DeleteRemovedMessagesCountForUser(username string) {
	delete(*removedMessages, username)
}
