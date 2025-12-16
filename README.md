# xBasic

A QBasic-compatible interpreter for macOS and Linux.

## Features

- **Core BASIC language support**: variables, arrays, control structures, subroutines
- **40+ built-in functions**: string manipulation, math, date/time
- **File I/O**: text and binary file operations
- **PRINT USING**: formatted numeric output
- **Graphics**: PSET, LINE, CIRCLE using Unicode block characters
- **Cross-platform**: works on macOS and Linux

## Quick Start

### Build

```bash
make build
```

Or build manually:

```bash
go build -o xbasic ./cmd/xbasic
```

### Run a BASIC program

```bash
./xbasic program.bas
```

## Language Features

### Data Types

- `INTEGER` (%) - 16-bit signed integer
- `LONG` (&) - 32-bit signed integer
- `SINGLE` (!) - 32-bit floating point
- `DOUBLE` (#) - 64-bit floating point
- `STRING` ($) - text string

### Control Structures

```basic
' IF/THEN/ELSE
IF x > 0 THEN
    PRINT "Positive"
ELSE
    PRINT "Non-positive"
END IF

' FOR/NEXT
FOR i = 1 TO 10 STEP 2
    PRINT i
NEXT i

' WHILE/WEND
WHILE x < 100
    x = x * 2
WEND

' DO/LOOP
DO
    PRINT n
    n = n + 1
LOOP UNTIL n > 10

' SELECT CASE
SELECT CASE grade
    CASE 90 TO 100
        PRINT "A"
    CASE 80 TO 89
        PRINT "B"
    CASE ELSE
        PRINT "Below B"
END SELECT
```

### Subroutines and Functions

```basic
' SUB definition
SUB PrintMessage (msg$)
    PRINT msg$
END SUB

' FUNCTION definition
FUNCTION Square (n)
    Square = n * n
END FUNCTION

' Calling
PrintMessage "Hello"
result = Square(5)
```

### Built-in Functions

**String Functions:**
- `LEN`, `LEFT$`, `RIGHT$`, `MID$`, `INSTR`
- `UCASE$`, `LCASE$`, `LTRIM$`, `RTRIM$`, `TRIM$`
- `STR$`, `VAL`, `CHR$`, `ASC`, `STRING$`, `SPACE$`

**Math Functions:**
- `ABS`, `SGN`, `INT`, `FIX`, `SQR`
- `SIN`, `COS`, `TAN`, `ATN`, `ATAN2`, `LOG`, `EXP`
- `RND`, `RANDOMIZE`, `ROUND`, `PI`

**Date/Time:**
- `DATE$`, `TIME$`, `TIMER`

**Conversion:**
- `CINT`, `CLNG`, `CSNG`, `CDBL`

**File I/O:**
- `EOF`, `LOF`, `LOC`, `FREEFILE`

### File I/O

```basic
' Text file output
OPEN "data.txt" FOR OUTPUT AS #1
PRINT #1, "Hello World"
CLOSE #1

' Text file input
OPEN "data.txt" FOR INPUT AS #1
DO WHILE NOT EOF(1)
    LINE INPUT #1, line$
    PRINT line$
LOOP
CLOSE #1

' Binary file I/O
OPEN "data.bin" FOR BINARY AS #1
PUT #1, 1, value%
GET #1, 1, result%
CLOSE #1
```

### PRINT USING

```basic
amount = 1234.567
PRINT USING "###.##"; amount        ' 1234.57
PRINT USING "$$###.##"; 123.45      ' $123.45
PRINT USING "**###.##"; 45.6        ' ***45.60
```

### REDIM

```basic
DIM arr(10)
REDIM arr(20)              ' Resize and clear
REDIM PRESERVE arr(30)     ' Resize and keep data
```

### Graphics (Terminal Unicode)

```basic
PSET (10, 5), 15           ' Plot point
LINE (0, 0)-(79, 24), 14   ' Draw line
CIRCLE (40, 12), 10, 12    ' Draw circle
```

## Example Program

```basic
' FizzBuzz in xBasic
CLS
PRINT "FizzBuzz 1-100"
PRINT

FOR i = 1 TO 100
    IF i MOD 15 = 0 THEN
        PRINT "FizzBuzz"
    ELSEIF i MOD 3 = 0 THEN
        PRINT "Fizz"
    ELSEIF i MOD 5 = 0 THEN
        PRINT "Buzz"
    ELSE
        PRINT i
    END IF
NEXT i

END
```

## Building for Different Platforms

```bash
# macOS (Intel)
make build-macos

# macOS (Apple Silicon)
make build-macos-arm64

# Linux (amd64)
make build-linux

# Linux (arm64)
make build-linux-arm64

# All platforms
make build-all
```

## Project Structure

```
xBasic/
├── cmd/xbasic/main.go      # Entry point
├── internal/
│   ├── lexer/              # Tokenizer
│   ├── parser/             # Parser (Pratt expression parsing)
│   ├── ast/                # Abstract Syntax Tree nodes
│   ├── interpreter/        # Tree-walking interpreter
│   ├── builtins/           # Built-in functions
│   └── screen/             # Screen/display handling
├── examples/               # Sample BASIC programs
├── Makefile
└── README.md
```

## License

MIT License
