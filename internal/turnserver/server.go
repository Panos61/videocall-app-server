package turnserver

import (
	"flag"
	"log"
	"net"
	"strconv"

	"github.com/pion/turn/v4"
)

func StartTurnServer(publicIP string, port int, realm string) {
	// publicIP := flag.String("public-ip", "", "IP Address that TURN can be contacted by.")
	// port := flag.Int("port", 3478, "TURN server listening port.")
	// users := flag.String("users", "", "List of user==pass")
	// realm := flag.String("realm", "pion.ly", "Realm (defaults to \"pion.ly\")")

	flag.Parse()

	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		log.Panicf("Failed to create TURN server listener: %s", err)
	}

	server, err := turn.NewServer(turn.ServerConfig{
		Realm: realm,
		AuthHandler: func(username, realm string, srcAddr net.Addr) (key []byte, ok bool) {
			// todo use .env
			if username == "panos" {
				return turn.GenerateAuthKey(username, realm, "1234"), true
			}

			return nil, false
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(publicIP),
					Address:      "0.0.0.0",
				},
			},
		},
	})

	if err != nil {
		log.Panic(err)
	}

	if err = server.Close(); err != nil {
		log.Panic(err)
	}
}
