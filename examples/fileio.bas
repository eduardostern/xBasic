' File I/O Test Program
PRINT "File I/O Test"
PRINT

' Test FREEFILE
PRINT "Next free file number: "; FREEFILE
PRINT

' Test text file output
PRINT "Writing to test.txt..."
OPEN "test.txt" FOR OUTPUT AS #1
PRINT #1, "Line 1: Hello World"
PRINT #1, "Line 2: Testing 123"
PRINT #1, "Line 3: xBasic File I/O"
CLOSE #1
PRINT "Done writing."
PRINT

' Test text file input with file size
PRINT "Reading from test.txt..."
OPEN "test.txt" FOR INPUT AS #1
PRINT "File size: "; LOF(1); " bytes"
PRINT
DO WHILE NOT EOF(1)
    LINE INPUT #1, line$
    PRINT "Read: "; line$
LOOP
CLOSE #1
PRINT "Done reading."
PRINT

PRINT "Test completed!"
END
