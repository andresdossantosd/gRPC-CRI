package reflection_crio

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const gRPC_METHOD = "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"

type Reflection struct {
	Host     string
	Timeout  time.Duration
	Required bool
}

func listServicesv1Alpha(stream_grpc *grpc.ClientStream) (msg string, err error) {

	req := &grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{
			ListServices: "*",
		},
	}

	err = (*stream_grpc).SendMsg(req)
	if err != nil {
		log.Printf("Failed writing on gRPC stream : " + err.Error())
		return "", err
	}

	m := &grpc_reflection_v1alpha.ServerReflectionResponse{}
	err = (*stream_grpc).RecvMsg(m)
	if err != nil {
		log.Printf("Failed reading on gRPC stream : " + err.Error())
		return "", err
	}
	log.Printf("Stream " + m.String())
	return m.String(), err
}

func (r *Reflection) GetProtos(ctx context.Context) (msg string, err error) {
	// Get ClientConn Object
	insecure_creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(r.Host, grpc.WithTransportCredentials(insecure_creds))
	if err != nil {
		log.Printf("Failed connection : " + err.Error())
		return
	}

	// Stream descriptor
	desc := &grpc.StreamDesc{
		StreamName:    "reflection",
		ServerStreams: true,
		ClientStreams: true,
	}
	// Bidirectional streaming RPC for reflection service (stream key word on params and response on proto file)
	stream_grpc, err := conn.NewStream(ctx, desc, gRPC_METHOD)
	if err != nil {
		log.Printf("Failed gRPC stream : " + err.Error())
		return
	}

	msg, err = listServicesv1Alpha(&stream_grpc)
	return msg, err
}
