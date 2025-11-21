package events

const (
	//  system events
	UserJoined  = "user.joined"  // not used
	UserLeft    = "user.left"    // not used
	HostLeft    = "host.left"    // when initial host leaves the room
	HostUpdated = "host.updated" // when the new host is assigneds

	// user events
	MediaControlUpdated = "media.control.updated"
	SyncMedia           = "sync.media"
	ReactionSent        = "reaction.sent"
	RaisedHand          = "raised_hand.sent"
	ShareScreenStarted  = "share_screen.started"
	ShareScreenEnded    = "share_screen.ended"
)
