package elem_crio

type Field struct {
	Label  string `json:"label"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Number int32  `json:"number"`
}

type MethodInfo struct {
	Input   []*Field `json:"input_fields"`
	Output  []*Field `json:"output_fields"`
	NameIn  string   `json:"input_name"`
	NameOut string   `json:"output_name"`
}

type Method struct {
	Name         string      `json:"method"`
	MethodSpec   *MethodInfo `json:"method_spec"`
	ClientStream bool        `json:"request_stream"`
	ServerStream bool        `json:"response_stream"`
}

type Service struct {
	Methods []*Method `json:"methods"`
	Name    string    `json:"service"`
}
