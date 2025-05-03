package crio_proto

import (
	"context"
	"gRPC-CRI/crio/proto/elem_crio"
	reflection_crio "gRPC-CRI/crio/proto/reflection"
	"log"
	"os"
	"strconv"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

type ParserProto struct {
	Verbose bool
	*reflection_crio.Reflection
}

func (p *ParserProto) ManageProtoFile(ctx context.Context, invoke_method *string, paths []string) (grpcServices []*elem_crio.Service, err error) {
	localCtx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	if p.Required {
		if p.Verbose {
			log.Printf("Starting reflection gRPC request")
		}
		grpcServices, err = p.Reflection.GetProtosServices(localCtx, invoke_method)
	} else {
		q, err := localProtosFiles(paths...)
		if err != nil {
			log.Printf("Starting reflection gRPC request")
		} else {
			grpcServices = parseToUser(q)
		}
	}

	return
}

// ellipsis and not accesible from other packages
func localProtosFiles(paths ...string) (q []parser.Proto, err error) {
	q = make([]parser.Proto, 0)
	for _, path := range paths {
		parser, err := parseProtoFile(path)
		if err != nil {
			break
		}
		q = append(q, *parser)
	}

	return q, err
}

// --> := is for new assigments, instead = is for previously declared vars on context, such as proto and err
// --> becareful, cannot collision packages names of imported packages with local packages, previously this local
// non accesible form other packages
//
//	package was called proto and github.com/yoheimuta/go-protoparser/v4/parser has also a proto package that we need to rference it.
func parseProtoFile(path string) (parse_file *parser.Proto, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	parse_file, err = protoparser.Parse(file)
	if err != nil {
		return nil, err
	}
	return parse_file, nil
}

// Map insertion with order
type MapChronos struct {
	Umap  map[string][]*elem_crio.Field
	Items []*string
}

// ALERT !!! Visitor design pattern. We need to implement Visitor's methods for Visitee object that []ProtoBody ( from parser package ) have.
type ServiceVisitors struct {
	OwnServices  []*elem_crio.Service
	MessageTypes *MapChronos
}

func (v *ServiceVisitors) VisitService(s *parser.Service) bool {
	v.OwnServices = append(v.OwnServices, &elem_crio.Service{
		Name:    s.ServiceName,
		Methods: []*elem_crio.Method{},
	})
	for _, methods := range s.ServiceBody {
		methods.Accept(v)
	}
	return true
}

func (v *ServiceVisitors) VisitRPC(r *parser.RPC) bool {
	v.OwnServices[len(v.OwnServices)-1].Methods =
		append(v.OwnServices[len(v.OwnServices)-1].Methods, &elem_crio.Method{
			Name: r.RPCName,
			MethodSpec: &elem_crio.MethodInfo{
				Input:   nil,
				Output:  nil,
				NameIn:  r.RPCRequest.MessageType,
				NameOut: r.RPCResponse.MessageType,
			},
			ClientStream: r.RPCRequest.IsStream,
			ServerStream: r.RPCRequest.IsStream,
		})
	return true
}

func (v *ServiceVisitors) VisitMessage(m *parser.Message) bool {
	if v.MessageTypes == nil {
		v.MessageTypes = &MapChronos{
			Umap:  make(map[string][]*elem_crio.Field),
			Items: []*string{},
		}
	}
	for _, msg := range m.MessageBody {
		v.MessageTypes.Umap[m.MessageName] = []*elem_crio.Field{}
		v.MessageTypes.Items = append(v.MessageTypes.Items, &m.MessageName)
		msg.Accept(v)
	}
	return true
}

// condition
type _c bool

// interface{} : An empty interface may hold values of any type
func (condicion _c) _d(value1, value2 interface{}) interface{} {
	if condicion {
		return value1
	}
	return value2
}

func (v *ServiceVisitors) VisitField(f *parser.Field) bool {
	// mantain order on insert ops. on map !
	lastItem := v.MessageTypes.Items[len(v.MessageTypes.Items)-1]
	ptrField := v.MessageTypes.Umap[*lastItem]
	fieldNumber, _ := strconv.ParseInt(f.FieldNumber, 10, 32)
	fieldLabel := _c(f.IsOptional)._d("", _c(f.IsRepeated)._d("", ""))
	v.MessageTypes.Umap[*lastItem] = append(ptrField, &elem_crio.Field{
		Label: fieldLabel.(string),
		Type:  f.Type,
		Name:  f.FieldName,
		// transform int32
		Number: int32(fieldNumber),
	})
	return true
}

/*
This error:

	``cannot use svcVisitor (variable of type *ServiceVisitors) as parser.Visitor value
	in argument to protoFileElem.Accept: *ServiceVisitors does not implement parser.
	Visitor (missing method VisitComment)``

Is due to non implementing Visitor interface on visitors object struct . ServiceVisitors in our case !! It is need it for visitor design pattern.

We ServiceVisitor needs to implement Visitor interface, so all this methods should be included :

		type Visitor interface {
		VisitComment(*Comment)
		VisitDeclaration(*Declaration) (next bool)
		VisitEdition(*Edition) (next bool)
		VisitEmptyStatement(*EmptyStatement) (next bool)
		VisitEnum(*Enum) (next bool)
		VisitEnumField(*EnumField) (next bool)
		VisitExtend(*Extend) (next bool)
		VisitExtensions(*Extensions) (next bool)
		VisitField(*Field) (next bool)
		VisitGroupField(*GroupField) (next bool)
		VisitImport(*Import) (next bool)
		VisitMapField(*MapField) (next bool)
		VisitMessage(*Message) (next bool)
		VisitOneof(*Oneof) (next bool)
		VisitOneofField(*OneofField) (next bool)
		VisitOption(*Option) (next bool)
		VisitPackage(*Package) (next bool)
		VisitReserved(*Reserved) (next bool)
		VisitRPC(*RPC) (next bool)
		VisitService(*Service) (next bool)
		VisitSyntax(*Syntax) (next bool)
	}
*/
func (v *ServiceVisitors) VisitComment(c *parser.Comment)                    {}
func (v *ServiceVisitors) VisitSyntax(s *parser.Syntax) bool                 { return true }
func (v *ServiceVisitors) VisitEmptyStatement(e *parser.EmptyStatement) bool { return true }
func (v *ServiceVisitors) VisitEnum(e *parser.Enum) bool                     { return true }
func (v *ServiceVisitors) VisitEnumField(e *parser.EnumField) bool           { return true }
func (v *ServiceVisitors) VisitExtend(e *parser.Extend) bool                 { return true }
func (v *ServiceVisitors) VisitEdition(e *parser.Edition) bool               { return true }
func (v *ServiceVisitors) VisitExtensions(e *parser.Extensions) bool         { return true }
func (v *ServiceVisitors) VisitGroupField(g *parser.GroupField) bool         { return true }
func (v *ServiceVisitors) VisitImport(i *parser.Import) bool                 { return true }
func (v *ServiceVisitors) VisitMapField(f *parser.MapField) bool             { return true }
func (v *ServiceVisitors) VisitOneof(o *parser.Oneof) bool                   { return true }
func (v *ServiceVisitors) VisitOneofField(f *parser.OneofField) bool         { return true }
func (v *ServiceVisitors) VisitOption(o *parser.Option) bool                 { return true }
func (v *ServiceVisitors) VisitPackage(p *parser.Package) bool               { return true }
func (v *ServiceVisitors) VisitReserved(r *parser.Reserved) bool             { return true }
func (v *ServiceVisitors) VisitDeclaration(r *parser.Declaration) bool       { return true }

func parseToUser(protoFiles []parser.Proto) []*elem_crio.Service {
	svcVisitor := &ServiceVisitors{}
	for _, protoFile := range protoFiles {
		for _, protoFileElem := range protoFile.ProtoBody {
			protoFileElem.Accept(svcVisitor)
		}
	}

	mapMsgs := svcVisitor.MessageTypes.Umap
	// search message types for fields types
	for _, ownSvc := range svcVisitor.OwnServices {
		for _, ownMethod := range ownSvc.Methods {
			_, inOk := mapMsgs[ownMethod.MethodSpec.NameIn]
			if inOk {
				ownMethod.MethodSpec.Input = mapMsgs[ownMethod.MethodSpec.NameIn]
			}
			_, outOk := mapMsgs[ownMethod.MethodSpec.NameOut]
			if outOk {
				ownMethod.MethodSpec.Output = mapMsgs[ownMethod.MethodSpec.NameOut]
			}
		}
	}

	return svcVisitor.OwnServices
}
