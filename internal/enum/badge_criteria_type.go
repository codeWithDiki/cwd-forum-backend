package enum

type BadgeCriteriaType int

const (
	BadgeCriteriaPostCount BadgeCriteriaType = iota
	BadgeCriteriaLikeCount
	BadgeCriteriaCommentCount
	BadgeCriteriaThreadCount
	BadgeCriteriaReactionCount
	BadgeCriteriaSolutionPostCount
	BadgeCriteriaUpVoteCount
	BadgeCriteriaDownVoteCount
)

var badgeCriteriaTypeToString = map[BadgeCriteriaType]string{
	BadgeCriteriaPostCount:         "post_count",
	BadgeCriteriaLikeCount:         "like_count",
	BadgeCriteriaCommentCount:      "comment_count",
	BadgeCriteriaThreadCount:       "thread_count",
	BadgeCriteriaReactionCount:     "reaction_count",
	BadgeCriteriaSolutionPostCount: "solution_post_count",
	BadgeCriteriaUpVoteCount:       "up_vote_count",
	BadgeCriteriaDownVoteCount:     "down_vote_count",
}

func (b BadgeCriteriaType) String() string {
	return badgeCriteriaTypeToString[b]
}

func BadgeCriteriaTypeFromString(s string) (BadgeCriteriaType, bool) {
	for k, v := range badgeCriteriaTypeToString {
		if v == s {
			return k, true
		}
	}
	return BadgeCriteriaPostCount, false // default to BadgeCriteriaPostCount if not found
}

func BadgeCriteriaTypeFromInt(i int) (BadgeCriteriaType, bool) {
	for k := range badgeCriteriaTypeToString {
		if int(k) == i {
			return k, true
		}
	}
	return BadgeCriteriaPostCount, false // default to BadgeCriteriaPostCount if not found
}
