package events

const (
	//  system events
	UserJoined  = "user.joined"  // not used
	UserLeft    = "user.left"    // not used
	HostLeft    = "host.left"    // when initial host leaves the room
	HostUpdated = "host.updated" // when the new host is assigneds
	RoomKilled  = "room.killed"  // when the room is killed

	// user events
	MediaStateUpdated  = "media.state.updated"
	SyncMedia          = "media.synced"
	ReactionSent       = "reaction.sent"
	RaisedHand         = "raisedhand.sent"
	ShareScreenStarted = "sharescreen.started"
	ShareScreenEnded   = "sharescreen.ended"
)
