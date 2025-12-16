' Loop Examples - xBasic
CLS
PRINT "=== FOR Loop ==="
FOR i = 1 TO 5
    PRINT "i ="; i
NEXT i

PRINT
PRINT "=== WHILE Loop ==="
x = 0
WHILE x < 3
    PRINT "x ="; x
    x = x + 1
WEND

PRINT
PRINT "=== DO...LOOP UNTIL ==="
n = 1
DO
    PRINT "n ="; n
    n = n + 1
LOOP UNTIL n > 3

PRINT
PRINT "Done!"
END
