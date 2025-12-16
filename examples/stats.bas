#!/usr/local/bin/xbasic
' Calculate basic statistics from numbers (one per line)
' Usage: echo -e "10\n20\n30\n40\n50" | ./stats.bas
'        seq 1 100 | ./stats.bas

count = 0
total = 0
min# = 999999999
max# = -999999999

DO WHILE NOT EOF(0)
    LINE INPUT #0, line$
    n# = VAL(line$)
    count = count + 1
    total = total + n#
    IF n# < min# THEN min# = n#
    IF n# > max# THEN max# = n#
LOOP

IF count > 0 THEN
    PRINT "Count:"; count
    PRINT "Sum:  "; total
    PRINT "Min:  "; min#
    PRINT "Max:  "; max#
    PRINT "Avg:  "; total / count
ELSE
    PRINT "No data"
END IF
