package main

import (
	. "reflect"

	i "github.com/qlova/i"
)

type Package = struct {
	Binds    map[string]Value
	Types    map[string]Type
	Proxies  map[string]Type
	Untypeds map[string]string
	Wrappers map[string][]string
}

var Packages = make(map[string]Package)

// reflection: allow interpreted code to import "github.com/qlova/i"
func init() {
	Packages["github.com/qlova/i"] = Package{
		Binds: map[string]Value{
			"NewInteger": ValueOf(i.NewInteger),
			"IndexList":  ValueOf(i.IndexList),
			"IndexArray": ValueOf(i.IndexArray),
			"Atoi":       ValueOf(i.Atoi),
			"SetupTo":    ValueOf(i.SetupTo),
			"SetupStep":  ValueOf(i.SetupStep),
			"NewContext": ValueOf(i.NewContext),
		}, Types: map[string]Type{
			"Bit":     TypeOf((*i.Bit)(nil)).Elem(),
			"Integer": TypeOf((*i.Integer)(nil)).Elem(),
			"Context": TypeOf((*i.Context)(nil)).Elem(),
		},
	}
}
