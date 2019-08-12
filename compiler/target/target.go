package target

import "bytes"

//Target is a compiler target.
type Target struct {
	string
}

func (target Target) Valid() bool {
	return target.string != ""
}

//Targets is a list of all possible targets.
var Targets = []Target{
	Target{"go"},
	Target{"rs"},
	Target{"java"},
	Target{"js"},
	Target{"cs"},
	Target{"py"},
	Target{"lua"},
}

//FromString converts a string to a valid target or empty.
func FromString(s string) (target Target) {
	switch s {
	case "go",
		"rs",
		"java",
		"js",
		"cs",
		"py",
		"lua":
		target = Target{s}
	}
	return
}

//Buffer because, each target has a buffer
type Buffer struct {
	Go, Rust, Java, JS, CSharp, Lua, Python Mode
}

//Get the target mode by string.
func (buffer *Buffer) Get(target Target) *Mode {
	switch target.string {
	case "go":
		return &buffer.Go
	case "rs":
		return &buffer.Rust
	case "java":
		return &buffer.Java
	case "js":
		return &buffer.JS
	case "cs":
		return &buffer.CSharp
	case "lua":
		return &buffer.Lua
	default:
		panic("invalid target")
	}
}

//Go target.
var Go = Buffer{Go: Mode{Enabled: true}}

type Mode struct {
	Enabled                bool
	Head, Neck, Body, Tail bytes.Buffer
}

func (target *Mode) Bytes() []byte {
	if target.Enabled {
		return target.Body.Bytes()
	}
	return nil
}

func (target *Mode) String() string {
	if target.Enabled {
		return target.Body.String()
	}
	return ""
}

func (target *Mode) Write(data []byte) (int, error) {
	if target.Enabled {
		return target.Body.Write(data)
	}
	return 0, nil
}

func (target *Mode) WriteString(s string) (int, error) {
	if target.Enabled {
		return target.Body.WriteString(s)
	}
	return 0, nil
}

func (target *Mode) WriteByte(b byte) error {
	if target.Enabled {
		return target.Body.WriteByte(b)
	}
	return nil
}
