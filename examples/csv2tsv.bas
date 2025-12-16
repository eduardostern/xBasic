#!/usr/local/bin/xbasic
' Convert CSV to TSV (comma to tab)
' Usage: cat data.csv | ./csv2tsv.bas > data.tsv

DO WHILE NOT EOF(0)
    LINE INPUT #0, line$
    ' Replace commas with tabs
    result$ = ""
    FOR i = 1 TO LEN(line$)
        c$ = MID$(line$, i, 1)
        IF c$ = "," THEN result$ = result$ + CHR$(9) ELSE result$ = result$ + c$
    NEXT i
    PRINT result$
LOOP
