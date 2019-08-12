package target

import "bytes"

//Go target.
var Go = Buffer{Go: Target{Enabled: true}}

//Target is a compiler target.
type Target struct {
	Enabled                bool
	Head, Neck, Body, Tail bytes.Buffer
}

func (target *Target) Bytes() []byte {
	if target.Enabled {
		return target.Body.Bytes()
	}
	return nil
}

func (target *Target) String() string {
	if target.Enabled {
		return target.Body.String()
	}
	return ""
}

func (target *Target) Write(data []byte) (int, error) {
	if target.Enabled {
		return target.Body.Write(data)
	}
	return 0, nil
}

func (target *Target) WriteString(s string) (int, error) {
	if target.Enabled {
		return target.Body.WriteString(s)
	}
	return 0, nil
}

func (target *Target) WriteByte(b byte) error {
	if target.Enabled {
		return target.Body.WriteByte(b)
	}
	return nil
}

//Buffer because, each target has a buffer
type Buffer struct {
	Go Target
}
