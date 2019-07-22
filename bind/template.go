// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"reflect"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// tmplData is the data structure required to fill the binding template.
type tmplData struct {
	Package   string               // Name of the package to place the generated file in
	Contracts map[string]*contract // List of contracts to generate into this file
}

func parseData(name string, evmABI abi.ABI, pkg string) *tmplData {
	log.SetFlags(log.Llongfile)

	// Process each individual tmplContract requested binding
	contracts := make(map[string]*contract)
	contracts[name] = parseContract(name, evmABI)

	// Generate the tmplContract template data content and render it
	return &tmplData{
		Package:   pkg,
		Contracts: contracts,
	}
}

const templatePath = "./bind/templates/*"

func render(writer io.Writer, data *tmplData) error {
	funcs := template.FuncMap{
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"bindtype":      bindType,
		"bindtopictype": bindTopicType,
		"capitalise":    capitalise,
		"decapitalise":  decapitalise,
	}

	tmpl := template.Must(template.New("Bind").Funcs(funcs).ParseGlob(templatePath))
	if err := tmpl.ExecuteTemplate(writer, "Bind", data); err != nil {
		return err
	}
	return nil
}

func Render(data *tmplData) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := render(buffer, data); err != nil {
		return nil, err
	}

	// For Go bindings pass the code through gofmt to clean it up
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		return []byte{}, fmt.Errorf("%v\n%s", err, buffer)
	}
	return code, nil
}

func RenderFile(path string, data *tmplData) error {
	var out *os.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		out, err = os.Create(path)
		if err != nil {
			return err
		}
	} else {
		out, err = os.OpenFile(path, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if data, err := Render(data); err != nil {
		return err
	} else {
		if _, err = out.Write(data); err != nil {
			return err
		}
	}

	return nil
}
