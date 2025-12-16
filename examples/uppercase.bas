#!/usr/local/bin/xbasic
' Convert stdin to uppercase
' Usage: echo "hello world" | ./uppercase.bas
'        cat file.txt | ./uppercase.bas

DO WHILE NOT EOF(0)
    LINE INPUT #0, line$
    PRINT UCASE$(line$)
LOOP
