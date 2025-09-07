package events

const (
	//  system events
	UserJoined = "user.joined"
	UserLeft   = "user.left"
	// user events
	MediaControlUpdated = "media.control.updated"
	SyncMedia           = "sync.media"
	ReactionSent        = "reaction.sent"
	RaisedHand          = "raised_hand.sent"
	ShareScreenStarted  = "share_screen.started"
	ShareScreenEnded    = "share_screen.ended"
)
