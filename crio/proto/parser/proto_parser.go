package crio_proto

import (
	"context"
	reflection_crio "gRPC-CRI/crio/proto/reflection"
	"log"
	"os"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// Becareful, fields need it on others packages should start with Uppercase !!
type ParserProto struct {
	Verbose bool
	// embebido: Go promueve automáticamente los campos de Reflection a ParserProto.
	// Es decir, podés acceder a parser.URL o parser.Need directamente sin escribir parser.Reflection.URL.
	*reflection_crio.Reflection
}

func (p *ParserProto) ManageProtoFile(ctx context.Context, paths []string) (q []parser.Proto, err error) {
	localCtx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	if p.Required {
		if p.Verbose {
			log.Printf("Starting reflection gRPC request")
		}
		p.Reflection.GetProtos(localCtx)
	} else {
		q, err = localProtosFiles(paths...)
	}

	return q, err
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
