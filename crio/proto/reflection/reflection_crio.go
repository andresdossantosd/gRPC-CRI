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

type Field struct {
	Label string
	Type  string
	Name  string
}

type Method_Param struct {
	Input  []*Field
	Output []*Field
	Name   string
}

type Method struct {
	InputParameter  *Method_Param
	OutputParameter *Method_Param
	ClientStream    bool
	ServerStream    bool
}

func searchMethodMsgTypev1Alpha(proto_msg_types []*descriptorpb.DescriptorProto, msg_input string, msg_output string) (grpc_param *Method_Param) {
	// TODO: Obtener todos los fields de un message, su tipo y ponerlo en un json
	for _, msg_types := range proto_msg_types {
		if msg_types.Name == nil {
			continue
		}
		if *(msg_types.Name) == msg_input || *(msg_types.Name) == msg_output {
			if grpc_param == nil {
				grpc_param = &Method_Param{
					Input:  []*Field{},
					Output: []*Field{},
					Name:   "",
				}
			}
			if *(msg_types.Name) == msg_input {
				// TODO: Terminar de parsear fields, que ahora sean los fields normales
				/*

				 */
				for _, fields := range msg_types.Field {
					grpc_param.Input = &Field{
						Name:  *fields.Name,
						Type:  fields.Type.String(),
						Label: fields.Label.String(),
					}
				}
				grpc_param.Name = msg_input
			}
			if *(msg_types.Name) == msg_output {
				for _, fields := range msg_types.Field {
					grpc_param.Output = &Field{
						Name:  *fields.Name,
						Type:  fields.Type.String(),
						Label: fields.Label.String(),
					}
				}
				grpc_param.Name = msg_output
			}
		}
	}
	return
}

// return how to call method
func listServicesMethodsv1Alpha(stream_grpc *grpc.ClientStream, service string, invoke_method string) (method_obj *Method, err error) {

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
				if method.GetName() == invoke_method {
					method_params := searchMethodMsgTypev1Alpha(proto_file.GetMessageType(), method.GetInputType(), method.GetOutputType())
					return &Method{
						Parameters:    method_params,
						Client_stream: method.GetClientStreaming(),
						Server_stream: method.GetServerStreaming(),
					}, nil
				}
			}
		}
	}
	return method_obj, nil

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

func (r *Reflection) GetProtos(ctx context.Context, invoke_method string) (method_obj *Method, err error) {
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
	// TODO: obtener de Method_Param un json o proto de los Input e Output
	services, err := listServicesv1Alpha(&stream_grpc)
	for _, service := range services {
		method_obj, err = listServicesMethodsv1Alpha(&stream_grpc, service.Name, invoke_method)
		if err != nil || method_obj == nil {
			continue
		} else {
			return method_obj, nil
		}

	}
	return nil, err
}
