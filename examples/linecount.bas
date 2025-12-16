#!/usr/local/bin/xbasic
' Count lines from stdin (like wc -l)
' Usage: cat file.txt | ./linecount.bas

count = 0
DO WHILE NOT EOF(0)
    LINE INPUT #0, line$
    count = count + 1
LOOP
PRINT count
