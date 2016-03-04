package main

import (
	"fmt"
	"log"

	"github.com/rgbkrk/libvirt-go"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// ==== SERVER:
// /etc/default/libvirtd: Add --listen or -l parameter
// /etc/libvirt/libvirtd.conf: listen_tls=0, listen_tcp=1, listen_addr="0.0.0.0"
// saslpasswd2 -a libvirt root (-> Enter password rootpw)
// Check user: sudo sasldblistusers2 -f /etc/libvirt/passwd.db
// sudo /etc/init.d/libvirt-bin restart
// ==== CLIENT:
// sudo apt-get install libsasl2-modules
// /etc/libvirt/auth.conf
// [credentials-root]
// authname=root@wally131
// password=rootpw
// [auth-libvirt-wally131]
// credentials=root
// $ virsh -c 'qemu+tcp://wally131/system' list

func main() {
	con, err := libvirt.NewVirConnection("qemu+tcp://root@wally131.cit.tu-berlin.de/system")
	check(err)
	fmt.Println(con.GetHostname())
	fmt.Println(con.GetLibVersion())
	fmt.Println(con.GetType())
	fmt.Println(con.GetCapabilities())
	fmt.Println(con.GetNodeInfo())
}
