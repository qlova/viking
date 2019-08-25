.variable(string(name))
	go `import "os"`; head

	return string.if
		go
			`os.Getenv(name)`
		rust
			`match env::var(name) {
				Ok(val) => val
				Err(e) => ""
			}`
	}
}
