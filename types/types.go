package types

type ConnectionState string

const (
	ConnectionStateIdle       ConnectionState = "idle"
	ConnectionStateConnecting ConnectionState = "connecting"
	ConnectionStateConnected  ConnectionState = "connected"
	ConnectionStateBackoff    ConnectionState = "backoff"
	ConnectionStateError      ConnectionState = "error"
	ConnectionStateStopped    ConnectionState = "stopped"
)
