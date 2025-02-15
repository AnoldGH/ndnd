package codegen

import "fmt"

// StringField represents a UTF-8 encoded string.
type StringField struct {
	BaseTlvField

	opt bool
}

func NewStringField(name string, typeNum uint64, annotation string, _ *TlvModel) (TlvField, error) {
	return &StringField{
		BaseTlvField: BaseTlvField{
			name:    name,
			typeNum: typeNum,
		},
		opt: annotation == "optional",
	}, nil
}

func (f *StringField) GenEncodingLength() (string, error) {
	g := strErrBuf{}
	if f.opt {
		g.printlnf("if optval, ok := value.%s.Get(); ok {", f.name)
		g.printlne(GenTypeNumLen(f.typeNum))
		g.printlne(GenNaturalNumberLen("len(optval)", true))
		g.printlnf("l += uint(len(optval))")
		g.printlnf("}")
	} else {
		g.printlne(GenTypeNumLen(f.typeNum))
		g.printlne(GenNaturalNumberLen("len(value."+f.name+")", true))
		g.printlnf("l += uint(len(value.%s))", f.name)
	}
	return g.output()
}

func (f *StringField) GenEncodingWirePlan() (string, error) {
	return f.GenEncodingLength()
}

func (f *StringField) GenEncodeInto() (string, error) {
	g := strErrBuf{}
	if f.opt {
		g.printlnf("if optval, ok := value.%s.Get(); ok {", f.name)
		g.printlne(GenEncodeTypeNum(f.typeNum))
		g.printlne(GenNaturalNumberEncode("len(optval)", true))
		g.printlnf("copy(buf[pos:], optval)")
		g.printlnf("pos += uint(len(optval))")
		g.printlnf("}")
	} else {
		g.printlne(GenEncodeTypeNum(f.typeNum))
		g.printlne(GenNaturalNumberEncode("len(value."+f.name+")", true))
		g.printlnf("copy(buf[pos:], value.%s)", f.name)
		g.printlnf("pos += uint(len(value.%s))", f.name)
	}
	return g.output()
}

func (f *StringField) GenReadFrom() (string, error) {
	g := strErrBuf{}
	g.printlnf("{")
	g.printlnf("var builder strings.Builder")
	g.printlnf("_, err = reader.CopyN(&builder, int(l))")
	g.printlnf("if err == nil {")
	if f.opt {
		g.printlnf("value.%s.Set(builder.String())", f.name)
	} else {
		g.printlnf("value.%s = builder.String()", f.name)
	}
	g.printlnf("}")
	g.printlnf("}")
	return g.output()
}

func (f *StringField) GenSkipProcess() (string, error) {
	if f.opt {
		return fmt.Sprintf("value.%s.Unset()", f.name), nil
	} else {
		return fmt.Sprintf("err = enc.ErrSkipRequired{Name: \"%s\", TypeNum: %d}", f.name, f.typeNum), nil
	}
}
