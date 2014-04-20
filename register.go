package main

import (
	"bitbucket.org/evolutek/cellaserv2-protobuf"
	"encoding/json"
	"net"
)

var logNewService = "log.new-service"

// Add service to service map
func handleRegister(conn net.Conn, msg *cellaserv.Register) {
	name := msg.GetName()
	ident := msg.GetIdentification()
	service := newService(conn, name, ident)
	log.Info("[Services] New %s", service)

	if _, ok := services[name]; !ok {
		services[name] = make(map[string]*Service)
	}

	// Check duplicate
	if s, ok := services[name][ident]; ok {
		log.Warning("[Services] Replace %s", s)
		sc := servicesConn[s.Conn]
		for i, ss := range sc {
			if ss.Name == name && ss.Identification == ident {
				// Remove from slice
				sc[i] = sc[len(sc)-1]
				servicesConn[s.Conn] = sc[:len(sc)-1]
			}
		}
	} else {
		pub, _ := json.Marshal(service.JSONStruct())
		cellaservPublish(logNewService, pub)
	}
	services[name][ident] = service

	// Keep track of origin connection in order to remove when the connection is closed
	servicesConn[conn] = append(servicesConn[conn], service)
}

// vim: set nowrap tw=100 noet sw=8:
