// Package ice ...
//
//nolint:dupl
package ice

import "net"

// MultiTCPMuxDefault implements both TCPMux and AllConnsGetter,
// allowing users to pass multiple TCPMux instances to the ICE agent
// configuration.
type MultiTCPMuxDefault struct {
	muxs []TCPMux
}

// NewMultiTCPMuxDefault creates an instance of MultiTCPMuxDefault that
// uses the provided TCPMux instances.
func NewMultiTCPMuxDefault(muxs ...TCPMux) *MultiTCPMuxDefault {
	return &MultiTCPMuxDefault{
		muxs: muxs,
	}
}

// GetConnByUfrag returns a PacketConn given the connection's ufrag and network
// creates the connection if an existing one can't be found. This, unlike
// GetAllConns, will only return a single PacketConn from the first mux that was
// passed in to NewMultiTCPMuxDefault.
func (m *MultiTCPMuxDefault) GetConnByUfrag(ufrag string, isIPv6 bool) (net.PacketConn, error) {
	// NOTE: We always use the first element here in order to maintain the
	// behavior of using an existing connection if one exists.
	if len(m.muxs) == 0 {
		return nil, errNoTCPMuxAvailable
	}
	return m.muxs[0].GetConnByUfrag(ufrag, isIPv6)
}

// RemoveConnByUfrag stops and removes the muxed packet connection
// from all underlying TCPMux instances.
func (m *MultiTCPMuxDefault) RemoveConnByUfrag(ufrag string) {
	for _, mux := range m.muxs {
		mux.RemoveConnByUfrag(ufrag)
	}
}

// GetAllConns returns a PacketConn for each underlying TCPMux
func (m *MultiTCPMuxDefault) GetAllConns(ufrag string, isIPv6 bool) ([]net.PacketConn, error) {
	if len(m.muxs) == 0 {
		// Make sure that we either return at least one connection or an error.
		return nil, errNoTCPMuxAvailable
	}
	var conns []net.PacketConn
	for _, mux := range m.muxs {
		conn, err := mux.GetConnByUfrag(ufrag, isIPv6)
		if err != nil {
			// For now, this implementation is all or none.
			return nil, err
		}
		if conn != nil {
			conns = append(conns, conn)
		}
	}
	return conns, nil
}

// Close the multi mux, no further connections could be created
func (m *MultiTCPMuxDefault) Close() error {
	var err error
	for _, mux := range m.muxs {
		if e := mux.Close(); e != nil {
			err = e
		}
	}
	return err
}
