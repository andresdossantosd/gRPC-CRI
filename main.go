package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"gRPC-CRI/conn_tools"
	crio_proto "gRPC-CRI/crio/proto/parser"
	reflection_crio "gRPC-CRI/crio/proto/reflection"
	"log"
	"strconv"
	"strings"
	"time"
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
	// Root context
	ctx := context.Background()
	var protos ProtosDir

	// all defined flags are technically optional. There is no mandatory concept
	grpc_method := flag.String("method", "", "CRI-O gRPC method to request information")
	flag.Var(&protos, "import-proto", "Proto's file comma separated list")
	timeout := flag.Int("timeout", 15, "gRPC connection timeout")
	verbose_ptr := flag.Bool("verbose", false, "verbose mode")
	flag.Parse()

	// After processing flags, only remains non-flagged arguments, programm name is not included also on flag.Args() --> 2025/04/18 03:47:33 [unix:///var/lib/crio/crio.sock]
	if flag.NArg() != 1 {
		// if no host, fatal !
		log.Fatal("Host is missing")
	}

	// check if host is reachable
	host := flag.Arg(0)
	test_conn := &conn_tools.ConnTools{
		Host:    host,
		Timeout: time.Duration(*timeout) * time.Second,
	}

	if !test_conn.ConnHost(ctx) {
		log.Fatal("Cannot reach host" + host)
	}

	if *verbose_ptr {
		// convert interface{} to string also %v	the value in a default format
		log.Printf("service : " + fmt.Sprintf("%v", _c(grpc_method == nil)._d("", *grpc_method)))
		log.Printf("import-proto : " + fmt.Sprintf("%s", protos))
		log.Printf("timeout : " + strconv.FormatInt(int64(*timeout), 10))
	}

	// if no import-proto, reflections is going to be used to obtained and discover the service
	if *verbose_ptr && len(protos) == 0 {
		log.Printf("Missed import-prot, gRPC reflection it is going to be used")
	}

	if grpc_method == nil || *grpc_method == "" {
		log.Printf("gRPC service is missing, just listing methods")
	}

	parser := &crio_proto.ParserProto{
		Verbose: *verbose_ptr,
		Reflection: &reflection_crio.Reflection{
			Host:     host,
			Required: len(protos) == 0,
			Timeout:  time.Duration(*timeout) * time.Second,
		},
	}

	grpcServices, err := parser.ManageProtoFile(ctx, grpc_method, protos)
	if err != nil {
		log.Fatal("Cannot list services due to: " + err.Error())
	}
	grpcServicesJson, _ := json.MarshalIndent(grpcServices, "", "  ")
	fmt.Println(string(grpcServicesJson))
}
