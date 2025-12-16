' Function Examples - xBasic
CLS
PRINT "=== String Functions ==="
s$ = "Hello, World!"
PRINT "String:"; s$
PRINT "Length:"; LEN(s$)
PRINT "Upper:"; UCASE$(s$)
PRINT "Lower:"; LCASE$(s$)
PRINT "Left 5:"; LEFT$(s$, 5)
PRINT "Right 6:"; RIGHT$(s$, 6)
PRINT "Mid(8,5):"; MID$(s$, 8, 5)

PRINT
PRINT "=== Math Functions ==="
PRINT "ABS(-42) ="; ABS(-42)
PRINT "SQR(16) ="; SQR(16)
PRINT "INT(3.7) ="; INT(3.7)
PRINT "SIN(0) ="; SIN(0)
PRINT "COS(0) ="; COS(0)

PRINT
PRINT "=== Random Numbers ==="
RANDOMIZE TIMER
FOR i = 1 TO 5
    PRINT "RND ="; RND
NEXT i

PRINT
PRINT "=== Date/Time ==="
PRINT "Date:"; DATE$
PRINT "Time:"; TIME$
PRINT "Timer:"; TIMER

END
