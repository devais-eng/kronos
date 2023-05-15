package sync

import "devais.it/kronos/internal/pkg/sync/messages"

type ConnectionCallback func()
type DisconnectionCallback func(err error)
type SyncCallback func(message messages.Sync)
type CommandCallback func(message *messages.ServerCommand)

// Client defines methods that a synchronization client
// must implement through its own protocol.
type Client interface {
	Connect() error
	Disconnect() error
	SetConnectionCallback(cb ConnectionCallback)
	SetDisconnectionCallback(cb DisconnectionCallback)
	SetSyncCallback(cb SyncCallback)
	SetCommandCallback(cb CommandCallback)
	Subscribe() error
	PublishVersions() error
	PublishEvents(events []messages.Event) error
	PublishCommandResponse(message *messages.CommandResponse) error
}
