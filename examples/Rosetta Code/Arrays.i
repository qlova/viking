//output: 2\n2\n
main
	//Fixed-length arrays.
	f $= array[1].integer()
	f[0] $= 2
	print(f[0])

	//Dynamic arrays.
	d $= list.integer()
	d[+] $= 2
	print(d[1])
}