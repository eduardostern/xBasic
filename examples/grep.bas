#!/usr/local/bin/xbasic
' Simple grep: filter lines containing a pattern
' Usage: echo -e "hello\nfoo\nHELLO\nbar" | ./grep.bas
'        (reads pattern from first line, then filters remaining lines)

LINE INPUT #0, pattern$
DO WHILE NOT EOF(0)
    LINE INPUT #0, line$
    IF INSTR(line$, pattern$) > 0 THEN
        PRINT line$
    END IF
LOOP
