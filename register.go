package main

import (
	"bitbucket.org/evolutek/gocellaserv-protobuf"
	"net"
)

// Add service to service map
func handleRegister(conn net.Conn, msg *cellaserv.Register) error {
	name := msg.GetName()
	ident := msg.GetIdentification()
	service := &Service{conn, name, ident}
	log.Info("[Register] New service: %+v", service)

	if _, ok := services[name]; !ok {
		services[name] = make(map[string]*Service)
	}

	// Check duplicate
	if s, ok := services[name][ident]; ok {
		log.Error("[Register] Replacing service %+v", s)
		sc := servicesConn[s.conn]
		for i, ss := range sc {
			if ss.name == name && ss.identification == ident {
				// Remove from slice
				sc[i] = sc[len(sc)-1]
				servicesConn[s.conn] = sc[:len(sc)-1]
			}
		}
	}
	services[name][ident] = service

	// Keep track of origin connexion in order to remove when the connexion is closed
	servicesConn[conn] = append(servicesConn[conn], service)

	// No error
	return nil
}

// vim: set nowrap tw=100 noet sw=8: