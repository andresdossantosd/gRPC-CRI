package crio_proto

import (
	"log"
	"os"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// Becareful, fields need it on others packages should start with Uppercase !!
// If Reflection type was reflection, it could not be exported to other packages
type Reflection struct {
	Host     string
	Required bool
}

// Becareful, fields need it on others packages should start with Uppercase !!
type ParserProto struct {
	Verbose bool
	// embebido: Go promueve automáticamente los campos de Reflection a ParserProto.
	// Es decir, podés acceder a parser.URL o parser.Need directamente sin escribir parser.Reflection.URL.
	*Reflection
}

// ellipsis
func (p *ParserProto) ManagedProtoFile(paths ...string) (err error) {

	if p.Required {
		if p.Verbose {
			log.Printf("Starting reflection gRPC request")
		}
		// TODO: implementar reflection
	}

	for _, path := range paths {
		p.ParseProtoFile(path)
	}
	return nil
}

// --> := is for new assigments, instead = is for previously declared vars on context, such as proto and err
// --> becareful, cannot collision packages names of imported packages with local packages, previously this local
//
//	package was called proto and github.com/yoheimuta/go-protoparser/v4/parser has also a proto package that we need to rference it.
func (p *ParserProto) ParseProtoFile(path string) (parse_file *parser.Proto, err error) {
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
