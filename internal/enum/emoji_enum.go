package enum

type Emoji int

const (
	EmojiThumbsUp Emoji = iota
	EmojiThumbsDown
	EmojiLaugh
	EmojiSad
	EmojiAngry
)

var emojiToString = map[Emoji]string{
	EmojiThumbsUp:   "thumbs_up",
	EmojiThumbsDown: "thumbs_down",
	EmojiLaugh:      "laugh",
	EmojiSad:        "sad",
	EmojiAngry:      "angry",
}

func (e Emoji) String() string {
	return emojiToString[e]
}

func EmojiFromString(s string) (Emoji, bool) {
	for k, v := range emojiToString {
		if v == s {
			return k, true
		}
	}
	return EmojiThumbsUp, false // default to EmojiThumbsUp if not found
}

func EmojiFromInt(i int) (Emoji, bool) {
	for k := range emojiToString {
		if int(k) == i {
			return k, true
		}
	}
	return EmojiThumbsUp, false // default to EmojiThumbsUp if not found
}
