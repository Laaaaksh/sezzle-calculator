package boot

import "net"

// UseListen replaces how Run binds its address, so tests can serve over an
// in-memory listener. It is in a _test file so production code cannot swap it.
func (a *App) UseListen(listen func(network, address string) (net.Listener, error)) {
	a.listen = listen
}
