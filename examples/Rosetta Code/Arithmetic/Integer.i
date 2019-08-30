//input: 3 5\n
//output: Sum: 8\nDifference: -2\nProduct: 15\nQuotient: 0\nModulus: 3\nExponent: 243\n
main
	a $= integer(in(' ')); ignore
	b $= integer(in('\n')); ignore
	
	print("Sum:"		, a + b)
	print("Difference:" , a - b)
	print("Product:"	, a * b)
	print("Quotient:"	, a / b) // rounds towards zero
	print("Modulus:"	, a % b) // same sign as first operand
	print("Exponent:"	, a ^ b)
}
