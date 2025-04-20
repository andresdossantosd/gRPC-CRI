package reflection_crio

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const gRPC_METHOD = "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"

type Reflection struct {
	Host     string
	Timeout  time.Duration
	Required bool
}

// private func
func listServicesMethodsv1Alpha(stream_grpc *grpc.ClientStream, service string) (err error) {

	req := &grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: service,
		},
	}
	err = (*stream_grpc).SendMsg(req)
	if err != nil {
		log.Printf("Failed writing on gRPC stream : " + err.Error())
		return err
	}
	m := &grpc_reflection_v1alpha.ServerReflectionResponse{}
	err = (*stream_grpc).RecvMsg(m)
	if err != nil {
		log.Printf("Failed reading on gRPC stream : " + err.Error())
		return err
	}
	// get all dependencies (imports) protos file
	protos_files_bytes := m.GetFileDescriptorResponse().GetFileDescriptorProto()

	// iterarte over all protobuf files
	for _, proto_file_bytes := range protos_files_bytes {

		// non-nil pointer message
		var proto_file descriptorpb.FileDescriptorProto

		// parse from bytes to protobuf (ProtoMessage())
		err = proto.Unmarshal(proto_file_bytes, &proto_file)
		if err != nil {
			log.Printf("Failed reading on gRPC stream : " + err.Error())
			return err
		}
		for _, svc := range proto_file.GetService() {
			for _, method := range svc.GetMethod() {
				log.Printf("  RPC Method: %s", method.GetName())
				log.Printf("    Input: %s", method.GetInputType())
				log.Printf("    Output: %s", method.GetOutputType())
				log.Printf("    bidi-stream = %t", method.GetClientStreaming() && method.GetServerStreaming())
			}
		}

	}
	return nil

}

// private func
func listServicesv1Alpha(stream_grpc *grpc.ClientStream) (services []*grpc_reflection_v1alpha.ServiceResponse, err error) {

	// v1alpha deprecated, but Go/Golang does not have an v1 implementation yet for server reflection request
	req := &grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{
			ListServices: "7",
		},
	}

	err = (*stream_grpc).SendMsg(req)
	if err != nil {
		log.Printf("Failed writing on gRPC stream : " + err.Error())
		return nil, err
	}

	m := &grpc_reflection_v1alpha.ServerReflectionResponse{}
	err = (*stream_grpc).RecvMsg(m)
	if err != nil {
		log.Printf("Failed reading on gRPC stream : " + err.Error())
		return nil, err
	}
	services = m.GetListServicesResponse().Service
	return services, err
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

	// TODO: format response based on user service it wants to execute
	// TODO: create logging package to print logs if verbose active
	// TODO: logging with verbose
	services, err := listServicesv1Alpha(&stream_grpc)
	for _, service := range services {
		listServicesMethodsv1Alpha(&stream_grpc, service.Name)
	}
	return "", err
}
