' Math Functions Test Program
PRINT "Math Functions Test"
PRINT

' Test ATAN2
PRINT "ATAN2(1, 1) = "; ATAN2(1, 1)
PRINT "ATAN2(1, 0) = "; ATAN2(1, 0)
PRINT

' Test ROUND
PRINT "ROUND(3.14159) = "; ROUND(3.14159)
PRINT "ROUND(3.14159, 2) = "; ROUND(3.14159, 2)
PRINT "ROUND(3.14159, 4) = "; ROUND(3.14159, 4)
PRINT

' Test PI constant
PRINT "PI = "; PI
PRINT "SIN(PI/2) = "; SIN(PI/2)
PRINT

' Test PRINT USING
PRINT "PRINT USING Tests:"
amount = 1234.567
PRINT USING "###.##"; amount
PRINT USING "####.##"; 99.5
PRINT USING "$$###.##"; 123.45

END
