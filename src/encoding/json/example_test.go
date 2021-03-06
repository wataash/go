// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func Example_wataash_Unmarshal_non_array() {
	var jsonBlob = []byte(`{"Name": "Platypus", "Order": "Monotremata"}`)

	type Animal struct {
		Name  string
		Order string
	}
	var a Animal
	err := json.Unmarshal(jsonBlob, &a)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", a)
	// Output: {Name:Platypus Order:Monotremata}
}

func ExampleMarshal() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	b, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
	// Output:
	// {"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}
}

func ExampleMarshal_tag() {
	type JSONStruct struct {
		Field1      int   `json:"myName1"`
		Field2      int   `json:"myName2,omitempty"`
		Field3      int   `json:",omitempty"`
		Field4      int   `json:"-"`
		Field5      int   `json:"-,"`
		Int64String int64 `json:",string"`
	}
	for _, jsonStruct := range []JSONStruct{
		{
			Field1:      1,
			Field2:      2,
			Field3:      3,
			Field4:      4,
			Field5:      5,
			Int64String: 64,
		},
		{
			// empty struct
			// fields with "omitempty" will not be omitted
		},
	} {
		b, err := json.Marshal(jsonStruct)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println(string(b))
	}
	// Output:
	// {"myName1":1,"myName2":2,"Field3":3,"-":5,"Int64String":"64"}
	// {"myName1":0,"-":0,"Int64String":"0"}
}

func ExampleMarshal_export() {
	jsonStruct := struct {
		ExportedField   string
		unexportedField string
	}{
		ExportedField:   "I'll be marshaled",
		unexportedField: "I'll not be",
	}
	b, err := json.Marshal(jsonStruct)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
	// Output:
	// {"ExportedField":"I'll be marshaled"}
}

func ExampleUnmarshal() {
	var jsonBlob = []byte(`[
	{"Name": "Platypus", "Order": "Monotremata"},
	{"Name": "Quoll",    "Order": "Dasyuromorphia"}
]`)
	type Animal struct {
		Name  string
		Order string
	}
	var animals []Animal
	err := json.Unmarshal(jsonBlob, &animals)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animals)
	// Output:
	// [{Name:Platypus Order:Monotremata} {Name:Quoll Order:Dasyuromorphia}]
}

func ExampleUnmarshal_tag() {
	var jsonBlob = []byte(`{"Field1": 1, "myName2": 2, "Int64String": "64"}`)
	jsonStruct := struct {
		Field1      int
		Field2      int   `json:"myName2"`
		Int64String int64 `json:",string"`
	}{}
	err := json.Unmarshal(jsonBlob, &jsonStruct)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", jsonStruct)
	// Output:
	// {Field1:1 Field2:2 Int64String:64}
}

func ExampleUnmarshal_export() {
	var jsonBlob = []byte(`{
		"ExportedField": "I'll be unmarshaled",
		"unexportedField": "I'll not be"
	}`)
	jsonStruct := struct {
		ExportedField   string
		unexportedField string
	}{}
	err := json.Unmarshal(jsonBlob, &jsonStruct)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", jsonStruct)
	// Output:
	// {ExportedField:I'll be unmarshaled unexportedField:}
}

// This example uses a Decoder to decode a stream of distinct JSON values.
func ExampleDecoder() {
	const jsonStream = `
	{"Name": "Ed", "Text": "Knock knock."}
	{"Name": "Sam", "Text": "Who's there?"}
	{"Name": "Ed", "Text": "Go fmt."}
	{"Name": "Sam", "Text": "Go fmt who?"}
	{"Name": "Ed", "Text": "Go fmt yourself!"}
`
	type Message struct {
		Name, Text string
	}
	dec := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", m.Name, m.Text)
	}
	// Output:
	// Ed: Knock knock.
	// Sam: Who's there?
	// Ed: Go fmt.
	// Sam: Go fmt who?
	// Ed: Go fmt yourself!
}

// This example uses a Decoder to decode a stream of distinct JSON values.
func ExampleDecoder_Token() {
	const jsonStream = `
	{"Message": "Hello", "Array": [1, 2, 3], "Null": null, "Number": 1.234}
`
	dec := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%T: %v", t, t)
		if dec.More() {
			fmt.Printf(" (more)")
		}
		fmt.Printf("\n")
	}
	// Output:
	// json.Delim: { (more)
	// string: Message (more)
	// string: Hello (more)
	// string: Array (more)
	// json.Delim: [ (more)
	// float64: 1 (more)
	// float64: 2 (more)
	// float64: 3
	// json.Delim: ] (more)
	// string: Null (more)
	// <nil>: <nil> (more)
	// string: Number (more)
	// float64: 1.234
	// json.Delim: }
}

// This example uses a Decoder to decode a streaming array of JSON objects.
func ExampleDecoder_Decode_stream() {
	const jsonStream = `
	[
		{"Name": "Ed", "Text": "Knock knock."},
		{"Name": "Sam", "Text": "Who's there?"},
		{"Name": "Ed", "Text": "Go fmt."},
		{"Name": "Sam", "Text": "Go fmt who?"},
		{"Name": "Ed", "Text": "Go fmt yourself!"}
	]
`
	type Message struct {
		Name, Text string
	}
	dec := json.NewDecoder(strings.NewReader(jsonStream))

	// read open bracket
	t, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	// while the array contains values
	for dec.More() {
		var m Message
		// decode an array value (Message)
		err := dec.Decode(&m)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v: %v\n", m.Name, m.Text)
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	// Output:
	// json.Delim: [
	// Ed: Knock knock.
	// Sam: Who's there?
	// Ed: Go fmt.
	// Sam: Go fmt who?
	// Ed: Go fmt yourself!
	// json.Delim: ]
}

// This example uses RawMessage to delay parsing part of a JSON message.
func ExampleRawMessage_unmarshal() {
	type Color struct {
		Space string
		Point json.RawMessage // delay parsing until we know the color space
	}
	type RGB struct {
		R uint8
		G uint8
		B uint8
	}
	type YCbCr struct {
		Y  uint8
		Cb int8
		Cr int8
	}

	var j = []byte(`[
	{"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10}},
	{"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255}}
]`)
	var colors []Color
	err := json.Unmarshal(j, &colors)
	if err != nil {
		log.Fatalln("error:", err)
	}

	for _, c := range colors {
		var dst interface{}
		switch c.Space {
		case "RGB":
			dst = new(RGB)
		case "YCbCr":
			dst = new(YCbCr)
		}
		err := json.Unmarshal(c.Point, dst)
		if err != nil {
			log.Fatalln("error:", err)
		}
		fmt.Println(c.Space, dst)
	}
	// Output:
	// YCbCr &{255 0 -10}
	// RGB &{98 218 255}
}

// This example uses RawMessage to use a precomputed JSON during marshal.
func ExampleRawMessage_marshal() {
	h := json.RawMessage(`{"precomputed": true}`)

	c := struct {
		Header *json.RawMessage `json:"header"`
		Body   string           `json:"body"`
	}{Header: &h, Body: "Hello Gophers!"}

	b, err := json.MarshalIndent(&c, "", "\t")
	b, err = json.MarshalIndent(c, "", "\t") // same
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)

	// Output:
	// {
	// 	"header": {
	// 		"precomputed": true
	// 	},
	// 	"body": "Hello Gophers!"
	// }
}

func ExampleIndent() {
	type Road struct {
		Name   string
		Number int
	}
	roads := []Road{
		{"Diamond Fork", 29},
		{"Sheep Creek", 51},
	}

	b, err := json.Marshal(roads)
	if err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "=", "\t")
	out.WriteTo(os.Stdout)
	// Output:
	// [
	// =	{
	// =		"Name": "Diamond Fork",
	// =		"Number": 29
	// =	},
	// =	{
	// =		"Name": "Sheep Creek",
	// =		"Number": 51
	// =	}
	// =]
}

func ExampleMarshalIndent() {
	data := map[string]int{
		"a": 1,
		"b": 2,
	}

	json, err := json.MarshalIndent(data, "<prefix>", "<indent>")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(json))
	// Output:
	// {
	// <prefix><indent>"a": 1,
	// <prefix><indent>"b": 2
	// <prefix>}
}

func ExampleValid() {
	goodJSON := `{"example": 1}`
	badJSON := `{"example":2:]}}`

	fmt.Println(json.Valid([]byte(goodJSON)), json.Valid([]byte(badJSON)))
	// Output:
	// true false
}

func ExampleHTMLEscape() {
	var out bytes.Buffer
	json.HTMLEscape(&out, []byte(`{"Name":"<b>HTML content</b>"}`))
	out.WriteTo(os.Stdout)
	// Output:
	//{"Name":"\u003cb\u003eHTML content\u003c/b\u003e"}
}
