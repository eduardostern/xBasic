#!/usr/local/bin/xbasic
' Sum numbers from stdin (one per line)
' Usage: echo -e "10\n20\n30" | ./sum.bas
'        seq 1 100 | ./sum.bas

total = 0
DO WHILE NOT EOF(0)
    LINE INPUT #0, line$
    total = total + VAL(line$)
LOOP
PRINT total
