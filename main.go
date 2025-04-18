package main

import (
	"flag"
	"fmt"
	crio_proto "gRPC-CRI/crio/proto"
	"log"
	"strconv"
	"strings"
)

// condition
type _c bool

// interface{} : An empty interface may hold values of any type
func (condicion _c) _d(value1, value2 interface{}) interface{} {
	if condicion {
		return value1
	}
	return value2
}

// ProtosDir implements Values type methods Set() and String()
type ProtosDir []string

// ProtosDir type methods
func (p *ProtosDir) String() string {
	return fmt.Sprint(*p)
}

func (p *ProtosDir) Set(value string) (err error) {
	for _, item := range strings.Split(value, ",") {
		*p = append(*p, item)
	}
	return nil
}

// Methods started with low case, they are not accesible aoutside package

func main() {
	var protos ProtosDir
	grpc_service := flag.String("service", "", "CRI-O gRPC service to request information")
	flag.Var(&protos, "import-proto", "Proto's file comma separated list")
	timeout := flag.Int("timeout", 15, "gRPC connection timeout")
	verbose_ptr := flag.Bool("verbose", false, "verbose mode")
	flag.Parse()

	// TODO: Host:Port argument, must be the unique non-parse flaged and is used without it
	log.Printf("" + fmt.Sprintf("%s", flag.Args()))

	// After processing flags, only remains non-flagged arguments, programm name is not included also on flag.Args() --> 2025/04/18 03:47:33 [unix:///var/lib/crio/crio.sock]
	if flag.NArg() != 1 {
		log.Fatal("Host is missing")
	}

	host := flag.Arg(0)

	// TODO: specified if unix socket as flag or implement parsing all address (add rule that if user want unix socket, Host must be unix:// ...)
	if *verbose_ptr {
		// convert interface{} to string also %v	the value in a default format
		log.Printf("service : " + fmt.Sprintf("%v", _c(grpc_service == nil)._d("", *grpc_service)))
		log.Printf("import-proto : " + fmt.Sprintf("%s", protos))
		log.Printf("timeout : " + strconv.FormatInt(int64(*timeout), 10))
	}

	// if no import-proto, reflections is going to be used to obtained and discover the service
	if *verbose_ptr && len(protos) == 0 {
		log.Printf("Missed import-prot, gRPC reflection it is going to be used")
	}

	if grpc_service == nil || *grpc_service == "" {
		log.Fatal("gRPC service is missing")
	}

	//
	parser := &crio_proto.ParserProto{Verbose: *verbose_ptr, Reflection: &crio_proto.Reflection{Host: host, Required: len(protos) == 0}}
	parser.ManagedProtoFile(protos...)

}
