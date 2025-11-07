package events

const (
	//  system events
	UserJoined = "user.joined"
	UserLeft   = "user.left"

	HostLeft     = "host.left"     // when initial host leaves the room
	HostHandover = "host.handover" // when the randomly selected candidate rejects the host promotion, the event is triggered again excluding the rejected candidate
	HostUpdated  = "host.updated"  // when the new host finally accepts the host promotion

	// user events
	MediaControlUpdated = "media.control.updated"
	SyncMedia           = "sync.media"
	ReactionSent        = "reaction.sent"
	RaisedHand          = "raised_hand.sent"
	ShareScreenStarted  = "share_screen.started"
	ShareScreenEnded    = "share_screen.ended"
)
