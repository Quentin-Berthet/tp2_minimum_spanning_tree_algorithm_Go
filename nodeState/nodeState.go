package nodeState

type NodeState string

const (
	Sleeping NodeState = "Sleeping"
	Find     NodeState = "Find"
	Found    NodeState = "Found"
)
