package enum

import "errors"

type Vote int

const (
	VoteDown Vote = -1 // -1
	VoteUp   Vote = 1  // +1
)

func GetVoteFromValue(value int) (Vote, error) {
	switch value {
	case -1:
		return VoteDown, nil
	case 1:
		return VoteUp, nil
	default:
		return 0, errors.New("Value is not supported.") // default to VoteDown if not found
	}
}
