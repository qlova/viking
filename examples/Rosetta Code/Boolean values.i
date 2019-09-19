//output: true\nfalse\ntrue\ntrue\nfalse\n
main
	//Logicals aka Booleans.
	b $= logical()

	b $= true
	print(b)

	b $= false
	print(b)
	
	//Non-zero values are true.
	b $= 1
	print(b)

	b $= -1
	print(b)

	//Zero values are false
	b $= 0
	print(b)
}