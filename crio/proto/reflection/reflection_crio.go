package reflection_crio

import (
	"context"
	"gRPC-CRI/crio/proto/elem_crio"
	"log"
	"strings"
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

func searchMethodMsgTypev1Alpha(proto_msg_types []*descriptorpb.DescriptorProto, msg_input string, msg_output string) (method_ipts_outs *elem_crio.MethodInfo) {

	for _, msg_types := range proto_msg_types {
		if msg_types.Name == nil {
			continue
		}

		// Why HasSuffix ? msg_input and msg_output are obtained by calling method.GetInputType(), this will return input/output message type including package.<message> (package.message_name)
		if strings.HasSuffix(msg_input, *msg_types.Name) || strings.HasSuffix(msg_output, *msg_types.Name) {
			if method_ipts_outs == nil {
				method_ipts_outs = &elem_crio.MethodInfo{
					Input:   []*elem_crio.Field{},
					Output:  []*elem_crio.Field{},
					NameIn:  "",
					NameOut: "",
				}
			}
			if "."+*(msg_types.Name) == msg_input {
				// protobuf3, each message field has name, type and label (optional, required or repeated)
				for _, fields := range msg_types.Field {
					method_ipts_outs.Input = append(method_ipts_outs.Input, &elem_crio.Field{
						Name:   *fields.Name,
						Type:   strings.ToLower(strings.Replace(fields.Type.String(), "TYPE_", "", 1)),
						Label:  strings.ToLower(strings.Replace(fields.Label.String(), "LABEL_", "", 1)),
						Number: *fields.Number,
					})
				}
				method_ipts_outs.NameIn = *msg_types.Name
			}
			if "."+*(msg_types.Name) == msg_output {
				for _, fields := range msg_types.Field {
					method_ipts_outs.Output = append(method_ipts_outs.Output, &elem_crio.Field{
						Name:   *fields.Name,
						Type:   strings.ToLower(strings.Replace(fields.Type.String(), "TYPE_", "", 1)),
						Label:  strings.ToLower(strings.Replace(fields.Label.String(), "LABEL_", "", 1)),
						Number: *fields.Number,
					})
				}
				method_ipts_outs.NameOut = *msg_types.Name
			}
		}
	}
	return
}

// return how to call method
func listServicesMethodsv1Alpha(stream_grpc *grpc.ClientStream, service string, invoke_method *string) (method_obj []*elem_crio.Method, err error) {

	req := &grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: service,
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
			return nil, err
		}
		for _, svc := range proto_file.GetService() {
			for _, method := range svc.GetMethod() {
				if *invoke_method != "" && *method.Name != *invoke_method {
					continue
				}
				// get inputs/outputs message types and their structure
				method_spec := searchMethodMsgTypev1Alpha(proto_file.GetMessageType(), method.GetInputType(), method.GetOutputType())
				method_obj = append(method_obj, &elem_crio.Method{
					Name:         *method.Name,
					MethodSpec:   method_spec,
					ClientStream: method.GetClientStreaming(),
					ServerStream: method.GetServerStreaming(),
				})
			}
		}
	}
	return

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

func (r *Reflection) GetProtosServices(ctx context.Context, invoke_method *string) (grpc_services []*elem_crio.Service, err error) {

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
	// TODO: obtener de MethodInfo un json o proto de los Input e Output
	services, err := listServicesv1Alpha(&stream_grpc)
	var methods []*elem_crio.Method
	for _, service := range services {

		methods, err = listServicesMethodsv1Alpha(&stream_grpc, service.Name, invoke_method)
		if err != nil || methods == nil {
			continue
		} else {
			grpc_services = append(grpc_services, &elem_crio.Service{
				Methods: methods,
				Name:    service.Name,
			})
		}

	}
	return
}
