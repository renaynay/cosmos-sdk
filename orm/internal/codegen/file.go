package codegen

import (
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-proto/generator"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	ormv1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1"
)

type fileGen struct {
	*generator.GeneratedFile
	file *protogen.File
}

func (f fileGen) gen() error {
	f.P("// Code generated by protoc-gen-go-cosmos-orm. DO NOT EDIT.")
	f.P()
	f.P("package ", f.file.GoPackageName)
	stores := make([]*protogen.Message, 0)
	for _, msg := range f.file.Messages {
		tableDesc := proto.GetExtension(msg.Desc.Options(), ormv1.E_Table).(*ormv1.TableDescriptor)
		if tableDesc != nil {
			tableGen, err := newTableGen(f, msg, tableDesc)
			if err != nil {
				return err
			}
			tableGen.gen()
		}
		singletonDesc := proto.GetExtension(msg.Desc.Options(), ormv1.E_Singleton).(*ormv1.SingletonDescriptor)
		if singletonDesc != nil {
			// do some singleton magic
			singletonGen, err := newSingletonGen(f, msg, singletonDesc)
			if err != nil {
				return err
			}
			singletonGen.gen()
		}

		if tableDesc != nil || singletonDesc != nil { // message is one of the tables,
			stores = append(stores, msg)
		}
	}
	f.genStoreInterface(stores)
	f.genStoreStruct(stores)
	f.genStoreMethods(stores)
	f.genStoreInterfaceGuard()
	f.genStoreConstructor(stores)
	return nil
}

func (f fileGen) genStoreInterface(stores []*protogen.Message) {
	f.P("type ", f.storeInterfaceName(), " interface {")
	for _, store := range stores {
		name := f.messageTableInterfaceName(store)
		f.P(name, "()", name)
	}
	f.P()
	f.P("doNotImplement()")
	f.P("}")
	f.P()
}

func (f fileGen) genStoreStruct(stores []*protogen.Message) {
	// struct
	f.P("type ", f.storeStructName(), " struct {")
	for _, message := range stores {
		f.P(f.param(message.GoIdent.GoName), " ", f.messageTableInterfaceName(message))
	}
	f.P("}")
}

func (f fileGen) storeAccessorName() string {
	return f.storeInterfaceName()
}

func (f fileGen) storeInterfaceName() string {
	return strcase.ToCamel(f.fileShortName()) + "Store"
}

func (f fileGen) storeStructName() string {
	return strcase.ToLowerCamel(f.fileShortName()) + "Store"
}

func (f fileGen) fileShortName() string {
	filename := f.file.Proto.GetName()
	shortName := filepath.Base(filename)
	i := strings.Index(shortName, ".")
	if i > 0 {
		return shortName[:i]
	}
	return strcase.ToCamel(shortName)
}

func (f fileGen) messageTableInterfaceName(m *protogen.Message) string {
	return m.GoIdent.GoName + "Table"
}

func (f fileGen) messageReaderInterfaceName(m *protogen.Message) string {
	return m.GoIdent.GoName + "Reader"
}

func (f fileGen) messageTableVar(m *protogen.Message) string {
	return f.param(m.GoIdent.GoName + "Table")
}

func (f fileGen) param(name string) string {
	return strcase.ToLowerCamel(name)
}

func (f fileGen) messageTableReceiverName(m *protogen.Message) string {
	return f.param(f.messageTableInterfaceName(m))
}

func (f fileGen) messageConstructorName(m *protogen.Message) string {
	return "New" + f.messageTableInterfaceName(m)
}

func (f fileGen) genStoreMethods(stores []*protogen.Message) {
	// getters
	for _, msg := range stores {
		name := f.messageTableInterfaceName(msg)
		f.P("func(x ", f.storeStructName(), ") ", name, "() ", name, "{")
		f.P("return x.", f.param(msg.GoIdent.GoName))
		f.P("}")
		f.P()
	}
	f.P("func(", f.storeStructName(), ") doNotImplement() {}")
	f.P()
}

func (f fileGen) genStoreInterfaceGuard() {
	f.P("var _ ", f.storeInterfaceName(), " = ", f.storeStructName(), "{}")
}

func (f fileGen) genStoreConstructor(stores []*protogen.Message) {
	f.P("func New", f.storeInterfaceName(), "(db ", ormTablePkg.Ident("Schema"), ") (", f.storeInterfaceName(), ", error) {")
	for _, store := range stores {
		f.P(f.messageTableReceiverName(store), ", err := ", f.messageConstructorName(store), "(db)")
		f.P("if err != nil {")
		f.P("return nil, err")
		f.P("}")
		f.P()
	}

	f.P("return ", f.storeStructName(), "{")
	for _, store := range stores {
		f.P(f.messageTableReceiverName(store), ",")
	}
	f.P("}, nil")
	f.P("}")
}

func (f fileGen) fieldsToCamelCase(fields string) string {
	splitFields := strings.Split(fields, ",")
	camelFields := make([]string, len(splitFields))
	for i, field := range splitFields {
		camelFields[i] = strcase.ToCamel(field)
	}
	return strings.Join(camelFields, "")
}
