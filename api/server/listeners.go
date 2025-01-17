package server

import "github.com/lean-mu/mu/fnext"

// AddCallListener adds a listener that will be fired before and after a function is executed.
func (s *Server) AddCallListener(listener fnext.CallListener) {
	s.agent.AddCallListener(listener)
}
